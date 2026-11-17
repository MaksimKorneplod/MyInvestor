package quotes

import (
	"sync"
	"time"

	"github.com/MaksimKorneplod/MyInvestor/internal/repository/moex"
)

type cacheEntry struct {
	rows []moex.Quote
	at   time.Time
}

type cache struct {
	mu  sync.RWMutex
	ttl time.Duration
	m   map[string]cacheEntry
}

func newCache(ttl time.Duration) *cache {
	return &cache{ttl: ttl, m: make(map[string]cacheEntry)}
}

func (c *cache) get(key string) ([]moex.Quote, bool) {
	c.mu.RLock()
	e, ok := c.m[key]
	c.mu.RUnlock()
	if !ok || time.Since(e.at) > c.ttl {
		return nil, false
	}
	return e.rows, true
}

func (c *cache) put(key string, rows []moex.Quote) {
	c.mu.Lock()
	c.m[key] = cacheEntry{rows: rows, at: time.Now()}
	c.mu.Unlock()
}
