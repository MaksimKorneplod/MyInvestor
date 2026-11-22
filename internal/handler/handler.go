// Котировки. HTTP-слой поверх internal/service/quotes.
//
// Отвечает только за четыре вещи:
//  1. HTTP: методы, коды, заголовки, тела.
//  2. Валидация входа: whitelist engine/market/board, размер тела, JSON.
//  3. DTO: quotes.Quote → QuoteDTO.
//  4. Маппинг ошибок service → HTTP-код.
//
// Репозиторий MOEX, кэш и singleflight этому пакету не видны.
package handler

import (
	"context"
	"net/http"

	"github.com/MaksimKorneplod/MyInvestor/internal/service/quotes"
)

// Service — то, что обработчику нужно от бизнес-слоя.
// *quotes.Service ему удовлетворяет.
type Service interface {
	One(ctx context.Context, ref quotes.Instrument) (*quotes.Quote, error)
	Market(ctx context.Context, ref quotes.Instrument) ([]quotes.Quote, error)
	Batch(ctx context.Context, ref quotes.Instrument, tickers []string) ([]quotes.Quote, error)
}

type Handler struct {
	svc Service
}

func New(svc Service) *Handler { return &Handler{svc: svc} }

// Register навешивает маршруты (Go 1.22+ patterns).
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/quotes", h.handleMarket)
	mux.HandleFunc("POST /api/v1/quotes:batch", h.handleBatch)
	mux.HandleFunc("GET /api/v1/quotes/{engine}/{market}/{security}", h.handleOne)
}

// handleMarket — GET /api/v1/quotes?engine=stock&market=shares&board=TQBR
// Снимок рынка/режима: экран котировок и рейтинг.
func (h *Handler) handleMarket(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	ref, err := validateInstrument(q.Get("engine"), q.Get("market"), q.Get("board"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_market", err.Error())
		return
	}
	rows, err := h.svc.Market(r.Context(), ref)
	if err != nil {
		writeMapped(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toDTOs(rows))
}

// handleBatch — POST /api/v1/quotes:batch
// Точный список тикеров: портфель.
func (h *Handler) handleBatch(w http.ResponseWriter, r *http.Request) {
	req, err := decodeBatch(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_body", err.Error())
		return
	}
	ref, err := validateInstrument(req.Engine, req.Market, req.Board)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_market", err.Error())
		return
	}
	if len(req.Tickers) == 0 {
		writeError(w, http.StatusBadRequest, "bad_tickers", "empty tickers")
		return
	}
	rows, err := h.svc.Batch(r.Context(), ref, req.Tickers)
	if err != nil {
		writeMapped(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toDTOs(rows))
}

// handleOne — GET /api/v1/quotes/{engine}/{market}/{security}?board=TQBR
// Одна бумага.
func (h *Handler) handleOne(w http.ResponseWriter, r *http.Request) {
	board := r.URL.Query().Get("board")
	ref, err := validateInstrument(r.PathValue("engine"), r.PathValue("market"), board)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_market", err.Error())
		return
	}
	ref.Security = r.PathValue("security")
	if err := validateTicker(ref.Security); err != nil {
		writeError(w, http.StatusBadRequest, "bad_security", err.Error())
		return
	}
	q, err := h.svc.One(r.Context(), ref)
	if err != nil {
		writeMapped(w, err)
		return
	}
	if q == nil {
		writeError(w, http.StatusNotFound, "not_found", "security not found")
		return
	}
	writeJSON(w, http.StatusOK, toDTO(*q))
}
