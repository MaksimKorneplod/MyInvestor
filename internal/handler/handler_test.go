package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MaksimKorneplod/MyInvestor/internal/service/quotes"
)

// --- Мок сервиса ---

type mockService struct {
	marketQuote []quotes.Quote
	marketErr   error

	batchQuote []quotes.Quote
	batchErr   error

	oneQuote *quotes.Quote
	oneErr   error
}

func (m *mockService) One(ctx context.Context, ref quotes.Instrument) (*quotes.Quote, error) {
	return m.oneQuote, m.oneErr
}

func (m *mockService) Market(ctx context.Context, ref quotes.Instrument) ([]quotes.Quote, error) {
	return m.marketQuote, m.marketErr
}

func (m *mockService) Batch(ctx context.Context, ref quotes.Instrument, tickers []string) ([]quotes.Quote, error) {
	return m.batchQuote, m.batchErr
}

// Хелпер для быстрой генерации ответа с ошибкой
func assertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedCode int, expectedReason string) {
	t.Helper()
	if w.Code != expectedCode {
		t.Fatalf("expected status %d, got %d", expectedCode, w.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if body["error"] != expectedReason {
		t.Fatalf("expected reason %q, got %q", expectedReason, body["error"])
	}
}

// --- Тесты маршрутизации и регистрации ---

func TestRegisterRoutes(t *testing.T) {
	mux := http.NewServeMux()
	h := New(&mockService{})
	h.Register(mux)

	// Проверяем, что маршруты действительно зарегистрированы
	// (Go 1.22+ вернет 405 Method Not Allowed, если путь есть, но метод другой)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/quotes", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed for POST /quotes, got %d", w.Code)
	}
}

// --- Тесты валидации (400 Bad Request) ---

func TestHandleMarket_Validation(t *testing.T) {
	h := New(&mockService{})

	tests := []struct {
		name       string
		query      string
		expectCode int
		expectErr  string
	}{
		{"Bad Engine", "?engine=forex&market=shares", http.StatusBadRequest, "bad_market"},
		{"Bad Market", "?engine=stock&market=options", http.StatusBadRequest, "bad_market"},
		{"Bad Board", "?engine=stock&market=shares&board=UNKNOWN", http.StatusBadRequest, "bad_market"},
		{"Valid Without Board", "?engine=stock&market=shares", http.StatusOK, ""},
		{"Valid With Board", "?engine=stock&market=shares&board=TQBR", http.StatusOK, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/quotes"+tt.query, nil)
			w := httptest.NewRecorder()

			h.handleMarket(w, req)

			if tt.expectCode != http.StatusOK {
				assertErrorResponse(t, w, tt.expectCode, tt.expectErr)
			} else if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", w.Code)
			}
		})
	}
}

func TestHandleOne_Validation(t *testing.T) {
	h := New(&mockService{})

	// Проверка регулярки тикера
	req := httptest.NewRequest(http.MethodGet, "/api/v1/quotes/stock/shares/INVALID_TICKER!", nil)
	// Так как мы тестируем хендлер напрямую (не через mux), нужно вручную задать PathValue
	req.SetPathValue("engine", "stock")
	req.SetPathValue("market", "shares")
	req.SetPathValue("security", "INVALID_TICKER!")

	w := httptest.NewRecorder()
	h.handleOne(w, req)

	assertErrorResponse(t, w, http.StatusBadRequest, "bad_security")
}

func TestHandleBatch_Validation(t *testing.T) {
	h := New(&mockService{})

	tests := []struct {
		name       string
		body       string
		expectCode int
		expectErr  string
	}{
		{"Invalid JSON", `{bad_json}`, http.StatusBadRequest, "bad_body"},
		{"Unknown Fields", `{"engine":"stock","market":"shares","tickers":["SBER"],"extra":1}`, http.StatusBadRequest, "bad_body"},
		{"Trailing Data", `{"engine":"stock","market":"shares","tickers":["SBER"]} extra`, http.StatusBadRequest, "bad_body"},
		{"Empty Tickers", `{"engine":"stock","market":"shares","tickers":[]}`, http.StatusBadRequest, "bad_tickers"},
		{"Bad Engine", `{"engine":"bad","market":"shares","tickers":["SBER"]}`, http.StatusBadRequest, "bad_market"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/quotes:batch", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.handleBatch(w, req)

			assertErrorResponse(t, w, tt.expectCode, tt.expectErr)
		})
	}
}

// --- Тесты успешных ответов и DTO ---

func TestHandleOne_Success(t *testing.T) {
	// Создаем фейковую бумагу
	// Внимание: если moex.Decimal - сложный тип, используйте его конструктор.
	// Здесь я делаю допущение, что он поддерживает конвертацию.
	q := quotes.Quote{
		SecID: "SBER",
	}

	h := New(&mockService{oneQuote: &q})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/quotes/stock/shares/SBER", nil)
	req.SetPathValue("engine", "stock")
	req.SetPathValue("market", "shares")
	req.SetPathValue("security", "SBER")
	w := httptest.NewRecorder()

	h.handleOne(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var dto QuoteDTO
	json.Unmarshal(w.Body.Bytes(), &dto)
	if dto.Ticker != "SBER" {
		t.Fatalf("expected ticker SBER, got %s", dto.Ticker)
	}
}

func TestHandleOne_NotFound(t *testing.T) {
	// Сервис вернул nil (бумага не найдена на бирже), без ошибки
	h := New(&mockService{oneQuote: nil, oneErr: nil})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/quotes/stock/shares/XXXX", nil)
	req.SetPathValue("engine", "stock")
	req.SetPathValue("market", "shares")
	req.SetPathValue("security", "XXXX")
	w := httptest.NewRecorder()

	h.handleOne(w, req)

	assertErrorResponse(t, w, http.StatusNotFound, "not_found")
}

// --- Тесты маппинга ошибок (500, 502, 503, 504) ---

func TestWriteMapped_ErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		expectCode int
		expectErr  string
	}{
		{"Bad Request", quotes.ErrInvalidRequest, http.StatusBadRequest, "bad_request"},
		{"Canceled", context.Canceled, http.StatusServiceUnavailable, "canceled"},
		{"Timeout", quotes.ErrTimeout, http.StatusGatewayTimeout, "upstream_timeout"},
		{"Parse Error", quotes.ErrBadUpstream, http.StatusBadGateway, "upstream_parse"},
		{"Network Error", quotes.ErrUnavailable, http.StatusServiceUnavailable, "upstream_unavailable"},
		{"Status 429", quotes.ErrRateLimited, http.StatusServiceUnavailable, "rate_limited"},
		{"Status 404", quotes.ErrNotFound, http.StatusNotFound, "not_found"},
		{"Status 500", &quotes.UpstreamStatusError{Code: http.StatusInternalServerError}, http.StatusBadGateway, "upstream_status"},
		{"Unknown", errors.New("some random error"), http.StatusInternalServerError, "internal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			writeMapped(w, tt.err)
			assertErrorResponse(t, w, tt.expectCode, tt.expectErr)
		})
	}
}
