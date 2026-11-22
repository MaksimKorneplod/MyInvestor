package handler

import (
	"time"

	"github.com/MaksimKorneplod/MyInvestor/internal/service/quotes"
)

// BatchRequest — тело POST /v1/quotes:batch.
type BatchRequest struct {
	Engine  string   `json:"engine"`
	Market  string   `json:"market"`
	Board   string   `json:"board,omitempty"`
	Tickers []string `json:"tickers"`
}

// QuoteDTO — элемент ответа. Все три поля с указателем:
// null в ISS → omitempty в JSON, фронт отличает «нет данных» от «0».
type QuoteDTO struct {
	Ticker     string  `json:"ticker"`
	Last       *string `json:"last,omitempty"`
	ChangePct  *string `json:"changePercent,omitempty"`
	QuoteTime  *string `json:"quoteTime,omitempty"`
	ReceivedAt string  `json:"receivedAt"`
}

// toDTO — домен → DTO. Единственное место, где moex.Quote
// превращается в то, что видит фронт.
func toDTO(q quotes.Quote) QuoteDTO {
	dto := QuoteDTO{
		Ticker:     q.SecID,
		ReceivedAt: q.UpdateTime.Format(time.RFC3339),
	}
	if q.Last != nil {
		s := q.Last.String()
		dto.Last = &s
	}
	if q.LastChangePercent != nil {
		s := q.LastChangePercent.String()
		dto.ChangePct = &s
	}
	if q.QuoteTime != nil {
		s := q.QuoteTime.Format(time.RFC3339)
		dto.QuoteTime = &s
	}
	return dto
}

func toDTOs(rows []quotes.Quote) []QuoteDTO {
	if rows == nil {
		return []QuoteDTO{}
	}
	out := make([]QuoteDTO, len(rows))
	for i, r := range rows {
		out[i] = toDTO(r)
	}
	return out
}
