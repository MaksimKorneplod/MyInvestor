package moex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"time"
)

// ChartCandle — свеча графика с OHLC.
// High и Low равны nil, если ISS прислал null.
type ChartCandle struct {
	Begin time.Time
	Open  Decimal
	High  *Decimal
	Low   *Decimal
	Close Decimal
}

// Chart возвращает свечи с high и low. History по-прежнему отдаёт только open и close.
func (c *Client) Chart(ctx context.Context, q HistoryQuery) ([]ChartCandle, error) {
	rows, err := c.chart(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("moex chart: %w", err)
	}
	return rows, nil
}

func (c *Client) chart(ctx context.Context, q HistoryQuery) ([]ChartCandle, error) {
	if q.Interval <= 0 {
		return nil, fmt.Errorf("%w: interval", ErrRequest)
	}
	if q.From.IsZero() || q.Till.IsZero() {
		return nil, fmt.Errorf("%w: from/till", ErrRequest)
	}
	from := q.From.In(moscow).Format(dateLayout)
	till := q.Till.In(moscow).Format(dateLayout)
	if from > till {
		return nil, fmt.Errorf("%w: from after till", ErrRequest)
	}
	path, err := q.Instrument.candlesPath()
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("iss.meta", "off")
	query.Set("from", from)
	query.Set("till", till)
	query.Set("interval", fmt.Sprintf("%d", q.Interval))

	pages, err := c.collect(ctx, path, query, "candles", "begin", "open", "high", "low", "close")
	if err != nil {
		return nil, err
	}
	rows, err := parseChart(pages)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []ChartCandle{}
	}
	return rows, nil
}

func parseChart(pages []table) ([]ChartCandle, error) {
	out := make([]ChartCandle, 0)
	for _, page := range pages {
		index, err := page.indexes("begin", "open", "high", "low", "close")
		if err != nil {
			return nil, err
		}
		for _, row := range page.Data {
			begin, err := chartBegin(row, index["begin"])
			if err != nil {
				return nil, err
			}
			open, err := requiredChartDecimal(row, index["open"], "open")
			if err != nil {
				return nil, err
			}
			closePrice, err := requiredChartDecimal(row, index["close"], "close")
			if err != nil {
				return nil, err
			}
			high, err := optionalChartDecimal(row, index["high"])
			if err != nil {
				return nil, err
			}
			low, err := optionalChartDecimal(row, index["low"])
			if err != nil {
				return nil, err
			}
			out = append(out, ChartCandle{
				Begin: begin,
				Open:  open,
				High:  high,
				Low:   low,
				Close: closePrice,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Begin.Before(out[j].Begin)
	})
	return out, nil
}

func chartBegin(row []json.RawMessage, index int) (time.Time, error) {
	raw, err := at(row, index)
	if err != nil {
		return time.Time{}, err
	}
	text, ok, err := cellString(raw)
	if err != nil {
		return time.Time{}, err
	}
	if !ok {
		return time.Time{}, &ParseError{Reason: "begin"}
	}
	begin, err := time.ParseInLocation(timeLayout, text, moscow)
	if err != nil {
		return time.Time{}, &ParseError{Reason: "begin"}
	}
	return begin, nil
}

func requiredChartDecimal(row []json.RawMessage, index int, name string) (Decimal, error) {
	raw, err := at(row, index)
	if err != nil {
		return Decimal{}, err
	}
	value, ok, err := cellDecimal(raw)
	if err != nil {
		return Decimal{}, err
	}
	if !ok {
		return Decimal{}, &ParseError{Reason: name}
	}
	return value, nil
}

func optionalChartDecimal(row []json.RawMessage, index int) (*Decimal, error) {
	raw, err := at(row, index)
	if err != nil {
		return nil, err
	}
	value, ok, err := cellDecimal(raw)
	if err != nil || !ok {
		return nil, err
	}
	return &value, nil
}
