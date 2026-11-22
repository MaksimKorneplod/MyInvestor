package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/MaksimKorneplod/MyInvestor/internal/service/quotes"
)

// writeMapped превращает ошибку service в HTTP-ответ.
//
// Классы ошибок сервиса котировок:
//
//	ErrInvalidRequest    → 400
//	context.Canceled     → 503
//	ErrTimeout           → 504
//	ErrBadUpstream       → 502
//	ErrUnavailable       → 503
//	ErrRateLimited       → 503
//	ErrNotFound          → 404
//	UpstreamStatusError  → 502
//	неизвестное          → 500
//
// Наружу отдаём только стабильный code и короткое сообщение.
// Текст ошибки ISS, URL, детали парсинга — только в логи, не в JSON.
func writeMapped(w http.ResponseWriter, err error) {
	var statusErr *quotes.UpstreamStatusError
	switch {
	case errors.Is(err, quotes.ErrInvalidRequest):
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
	case errors.Is(err, context.Canceled):
		writeError(w, http.StatusServiceUnavailable, "canceled", "request canceled")
	case errors.Is(err, quotes.ErrTimeout), errors.Is(err, context.DeadlineExceeded):
		writeError(w, http.StatusGatewayTimeout, "upstream_timeout", "upstream timeout")
	case errors.Is(err, quotes.ErrBadUpstream):
		writeError(w, http.StatusBadGateway, "upstream_parse", "upstream returned bad data")
	case errors.Is(err, quotes.ErrUnavailable):
		writeError(w, http.StatusServiceUnavailable, "upstream_unavailable", "upstream unavailable")
	case errors.Is(err, quotes.ErrRateLimited):
		writeError(w, http.StatusServiceUnavailable, "rate_limited", "upstream rate limited")
	case errors.Is(err, quotes.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "security not found")
	case errors.As(err, &statusErr):
		writeError(w, http.StatusBadGateway, "upstream_status",
			fmt.Sprintf("upstream status %d", statusErr.Code))
	default:
		writeError(w, http.StatusInternalServerError, "internal", "internal error")
	}
}

// writeJSON — единая точка сериализации. Content-Type до WriteHeader,
// иначе заголовок уйдёт после первого байта тела и будет проигнорирован.
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError — тело ошибки. Одна форма на все коды, чтобы фронт
// не разбирал два разных формата.
func writeError(w http.ResponseWriter, code int, reason, msg string) {
	writeJSON(w, code, map[string]string{
		"error":   reason,
		"message": msg,
	})
}
