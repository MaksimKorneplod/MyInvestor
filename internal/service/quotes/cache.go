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
	mu   sync.RWMutex
	ttl  time.Duration
	m    map[string]cacheEntry
	stop chan struct{} // Канал для остановки фоновой очистки
}

func newCache(ttl time.Duration) *cache {
	c := &cache{
		ttl:  ttl,
		m:    make(map[string]cacheEntry),
		stop: make(chan struct{}),
	}
	// Запускаем фоновую очистку
	go c.janitor()
	return c
}

// Close останавливает фоновую горутину очистки.
// Полезно вызывать при завершении приложения (Graceful Shutdown).
func (c *cache) Close() {
	close(c.stop)
}

func (c *cache) get(key string) ([]moex.Quote, bool) {
	c.mu.RLock()
	e, ok := c.m[key]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}

	// Ленивое удаление: если элемент найден, но просрочен — удаляем его из map
	if time.Since(e.at) > c.ttl {
		c.mu.Lock()
		// Проверяем еще раз под Lock, так как другая горутина могла уже удалить или обновить
		if current, ok := c.m[key]; ok && time.Since(current.at) > c.ttl {
			delete(c.m, key)
		}
		c.mu.Unlock()
		return nil, false
	}

	return e.rows, true
}

func (c *cache) put(key string, rows []moex.Quote) {
	c.mu.Lock()
	c.m[key] = cacheEntry{rows: rows, at: time.Now()}
	c.mu.Unlock()
}

// janitor — фоновая горутина, которая периодически чистит просроченные записи.
func (c *cache) janitor() {
	// Интервал очистки равен TTL.
	// Если TTL 3 секунды, мы чистим каждые 3 секунды.
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stop:
			return
		}
	}
}

func (c *cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.m {
		if now.Sub(entry.at) > c.ttl {
			delete(c.m, key)
		}
	}
}
