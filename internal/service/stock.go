package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
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

// ErrSourceUnavailable indicates that the external data source is temporarily unavailable.
var ErrSourceUnavailable = errors.New("market data source is temporarily unavailable")

// Page is the paginated result of a search request.
type Page struct {
	Items []Instrument
	Total int
}

// Catalog caches a list of instruments returned by source for ttl duration.
type Catalog struct {
	source    StockSource
	ttl       time.Duration
	cache     []Instrument
	expiresAt time.Time
	mu        sync.RWMutex
	now       func() time.Time
}

// NewCatalog creates a new Catalog with given source and TTL.
func NewCatalog(source StockSource, ttl time.Duration) *Catalog {
	return &Catalog{
		source: source,
		ttl:    ttl,
		now:    time.Now,
	}
}

// Search performs a search across the catalog with pagination.
func (c *Catalog) Search(ctx context.Context, query string, limit, offset int) (Page, error) {
	items, err := c.listStocksLocked(ctx)
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

func (c *Catalog) listStocksLocked(ctx context.Context) ([]Instrument, error) {
	c.mu.RLock()
	if len(c.cache) > 0 && !c.expiresAt.IsZero() && c.expiresAt.After(c.now()) {
		items := make([]Instrument, len(c.cache))
		copy(items, c.cache)
		c.mu.RUnlock()
		return items, nil
	}
	c.mu.RUnlock()

	raw, err := c.source.ListStocks(ctx)
	if err != nil {
		return nil, ErrSourceUnavailable
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.cache) == 0 || c.expiresAt.Before(c.now()) {
		c.cache = raw
		c.expiresAt = c.now().Add(c.ttl)
	}
	result := make([]Instrument, len(c.cache))
	copy(result, c.cache)
	return result, nil
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
