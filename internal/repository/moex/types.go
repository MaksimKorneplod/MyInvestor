package moex

import (
	"fmt"
	"strconv"
	"time"
)

// Instrument — путь ISS, а не произвольный URL: движок, рынок, режим, бумага.
// Пустой Board повторяет iOS-адрес без режима: /engines/stock/markets/shares/...
type Instrument struct {
	Engine   string
	Market   string
	Board    string
	Security string
}

// SecuritiesQuery — список инструментов рынка или режима.
type SecuritiesQuery struct {
	Instrument Instrument
}

// Security — бумага из блока securities: SECID и SHORTNAME.
type Security struct {
	SecID     string
	ShortName string
}

// Quote — последняя котировка из блока marketdata.
// Last всегда заполнен: строка без LAST в список не попадает, как в iOS.
// LastChangePercent == nil, если в ячейке null.
type Quote struct {
	SecID             string
	Last              *Decimal
	LastChangePercent *Decimal
}

// HistoryQuery — свечи бумаги. Исторический ряд графика в MOEXService
// берётся из candles.json, не из дневного /iss/history.
type HistoryQuery struct {
	Instrument Instrument
	From       time.Time
	Till       time.Time
	Interval   int
}

// HistoryRow — одна свеча. ChartView рисует отрезок от Open к Close.
type HistoryRow struct {
	Begin time.Time
	Open  Decimal
	Close Decimal
}

// Decimal хранит десятичную запись числа так, как её прислал ISS.
type Decimal struct {
	text string
}

// String возвращает исходную запись числа.
func (d Decimal) String() string { return d.text }

// Float64 нужен графику. Для хранения котировки используется String.
func (d Decimal) Float64() (float64, error) {
	if d.text == "" {
		return 0, fmt.Errorf("%w: empty decimal", ErrParse)
	}
	return strconv.ParseFloat(d.text, 64)
}

// ChartPeriod повторяет enum графика iOS.
type ChartPeriod string

const (
	PeriodDay       ChartPeriod = "day"
	PeriodWeek      ChartPeriod = "week"
	PeriodMonth     ChartPeriod = "month"
	PeriodHalfYear  ChartPeriod = "halfYear"
	PeriodYear      ChartPeriod = "year"
	PeriodFiveYears ChartPeriod = "fiveYears"
)

type periodSpec struct {
	interval int
	daysBack int
}

// Коды interval — коды ISS из Candle.moexIntervalForPeriod.
var periodSpecs = map[ChartPeriod]periodSpec{
	PeriodDay:       {interval: 1, daysBack: 2},
	PeriodWeek:      {interval: 60, daysBack: 14},
	PeriodMonth:     {interval: 60, daysBack: 35},
	PeriodHalfYear:  {interval: 24, daysBack: 190},
	PeriodYear:      {interval: 7, daysBack: 370},
	PeriodFiveYears: {interval: 7, daysBack: 1900},
}

// HistoryQuery собирает запрос свечей так же, как Candle.fetchCandles:
// рынок акций, без режима, окно и interval от периода.
func (p ChartPeriod) HistoryQuery(security string, now time.Time) (HistoryQuery, error) {
	spec, ok := periodSpecs[p]
	if !ok || now.IsZero() {
		return HistoryQuery{}, fmt.Errorf("%w: period %q", ErrRequest, p)
	}
	if err := ident("security", security); err != nil {
		return HistoryQuery{}, err
	}
	today := now.In(moscow)
	return HistoryQuery{
		Instrument: Instrument{Engine: "stock", Market: "shares", Security: security},
		From:       today.AddDate(0, 0, -spec.daysBack),
		Till:       today,
		Interval:   spec.interval,
	}, nil
}

var moscow = loadMoscow()

func loadMoscow() *time.Location {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return time.FixedZone("MSK", 3*60*60)
	}
	return loc
}
