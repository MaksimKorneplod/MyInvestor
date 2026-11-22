package moex

import (
	"context"
	"fmt"
)

// StockLister provides securities from the MOEX shares market.
type StockLister struct {
	client *Client
}

// NewStockLister creates a StockLister backed by the given MOEX client.
func NewStockLister(client *Client) *StockLister {
	return &StockLister{client: client}
}

// List returns securities from the TQBR board.
func (s *StockLister) List(ctx context.Context) ([]Security, error) {
	if s.client == nil {
		return nil, fmt.Errorf("%w: nil MOEX client", ErrRequest)
	}

	return s.client.Securities(ctx, SecuritiesQuery{
		Instrument: Instrument{Engine: "stock", Market: "shares", Board: "TQBR"},
	})
}
