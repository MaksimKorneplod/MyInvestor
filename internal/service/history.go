package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MaksimKorneplod/MyInvestor/internal/repository/moex"
)

// Периоды графика совпадают с ChartPeriod клиента MOEX и с экраном инструмента.
const (
	PeriodDay       = string(moex.PeriodDay)
	PeriodWeek      = string(moex.PeriodWeek)
	PeriodMonth     = string(moex.PeriodMonth)
	PeriodHalfYear  = string(moex.PeriodHalfYear)
	PeriodYear      = string(moex.PeriodYear)
	PeriodFiveYears = string(moex.PeriodFiveYears)
)

var (
	// ErrInvalidPeriod — период не входит в список, который рисует клиент.
	ErrInvalidPeriod = errors.New("unsupported period")
	// ErrInvalidInstrument — тикер пустой, неверного формата или отсутствует в каталоге.
	ErrInvalidInstrument = errors.New("invalid instrument")
	// ErrSourceUnavailable — MOEX не ответил. Тот же смысл, что у каталога инструментов.
	ErrSourceUnavailable = errors.New("market data source is temporarily unavailable")
)

// Candle — одна точка графика. Цены — десятичные строки, как их прислал ISS.
// High и Low пустые, если у источника в ячейке был null.
type Candle struct {
	Begin time.Time `json:"begin"`
	Open  string    `json:"open"`
	High  string    `json:"high,omitempty"`
	Low   string    `json:"low,omitempty"`
	Close string    `json:"close"`
}

// HistoryPage — ответ API истории для одного инструмента и периода.
type HistoryPage struct {
	Ticker  string   `json:"ticker"`
	Period  string   `json:"period"`
	Candles []Candle `json:"candles"`
}

// candleSource — клиент свечей и каталога. Его закрывает *moex.Client.
type candleSource interface {
	Chart(ctx context.Context, q moex.HistoryQuery) ([]moex.ChartCandle, error)
	Listed(ctx context.Context, ticker string) (bool, error)
}

// History собирает свечи графика из MOEX.
type History struct {
	source candleSource
	now    func() time.Time
}

// NewHistory создаёт сервис поверх клиента ISS.
func NewHistory(source candleSource) History {
	return History{source: source, now: time.Now}
}

// Candles возвращает свечи тикера за период.
// Тикер — тот же SECID, который каталог инструментов отдаёт в ticker и id.
func (h History) Candles(ctx context.Context, ticker, period string) (HistoryPage, error) {
	chartPeriod, ok := chartPeriods[period]
	if !ok {
		return HistoryPage{}, ErrInvalidPeriod
	}
	if h.source == nil {
		return HistoryPage{}, ErrSourceUnavailable
	}
	now := time.Now()
	if h.now != nil {
		now = h.now()
	}
	query, err := chartPeriod.HistoryQuery(ticker, now)
	if err != nil {
		return HistoryPage{}, ErrInvalidInstrument
	}
	found, err := h.source.Listed(ctx, ticker)
	if err != nil {
		if ctx != nil && ctx.Err() != nil {
			return HistoryPage{}, ctx.Err()
		}
		return HistoryPage{}, fmt.Errorf("%w: %w", ErrSourceUnavailable, err)
	}
	if !found {
		return HistoryPage{}, ErrInvalidInstrument
	}
	rows, err := h.source.Chart(ctx, query)
	if err != nil {
		if ctx != nil && ctx.Err() != nil {
			return HistoryPage{}, ctx.Err()
		}
		return HistoryPage{}, fmt.Errorf("%w: %w", ErrSourceUnavailable, err)
	}
	candles := make([]Candle, 0, len(rows))
	for _, row := range rows {
		candles = append(candles, candleFrom(row))
	}
	return HistoryPage{Ticker: ticker, Period: period, Candles: candles}, nil
}

func candleFrom(row moex.ChartCandle) Candle {
	candle := Candle{
		Begin: row.Begin,
		Open:  row.Open.String(),
		Close: row.Close.String(),
	}
	if row.High != nil {
		candle.High = row.High.String()
	}
	if row.Low != nil {
		candle.Low = row.Low.String()
	}
	return candle
}

var chartPeriods = map[string]moex.ChartPeriod{
	PeriodDay:       moex.PeriodDay,
	PeriodWeek:      moex.PeriodWeek,
	PeriodMonth:     moex.PeriodMonth,
	PeriodHalfYear:  moex.PeriodHalfYear,
	PeriodYear:      moex.PeriodYear,
	PeriodFiveYears: moex.PeriodFiveYears,
}
