package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/MaksimKorneplod/MyInvestor/internal/repository/moex"
)

// Instrument represents a stock instrument with stable identifier, ticker and display name.
type Instrument struct {
	ID     string `json:"id"`     // stable identifier (ISIN)
	Ticker string `json:"ticker"` // exchange code (SECID), e.g. SBER
	Name   string `json:"name"`   // display name (SHORTNAME)
}

// StockSource describes where the catalog data comes from.
type StockSource interface {
	ListStocks(ctx context.Context) ([]Instrument, error)
}

type moexSecurityLister interface {
	List(ctx context.Context) ([]moex.Security, error)
}

type moexStockSource struct {
	lister moexSecurityLister
}

var _ StockSource = (*moexStockSource)(nil)

// NewMOEXStockSource converts MOEX securities into service instruments.
func NewMOEXStockSource(lister moexSecurityLister) StockSource {
	return &moexStockSource{lister: lister}
}

func (s *moexStockSource) ListStocks(ctx context.Context) ([]Instrument, error) {
	securities, err := s.lister.List(ctx)
	if err != nil {
		return nil, err
	}

	instruments := make([]Instrument, 0, len(securities))
	for _, security := range securities {
		if security.SecID == "" {
			continue
		}
		name := security.ShortName
		if name == "" {
			name = security.SecID
		}
		instruments = append(instruments, Instrument{
			ID:     security.SecID,
			Ticker: security.SecID,
			Name:   name,
		})
	}
	return instruments, nil
}

// ErrSourceUnavailable indicates that the external data source is temporarily unavailable.
var ErrSourceUnavailable = errors.New("market data source is temporarily unavailable")

const stockSourceLoadTimeout = 30 * time.Second

// Page is the paginated result of a search request.
type Page struct {
	Items []Instrument
	Total int
}

// Catalog caches a list of instruments returned by source for ttl duration.
type Catalog struct {
	source      StockSource
	ttl         time.Duration
	loadTimeout time.Duration
	cache       []Instrument
	loaded      bool
	expiresAt   time.Time
	inflight    *stockLoad
	mu          sync.Mutex
	now         func() time.Time
}

type stockLoad struct {
	done  chan struct{}
	items []Instrument
	err   error
}

// NewCatalog creates a new Catalog with given source and TTL.
func NewCatalog(source StockSource, ttl time.Duration) *Catalog {
	return &Catalog{
		source:      source,
		ttl:         ttl,
		loadTimeout: stockSourceLoadTimeout,
		now:         time.Now,
	}
}

// Search performs a search across the catalog with pagination.
func (c *Catalog) Search(ctx context.Context, query string, limit, offset int) (Page, error) {
	items, err := c.listStocks(ctx)
	if err != nil {
		return Page{}, err
	}
	fed := filterByQuery(query, items)
	sorted := sortInstruments(fed)
	total := len(sorted)
	page := Page{Items: make([]Instrument, 0, total), Total: total}
	if offset < 0 || offset > total {
		return page, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	page.Items = sorted[offset:end]
	return page, nil
}

func (c *Catalog) listStocks(ctx context.Context) ([]Instrument, error) {
	if ctx == nil {
		return nil, errors.New("nil context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	c.mu.Lock()
	if err := ctx.Err(); err != nil {
		c.mu.Unlock()
		return nil, err
	}
	now := c.now()
	if c.loaded && now.Before(c.expiresAt) {
		items := cloneInstruments(c.cache)
		c.mu.Unlock()
		return items, nil
	}

	if load := c.inflight; load != nil {
		c.mu.Unlock()
		return waitForStockLoad(ctx, load)
	}

	load := &stockLoad{done: make(chan struct{})}
	c.inflight = load
	loadCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), c.loadTimeout)
	go c.loadStocks(loadCtx, cancel, load)
	c.mu.Unlock()
	return waitForStockLoad(ctx, load)
}

func waitForStockLoad(ctx context.Context, load *stockLoad) ([]Instrument, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-load.done:
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if load.err != nil {
		return nil, load.err
	}
	return cloneInstruments(load.items), nil
}

func (c *Catalog) loadStocks(ctx context.Context, cancel context.CancelFunc, load *stockLoad) {
	defer cancel()

	items, err := c.source.ListStocks(ctx)
	if err != nil {
		err = fmt.Errorf("%w: %w", ErrSourceUnavailable, err)
	}

	c.mu.Lock()
	if err == nil {
		now := c.now()
		if !c.loaded || !now.Before(c.expiresAt) {
			c.cache = cloneInstruments(items)
			c.loaded = true
			c.expiresAt = now.Add(c.ttl)
		}
		load.items = cloneInstruments(c.cache)
	}
	load.err = err
	if c.inflight == load {
		c.inflight = nil
	}
	close(load.done)
	c.mu.Unlock()
}

func cloneInstruments(items []Instrument) []Instrument {
	cloned := make([]Instrument, len(items))
	copy(cloned, items)
	return cloned
}

func filterByQuery(query string, items []Instrument) []Instrument {
	trimmed := strings.TrimSpace(query)
	lower := strings.ToLower(trimmed)
	if lower == "" {
		out := make([]Instrument, len(items))
		copy(out, items)
		return out
	}
	var filtered []Instrument
	for _, inst := range items {
		lticker := strings.ToLower(inst.Ticker)
		lname := strings.ToLower(inst.Name)
		if strings.Contains(lticker, lower) || strings.Contains(lname, lower) {
			filtered = append(filtered, inst)
		}
	}
	return filtered
}

func sortInstruments(items []Instrument) []Instrument {
	sorted := make([]Instrument, len(items))
	copy(sorted, items)
	sort.SliceStable(sorted, func(i, j int) bool {
		a, b := strings.ToLower(sorted[i].Ticker), strings.ToLower(sorted[j].Ticker)
		if a != b {
			return a < b
		}
		return sorted[i].ID < sorted[j].ID
	})
	return sorted
}
