package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MaksimKorneplod/MyInvestor/internal/service"
)

type historyStub struct {
	ticker string
	period string
	page   service.HistoryPage
	err    error
}

func (s *historyStub) Candles(_ context.Context, ticker, period string) (service.HistoryPage, error) {
	s.ticker = ticker
	s.period = period
	if s.err != nil {
		return service.HistoryPage{}, s.err
	}
	return s.page, nil
}

func TestInstrumentHistory(t *testing.T) {
	begin := time.Date(2026, 3, 1, 9, 30, 0, 0, time.FixedZone("MSK", 3*60*60))
	stub := &historyStub{page: service.HistoryPage{
		Ticker: "SBER",
		Period: service.PeriodDay,
		Candles: []service.Candle{{
			Begin: begin,
			Open:  "300.5",
			High:  "301",
			Low:   "299",
			Close: "300.8",
		}},
	}}
	mux := newAPIMux(stub)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instruments/SBER/history?period=day", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	if stub.ticker != "SBER" || stub.period != "day" {
		t.Fatalf("called ticker=%s period=%s", stub.ticker, stub.period)
	}
	var page service.HistoryPage
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if page.Ticker != "SBER" || page.Period != "day" || len(page.Candles) != 1 {
		t.Fatalf("page = %+v", page)
	}
	candle := page.Candles[0]
	if candle.Open != "300.5" || candle.High != "301" || candle.Low != "299" || candle.Close != "300.8" {
		t.Fatalf("candle = %+v", candle)
	}
	if !candle.Begin.Equal(begin) {
		t.Fatalf("begin = %s", candle.Begin)
	}
}

func TestInstrumentHistoryEmpty(t *testing.T) {
	stub := &historyStub{page: service.HistoryPage{Ticker: "GAZP", Period: service.PeriodYear}}
	rec := httptest.NewRecorder()
	newAPIMux(stub).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/instruments/GAZP/history?period=year", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var page struct {
		Candles []service.Candle `json:"candles"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if page.Candles == nil {
		t.Fatal("empty candles encoded as null")
	}
}

func TestInstrumentHistoryErrors(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
		msg    string
	}{
		{name: "period", err: service.ErrInvalidPeriod, status: http.StatusBadRequest, msg: "unsupported period"},
		{name: "instrument", err: service.ErrInvalidInstrument, status: http.StatusBadRequest, msg: "invalid instrument"},
		{name: "moex", err: service.ErrSourceUnavailable, status: http.StatusServiceUnavailable, msg: "market data source is temporarily unavailable, please try again later"},
		{name: "other", err: errors.New("boom"), status: http.StatusInternalServerError, msg: "internal server error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			newAPIMux(&historyStub{err: tc.err}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/instruments/SBER/history?period=day", nil))
			if rec.Code != tc.status {
				t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
			}
			var payload struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatal(err)
			}
			if payload.Error != tc.msg {
				t.Fatalf("error = %q", payload.Error)
			}
		})
	}
}

func TestInstrumentHistorySharesServeMux(t *testing.T) {
	mux := newAPIMux(&historyStub{page: service.HistoryPage{
		Ticker:  "SBER",
		Period:  service.PeriodDay,
		Candles: []service.Candle{},
	}})

	health := httptest.NewRecorder()
	mux.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("health = %d", health.Code)
	}

	history := httptest.NewRecorder()
	mux.ServeHTTP(history, httptest.NewRequest(http.MethodGet, "/api/v1/instruments/SBER/history?period=day", nil))
	if history.Code != http.StatusOK {
		t.Fatalf("history = %d body = %s", history.Code, history.Body.String())
	}

	missing := httptest.NewRecorder()
	mux.ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "/api/v1/instruments/SBER", nil))
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing = %d", missing.Code)
	}
}

func TestInstrumentHistoryCanceled(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instruments/SBER/history?period=day", nil)
	ctx, cancel := context.WithCancel(req.Context())
	cancel()
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	InstrumentHistory(&historyStub{err: context.Canceled}, time.Minute).ServeHTTP(rec, req)
	if rec.Body.Len() != 0 {
		t.Fatalf("canceled request wrote a body: %s", rec.Body.String())
	}
}

func TestInstrumentHistoryDeadline(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instruments/SBER/history?period=day", nil)
	InstrumentHistory(blockingHistory{}, 20*time.Millisecond).ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
}

type blockingHistory struct{}

func (blockingHistory) Candles(ctx context.Context, _, _ string) (service.HistoryPage, error) {
	<-ctx.Done()
	return service.HistoryPage{}, ctx.Err()
}

func newAPIMux(history instrumentHistory) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", HealthHandler)
	MountInstrumentHistory(mux, history, time.Minute)
	return mux
}
