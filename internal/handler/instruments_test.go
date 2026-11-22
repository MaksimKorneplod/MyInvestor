package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MaksimKorneplod/MyInvestor/internal/service"
)

// mockSearcher implements InstrumentSearcher.
type mockSearcher struct {
	page service.Page
	err  error
}

func (m *mockSearcher) Search(ctx context.Context, query string, limit, offset int) (service.Page, error) {
	return m.page, m.err
}

func captureHandlerLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	previous := slog.Default()
	var logs bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() {
		slog.SetDefault(previous)
	})
	return &logs
}

func TestInstrumentsHandlerMethodNotAllowed(t *testing.T) {
	handler := InstrumentsHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/instruments", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 Method Not Allowed, got %d", w.Code)
	}
}

func TestInstrumentsHandlerDefaultLimitAndOffset(t *testing.T) {
	searcher := &mockSearcher{page: service.Page{Items: []service.Instrument{{ID: "TEST-ID-1", Ticker: "SBER", Name: "Sberbank"}}, Total: 4}}
	handler := InstrumentsHandler(searcher)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instruments", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var res struct {
		Items  []service.Instrument `json:"items"`
		Total  int                  `json:"total"`
		Limit  int                  `json:"limit"`
		Offset int                  `json:"offset"`
	}
	json.Unmarshal(w.Body.Bytes(), &res)

	if len(res.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(res.Items))
	}
	if res.Limit != 50 || res.Offset != 0 {
		t.Errorf("default limit/offset wrong: %+v", res)
	}
}

func TestInstrumentsHandlerLimitCappedAt100(t *testing.T) {
	searcher := &mockSearcher{page: service.Page{Items: []service.Instrument{}, Total: 3}}
	handler := InstrumentsHandler(searcher)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instruments?limit=150", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var res struct {
		Items  []service.Instrument `json:"items"`
		Total  int                  `json:"total"`
		Limit  int                  `json:"limit"`
		Offset int                  `json:"offset"`
	}
	json.Unmarshal(w.Body.Bytes(), &res)

	if res.Limit != 100 {
		t.Errorf("limit should be capped at 100, got %d", res.Limit)
	}
}

func TestInstrumentsHandlerInvalidLimitNonNumeric(t *testing.T) {
	handler := InstrumentsHandler(&mockSearcher{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instruments?limit=abc", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestInstrumentsHandlerInvalidLimitZero(t *testing.T) {
	handler := InstrumentsHandler(&mockSearcher{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instruments?limit=0", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestInstrumentsHandlerInvalidLimitNegative(t *testing.T) {
	handler := InstrumentsHandler(&mockSearcher{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instruments?limit=-5", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestInstrumentsHandlerInvalidOffsetNegative(t *testing.T) {
	handler := InstrumentsHandler(&mockSearcher{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instruments?offset=-1", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestInstrumentsHandlerInvalidOffsetNonNumeric(t *testing.T) {
	handler := InstrumentsHandler(&mockSearcher{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instruments?offset=abc", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestInstrumentsHandlerQLongRuneCount(t *testing.T) {
	longQuery := strings.Repeat("я", 101)
	handler := InstrumentsHandler(&mockSearcher{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instruments?q="+longQuery, nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for long q by rune count, got %d", w.Code)
	}
}

func TestInstrumentsHandlerEmptyResult(t *testing.T) {
	searcher := &mockSearcher{page: service.Page{Items: []service.Instrument{}, Total: 0}}
	handler := InstrumentsHandler(searcher)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instruments", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var res struct {
		Items  []service.Instrument `json:"items"`
		Total  int                  `json:"total"`
		Limit  int                  `json:"limit"`
		Offset int                  `json:"offset"`
	}
	json.Unmarshal(w.Body.Bytes(), &res)

	if res.Items == nil {
		t.Error("items should be empty slice, not nil")
	}
	if res.Total != 0 {
		t.Errorf("expected total 0, got %d", res.Total)
	}
}

func TestInstrumentsHandlerServiceUnavailable(t *testing.T) {
	searcher := &mockSearcher{err: service.ErrSourceUnavailable}
	handler := InstrumentsHandler(searcher)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instruments", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}

	var errResp struct {
		Error string `json:"error"`
	}
	json.Unmarshal(w.Body.Bytes(), &errResp)
	if errResp.Error == "" {
		t.Error("error message should not be empty")
	}
	if strings.Contains(errResp.Error, "down") || strings.Contains(errResp.Error, "internal") {
		t.Error("internal error details should not leak")
	}
}

func TestInstrumentsHandlerSourceUnavailableWinsOverDeadline(t *testing.T) {
	searcher := &mockSearcher{err: fmt.Errorf("%w: %w", service.ErrSourceUnavailable, context.DeadlineExceeded)}
	handler := InstrumentsHandler(searcher)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instruments", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
	var errResp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if errResp.Error != "market data source is temporarily unavailable, please try again later" {
		t.Errorf("error = %q, want generic source unavailable message", errResp.Error)
	}
}

func TestInstrumentsHandlerCanceledDoesNotRespondOrLog(t *testing.T) {
	logs := captureHandlerLogs(t)
	searcher := &mockSearcher{err: context.Canceled}
	handler := InstrumentsHandler(searcher)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instruments", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Body.Len() != 0 {
		t.Errorf("response body = %q, want empty", w.Body.String())
	}
	if logs.Len() != 0 {
		t.Errorf("logs = %q, want no log entries", logs.String())
	}
}

func TestInstrumentsHandlerDeadlineExceededReturns504WithoutLog(t *testing.T) {
	logs := captureHandlerLogs(t)
	searcher := &mockSearcher{err: context.DeadlineExceeded}
	handler := InstrumentsHandler(searcher)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instruments", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d", w.Code)
	}
	var errResp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if errResp.Error != "request timed out, please try again later" {
		t.Errorf("error = %q, want generic timeout message", errResp.Error)
	}
	if logs.Len() != 0 {
		t.Errorf("logs = %q, want no log entries", logs.String())
	}
}

func TestInstrumentsHandlerInternalError(t *testing.T) {
	searcher := &mockSearcher{err: errors.New("unexpected")}
	handler := InstrumentsHandler(searcher)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instruments", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	var errResp struct {
		Error string `json:"error"`
	}
	json.Unmarshal(w.Body.Bytes(), &errResp)
	if errResp.Error == "" {
		t.Error("error message should not be empty")
	}
	if strings.Contains(errResp.Error, "unexpected") {
		t.Error("internal error details should not leak")
	}
}

func TestInstrumentsHandlerContentType(t *testing.T) {
	handler := InstrumentsHandler(&mockSearcher{page: service.Page{Items: []service.Instrument{{ID: "TEST-ID-1", Ticker: "SBER", Name: "Sberbank"}}, Total: 1}})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instruments", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	ct := w.Header().Get("Content-Type")
	if ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json; charset=utf-8")
	}
}
