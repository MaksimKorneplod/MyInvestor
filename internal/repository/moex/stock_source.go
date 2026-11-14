package moex

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/MaksimKorneplod/MyInvestor/internal/service"
)

// StockAdapter implements service.StockSource: it provides catalog instruments from MOEX.
type StockAdapter struct {
	client *Client
}

var _ service.StockSource = (*StockAdapter)(nil)

// NewStockAdapter creates a StockAdapter backed by the given MOEX client.
func NewStockAdapter(client *Client) *StockAdapter {
	return &StockAdapter{client: client}
}

// ListStocks returns the catalog of instruments as service.Instrument.
func (s *StockAdapter) ListStocks(ctx context.Context) ([]service.Instrument, error) {
	if s.client == nil {
		err := errors.New("moex client is nil")
		slog.Error("moex stock source failed", "error", err)
		return nil, fmt.Errorf("%w: %w", service.ErrSourceUnavailable, err)
	}

	rows, err := s.client.Securities(ctx, SecuritiesQuery{
		Instrument: Instrument{Engine: "stock", Market: "shares", Board: "TQBR"},
	})
	if err != nil {
		if ctx != nil && ctx.Err() != nil {
			return nil, ctx.Err()
		}
		slog.Error("moex stock source failed", "error", err)
		return nil, fmt.Errorf("%w: %w", service.ErrSourceUnavailable, err)
	}
	return mapSecurities(rows), nil
}

func mapSecurities(rows []Security) []service.Instrument {
	instruments := make([]service.Instrument, 0, len(rows))
	for _, row := range rows {
		if row.SecID == "" {
			continue
		}
		name := row.ShortName
		if name == "" {
			name = row.SecID
		}
		instruments = append(instruments, service.Instrument{
			ID:     row.SecID, // ID is SECID because moex.Security has no ISIN.
			Ticker: row.SecID,
			Name:   name,
		})
	}
	return instruments
}
