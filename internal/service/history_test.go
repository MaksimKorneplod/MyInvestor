package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MaksimKorneplod/MyInvestor/internal/repository/moex"
)

type stubCandles struct {
	query moex.HistoryQuery
	rows  []moex.ChartCandle
	err   error
}

func (s *stubCandles) Chart(_ context.Context, q moex.HistoryQuery) ([]moex.ChartCandle, error) {
	s.query = q
	return s.rows, s.err
}

func TestHistoryPeriods(t *testing.T) {
	now := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		period   string
		interval int
	}{
		{PeriodDay, 1},
		{PeriodWeek, 60},
		{PeriodMonth, 60},
		{PeriodHalfYear, 24},
		{PeriodYear, 7},
		{PeriodFiveYears, 7},
	}
	for _, tc := range cases {
		t.Run(tc.period, func(t *testing.T) {
			source := &stubCandles{}
			history := NewHistory(source)
			history.now = func() time.Time { return now }
			page, err := history.Candles(context.Background(), "SBER", tc.period)
			if err != nil {
				t.Fatal(err)
			}
			if page.Ticker != "SBER" || page.Period != tc.period {
				t.Fatalf("page = %+v", page)
			}
			if page.Candles == nil {
				t.Fatal("candles must be an empty slice, not nil")
			}
			if source.query.Interval != tc.interval {
				t.Fatalf("interval = %d", source.query.Interval)
			}
			if source.query.Instrument.Security != "SBER" || source.query.Instrument.Engine != "stock" || source.query.Instrument.Market != "shares" || source.query.Instrument.Board != "" {
				t.Fatalf("instrument = %+v", source.query.Instrument)
			}
		})
	}
}

func TestHistoryEmptyCandles(t *testing.T) {
	source := &stubCandles{}
	history := NewHistory(source)
	page, err := history.Candles(context.Background(), "SBER", PeriodDay)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Candles) != 0 {
		t.Fatalf("candles = %+v", page.Candles)
	}
}

func TestHistoryKeepsOHLC(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("start") != "0" {
			_, _ = io.WriteString(w, `{"candles":{"columns":["begin","open","high","low","close"],"data":[]}}`)
			return
		}
		_, _ = io.WriteString(w, `{"candles":{"columns":["begin","open","high","low","close"],"data":[["2026-03-01 10:00:00",10,12,null,11]]}}`)
	}))
	t.Cleanup(server.Close)

	history := NewHistory(moex.New(server.URL))
	page, err := history.Candles(context.Background(), "SBER", PeriodWeek)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Candles) != 1 {
		t.Fatalf("candles = %+v", page.Candles)
	}
	candle := page.Candles[0]
	if candle.Open != "10" || candle.High != "12" || candle.Low != "" || candle.Close != "11" {
		t.Fatalf("candle = %+v", candle)
	}
	if candle.Begin.Format("2006-01-02 15:04:05") != "2026-03-01 10:00:00" {
		t.Fatalf("begin = %s", candle.Begin)
	}
}

func TestHistorySourceError(t *testing.T) {
	source := &stubCandles{err: errors.New("moex down")}
	history := NewHistory(source)
	_, err := history.Candles(context.Background(), "SBER", PeriodMonth)
	if !errors.Is(err, ErrSourceUnavailable) {
		t.Fatalf("err = %v", err)
	}
}

func TestHistoryValidation(t *testing.T) {
	source := &stubCandles{}
	history := NewHistory(source)
	cases := []struct {
		name   string
		ticker string
		period string
		want   error
		noCall bool
	}{
		{name: "period", ticker: "SBER", period: "decade", want: ErrInvalidPeriod, noCall: true},
		{name: "empty period", ticker: "SBER", period: "", want: ErrInvalidPeriod, noCall: true},
		{name: "empty ticker", ticker: "", period: PeriodDay, want: ErrInvalidInstrument, noCall: true},
		{name: "slash", ticker: "SBER/P", period: PeriodYear, want: ErrInvalidInstrument, noCall: true},
		{name: "space", ticker: "bad ticker", period: PeriodHalfYear, want: ErrInvalidInstrument, noCall: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source.query = moex.HistoryQuery{}
			_, err := history.Candles(context.Background(), tc.ticker, tc.period)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v", err)
			}
			if tc.noCall && source.query.Instrument.Security != "" {
				t.Fatalf("source was called: %+v", source.query)
			}
		})
	}
}

func TestHistoryCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	source := &stubCandles{err: context.Canceled}
	history := NewHistory(source)
	_, err := history.Candles(ctx, "SBER", PeriodDay)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
}
