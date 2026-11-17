package quotes

import (
	"context"
	"fmt"
	"time"

	"github.com/MaksimKorneplod/MyInvestor/internal/repository/moex"
	"golang.org/x/sync/singleflight"
)

const (
	defaultTTL        = 3 * time.Second
	maxTickers        = 200
	parallelThreshold = 5
)

// QuoteFetcher — узкий интерфейс к repository. *moex.Client ему удовлетворяет.
type QuoteFetcher interface {
	Quotes(ctx context.Context, ref moex.Instrument) ([]moex.Quote, error)
}

type Service struct {
	client QuoteFetcher
	cache  *cache
	group  singleflight.Group
}

func New(client QuoteFetcher, ttl time.Duration) *Service {
	if ttl <= 0 {
		ttl = defaultTTL
	}
	return &Service{client: client, cache: newCache(ttl)}
}

// One — одна бумага. Узкий кэш-ключ, точная свежесть.
// (nil, nil) — бумаги нет в выдаче ISS, это не ошибка.
func (s *Service) One(ctx context.Context, ref moex.Instrument) (*moex.Quote, error) {
	if err := ident("security", ref.Security); err != nil {
		return nil, err
	}
	key := cacheKey(ref.Engine, ref.Market, ref.Board, ref.Security)
	if rows, ok := s.cache.get(key); ok {
		return pick(rows, ref.Security), nil
	}
	v, err, _ := s.group.Do(key, func() (any, error) {
		rows, err := s.client.Quotes(ctx, ref)
		if err != nil {
			return nil, err
		}
		s.cache.put(key, rows)
		return rows, nil
	})
	if err != nil {
		return nil, err
	}
	return pick(v.([]moex.Quote), ref.Security), nil
}

// Market — снимок рынка/режима. ref.Security должен быть пустым.
func (s *Service) Market(ctx context.Context, ref moex.Instrument) ([]moex.Quote, error) {
	if ref.Security != "" {
		return nil, fmt.Errorf("%w: market does not take a security", moex.ErrRequest)
	}
	key := cacheKey(ref.Engine, ref.Market, ref.Board, "")
	if rows, ok := s.cache.get(key); ok {
		return rows, nil
	}
	v, err, _ := s.group.Do(key, func() (any, error) {
		rows, err := s.client.Quotes(ctx, ref)
		if err != nil {
			return nil, err
		}
		if rows == nil {
			rows = []moex.Quote{}
		}
		s.cache.put(key, rows)
		return rows, nil
	})
	if err != nil {
		return nil, err
	}
	return v.([]moex.Quote), nil
}

func pick(rows []moex.Quote, secid string) *moex.Quote {
	for i := range rows {
		if rows[i].SecID == secid {
			q := rows[i]
			return &q
		}
	}
	return nil
}
