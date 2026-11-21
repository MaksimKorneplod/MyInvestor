package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/MaksimKorneplod/MyInvestor/internal/service"
)

// instrumentHistory — сервис свечей, который нужен обработчику.
type instrumentHistory interface {
	Candles(ctx context.Context, ticker, period string) (service.HistoryPage, error)
}

// InstrumentHistoryPattern — маршрут свечей на общем ServeMux приложения.
const InstrumentHistoryPattern = "GET /api/v1/instruments/{ticker}/history"

// MountInstrumentHistory регистрирует свечи на переданном mux.
func MountInstrumentHistory(mux *http.ServeMux, history instrumentHistory) {
	mux.HandleFunc(InstrumentHistoryPattern, InstrumentHistory(history))
}

// InstrumentHistory отдаёт свечи для графика на странице инструмента.
// Инструмент — тикер каталога, период — day, week, month, halfYear, year или fiveYears.
func InstrumentHistory(history instrumentHistory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if history == nil {
			slog.Error("instrument history is nil")
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}
		page, err := history.Candles(r.Context(), r.PathValue("ticker"), r.URL.Query().Get("period"))
		switch {
		case errors.Is(err, service.ErrInvalidPeriod):
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported period"})
		case errors.Is(err, service.ErrInvalidInstrument):
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid instrument"})
		case errors.Is(err, service.ErrSourceUnavailable):
			slog.Error("market data source unavailable", "error", err)
			respondJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "market data source is temporarily unavailable, please try again later"})
		case err != nil:
			slog.Error("instrument history failed", "error", err)
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		default:
			if page.Candles == nil {
				page.Candles = []service.Candle{}
			}
			respondJSON(w, http.StatusOK, page)
		}
	}
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
