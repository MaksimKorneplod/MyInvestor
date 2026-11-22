package quotes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MaksimKorneplod/MyInvestor/internal/repository/moex"
	"golang.org/x/sync/singleflight"
)

const (
	defaultTTL         = 3 * time.Second
	maxTickers         = 200
	defaultExecTimeout = 10 * time.Second //таймаут для фонового запроса
	parallelThreshold  = 5
)

// QuoteFetcher — узкий интерфейс к repository. *moex.Client ему удовлетворяет.
type QuoteFetcher interface {
	Quotes(ctx context.Context, ref moex.Instrument) ([]moex.Quote, error)
}

type Service struct {
	client      QuoteFetcher
	cache       *cache
	group       singleflight.Group
	execTimeout time.Duration //храним таймаут
}

func New(client QuoteFetcher, ttl time.Duration) *Service {
	if ttl <= 0 {
		ttl = defaultTTL
	}
	return &Service{client: client, cache: newCache(ttl), execTimeout: defaultExecTimeout}
}

// One — одна бумага. Узкий кэш-ключ, точная свежесть.
// Engine и market остаются в нижнем регистре: ISS принимает /iss/engines/stock/markets/shares.
func (s *Service) One(ctx context.Context, ref moex.Instrument) (*moex.Quote, error) {
	quote, err := s.oneQuote(ctx, ref)
	if err != nil {
		return nil, mapSource(err)
	}
	return quote, nil
}

func (s *Service) oneQuote(ctx context.Context, ref moex.Instrument) (*moex.Quote, error) {
	ref.Security = strings.ToUpper(strings.TrimSpace(ref.Security))
	if err := ident("security", ref.Security); err != nil {
		return nil, err
	}
	key := cacheKey(ref.Engine, ref.Market, ref.Board, ref.Security)
	if rows, ok := s.cache.get(key); ok {
		return pick(rows, ref.Security), nil
	}

	// ИСПОЛЬЗУЕМ DoChan вместо Do
	ch := s.group.DoChan(key, func() (any, error) {
		execCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.execTimeout)
		defer cancel()

		rows, err := s.client.Quotes(execCtx, ref)
		if err != nil {
			return nil, err
		}
		s.cache.put(key, rows)
		return rows, nil
	})

	// Ждем либо результат, либо отмену контекста текущего вызывающего
	select {
	case <-ctx.Done():
		// Лидер упал с context.Canceled, но фоновый запрос (в ch) продолжает жить
		return nil, ctx.Err()
	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		return pick(res.Val.([]moex.Quote), ref.Security), nil
	}
}

// Market — снимок рынка/режима. ref.Security должен быть пустым.
func (s *Service) Market(ctx context.Context, ref moex.Instrument) ([]moex.Quote, error) {
	rows, err := s.market(ctx, ref)
	if err != nil {
		return nil, mapSource(err)
	}
	return rows, nil
}

func (s *Service) market(ctx context.Context, ref moex.Instrument) ([]moex.Quote, error) {
	if ref.Security != "" {
		return nil, fmt.Errorf("%w: market does not take a security", moex.ErrRequest)
	}
	key := cacheKey(ref.Engine, ref.Market, ref.Board, "")
	if rows, ok := s.cache.get(key); ok {
		return rows, nil
	}

	ch := s.group.DoChan(key, func() (any, error) {
		execCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.execTimeout)
		defer cancel()

		rows, err := s.client.Quotes(execCtx, ref)
		if err != nil {
			return nil, err
		}
		if rows == nil {
			rows = []moex.Quote{}
		}
		s.cache.put(key, rows)
		return rows, nil
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		return res.Val.([]moex.Quote), nil
	}
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
