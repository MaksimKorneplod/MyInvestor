package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"unicode/utf8"

	"github.com/MaksimKorneplod/MyInvestor/internal/service"
)

// InstrumentSearcher is the minimal interface needed by the handler.
type InstrumentSearcher interface {
	Search(ctx context.Context, query string, limit, offset int) (service.Page, error)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// InstrumentsHandler provides the /api/v1/instruments HTTP endpoint.
func InstrumentsHandler(searcher InstrumentSearcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		query := r.URL.Query().Get("q")
		if utf8.RuneCountInString(query) > 100 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "parameter q too long (max 100 runes)"})
			return
		}

		limitStr := r.URL.Query().Get("limit")
		limit := 50
		if limitStr != "" {
			n, err := strconv.Atoi(limitStr)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid limit parameter"})
				return
			}
			limit = n
		}

		offsetStr := r.URL.Query().Get("offset")
		offset := 0
		if offsetStr != "" {
			n, err := strconv.Atoi(offsetStr)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid offset parameter"})
				return
			}
			offset = n
		}

		if limit < 1 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid limit parameter (must be >= 1)"})
			return
		}
		if limit > 100 {
			limit = 100
		}
		if offset < 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid offset parameter (must be >= 0)"})
			return
		}

		if searcher == nil {
			slog.Error("instrument searcher is nil")
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		page, err := searcher.Search(r.Context(), query, limit, offset)
		if err != nil {
			if errors.Is(err, service.ErrSourceUnavailable) {
				slog.Error("market data source unavailable", "error", err)
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "market data source is temporarily unavailable, please try again later"})
				return
			}
			if errors.Is(err, context.Canceled) {
				return
			}
			if errors.Is(err, context.DeadlineExceeded) {
				writeJSON(w, http.StatusGatewayTimeout, map[string]string{"error": "request timed out, please try again later"})
				return
			}
			slog.Error("internal server error", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}
		if page.Items == nil {
			page.Items = []service.Instrument{}
		}

		response := struct {
			Items  []service.Instrument `json:"items"`
			Total  int                  `json:"total"`
			Limit  int                  `json:"limit"`
			Offset int                  `json:"offset"`
		}{
			Items:  page.Items,
			Total:  page.Total,
			Limit:  limit,
			Offset: offset,
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}
