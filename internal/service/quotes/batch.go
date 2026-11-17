package quotes

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/MaksimKorneplod/MyInvestor/internal/repository/moex"
	"golang.org/x/sync/errgroup"
)

var tickerRe = regexp.MustCompile(`^[A-Za-z0-9._-]{1,32}$`)

// Batch — точный список тикеров.
//
// Контракт наружу один: []moex.Quote + error. Внутри — три стратегии:
//
//	1 тикер       → One (узкий кэш, точная свежесть)
//	2..5 тикеров  → parallel: errgroup с лимитом, частичная устойчивость
//	>5 тикеров    → bulk: один запрос на рынок + фильтрация
//
// Порядок ответа соответствует порядку первого вхождения тикера
// во входном списке. Отсутствующие в ISS бумаги просто не попадают
// в результат — фронт дорисует прочерк.
func (s *Service) Batch(ctx context.Context, ref moex.Instrument, tickers []string) ([]moex.Quote, error) {
	uniq, err := normalizeTickers(tickers)
	if err != nil {
		return nil, err
	}
	switch {
	case len(uniq) == 0:
		return []moex.Quote{}, nil
	case len(uniq) == 1:
		return s.one(ctx, ref, uniq[0])
	case len(uniq) <= parallelThreshold:
		return s.parallel(ctx, ref, uniq)
	default:
		return s.bulk(ctx, ref, uniq)
	}
}

func (s *Service) one(ctx context.Context, ref moex.Instrument, ticker string) ([]moex.Quote, error) {
	r := ref
	r.Security = ticker
	q, err := s.One(ctx, r)
	if err != nil {
		return nil, err
	}
	if q == nil {
		return []moex.Quote{}, nil
	}
	return []moex.Quote{*q}, nil
}

// parallel — до parallelThreshold тикеров. Каждый идёт своим per-security
// вызовом через One, поэтому у каждого свой узкий кэш-ключ и свой
// singleflight. Если одна бумага упала — весь batch падает (контракт
// «всё или ничего»), но кэш успешных бумаг уже заполнен и переживёт ошибку.
func (s *Service) parallel(ctx context.Context, ref moex.Instrument, tickers []string) ([]moex.Quote, error) {
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(parallelThreshold)

	out := make([]*moex.Quote, len(tickers))
	for i, t := range tickers {
		i, t := i, t
		g.Go(func() error {
			r := ref
			r.Security = t
			q, err := s.One(gctx, r)
			if err != nil {
				return fmt.Errorf("moex quotes %s: %w", t, err)
			}
			out[i] = q
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	res := make([]moex.Quote, 0, len(tickers))
	for _, q := range out {
		if q != nil {
			res = append(res, *q)
		}
	}
	return res, nil
}

// bulk — больше parallelThreshold тикеров. Один запрос на весь рынок
// через Market (широкий кэш-ключ) + фильтрация по нужным SECID.
// Согласованный снимок: все цены получены в один момент.
func (s *Service) bulk(ctx context.Context, ref moex.Instrument, tickers []string) ([]moex.Quote, error) {
	r := ref
	r.Security = ""
	all, err := s.Market(ctx, r)
	if err != nil {
		return nil, err
	}

	wanted := make(map[string]int, len(tickers))
	for i, t := range tickers {
		wanted[t] = i
	}

	// Сохраняем порядок запроса: кладём в позицию по wanted.
	res := make([]*moex.Quote, len(tickers))
	for i := range all {
		if pos, ok := wanted[all[i].SecID]; ok {
			q := all[i]
			res[pos] = &q
		}
	}

	out := make([]moex.Quote, 0, len(tickers))
	for _, q := range res {
		if q != nil {
			out = append(out, *q)
		}
	}
	return out, nil
}

// normalizeTickers: TrimSpace → ToUpper → regexp → dedupe → лимит ПОСЛЕ dedupe.
// Порядок первого вхождения сохраняется.
func normalizeTickers(tickers []string) ([]string, error) {
	if len(tickers) == 0 {
		return nil, fmt.Errorf("%w: empty tickers", moex.ErrRequest)
	}
	//мапчик на вход
	tickersIn := make(map[string]struct{})
	//мапчик на выход
	tickersOut := make([]string, 0, len(tickers))
	//проверка на минетик
	for _, quot := range tickers {
		quotUpRegister := strings.ToUpper(quot)
		if quotUpRegister == "" {
			return nil, fmt.Errorf("%w: empty ticker", moex.ErrRequest)
		}
		if !tickerRegular.MatchString(quotUpRegister) {
			return nil, fmt.Errorf("%w: wrong ticker", moex.ErrRequest)
		}
		if _, ok := tickersIn[quotUpRegister]; ok {
			continue
		}
		tickersIn[quotUpRegister] = struct{}{}
		tickersOut = append(tickersOut, quotUpRegister)
	}
	if len(tickersOut) > maxTickers {
		return nil, fmt.Errorf("%w: too many tickers in ask (%d)", moex.ErrRequest, len(tickersOut))
	}
	return tickersOut, nil
}

func ident(kind, value string) error {
	if !tickerRe.MatchString(value) {
		return fmt.Errorf("%w: %s %q", moex.ErrRequest, kind, value)
	}
	return nil
}

func cacheKey(engine, market, board, security string) string {
	return strings.Join([]string{engine, market, board, security}, "|")
}
