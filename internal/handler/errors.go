package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/MaksimKorneplod/MyInvestor/internal/repository/moex"
)

// writeMapped превращает ошибку service в HTTP-ответ.
//
// Классы ошибок (см. moex.ErrRequest/ErrParse/ErrNetwork/StatusError):
//
//	ErrRequest             → 400 (вызывающий ошибся)
//	context.Canceled       → 503 (клиент отвалился; не 500)
//	context.DeadlineExceeded → 504 (upstream не успел)
//	ErrNetwork             → 503 (upstream недоступен)
//	ErrParse               → 502 (upstream вернул мусор)
//	StatusError            → 502/503 в зависимости от кода ISS
//	неизвестное            → 500
//
// Наружу отдаём только стабильный code и короткое сообщение.
// Текст ошибки ISS, URL, детали парсинга — только в логи, не в JSON.
func writeMapped(w http.ResponseWriter, err error) {
	var statusErr *moex.StatusError
	switch {
	case errors.Is(err, moex.ErrRequest):
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
	case errors.Is(err, context.Canceled):
		writeError(w, http.StatusServiceUnavailable, "canceled", "request canceled")
	case errors.Is(err, context.DeadlineExceeded):
		writeError(w, http.StatusGatewayTimeout, "upstream_timeout", "upstream timeout")
	case errors.Is(err, moex.ErrParse):
		writeError(w, http.StatusBadGateway, "upstream_parse", "upstream returned bad data")
	case errors.Is(err, moex.ErrNetwork):
		writeError(w, http.StatusServiceUnavailable, "upstream_unavailable", "upstream unavailable")
	case errors.As(err, &statusErr):
		switch statusErr.Code {
		case http.StatusTooManyRequests:
			writeError(w, http.StatusServiceUnavailable, "rate_limited", "upstream rate limited")
		case http.StatusNotFound:
			writeError(w, http.StatusNotFound, "not_found", "upstream: not found")
		default:
			writeError(w, http.StatusBadGateway, "upstream_status",
				fmt.Sprintf("upstream status %d", statusErr.Code))
		}
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
