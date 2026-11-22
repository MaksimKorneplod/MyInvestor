// Package moex — репозиторий внешнего API MOEX ISS.
//
// В MyInvestor это internal/repository: единственная точка для списка бумаг,
// котировок и свечей. Хендлер получает уже собранные структуры и не знает URL ISS.
//
// Список и котировки повторяют Security+fetchSecurity.swift.
// История графика повторяет Candle+fetchCandles.swift: это свечи, не дневной /iss/history.
// ChartView.swift только рисует эти свечи и отдельных запросов не добавляет.
package moex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://iss.moex.com"
	defaultTimeout = 15 * time.Second
	maxBodyBytes   = 10 << 20
)

// Client ходит в ISS. Его создают через New и передают в хендлеры Go API.
type Client struct {
	baseURL string
	http    *http.Client
	timeout time.Duration
	ownHTTP bool
}

// Option настраивает Client.
type Option func(*Client)

// New создаёт клиент. Пустой baseURL означает https://iss.moex.com.
// Таймаут по умолчанию — 15 секунд.
func New(baseURL string, opts ...Option) *Client {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultBaseURL
	}
	c := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: defaultTimeout},
		timeout: defaultTimeout,
		ownHTTP: true,
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.http == nil {
		c.http = &http.Client{Timeout: c.timeout}
		c.ownHTTP = true
	}
	if c.timeout <= 0 {
		c.timeout = defaultTimeout
	}
	return c
}

// WithTimeout задаёт предел ожидания ответа.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		if d <= 0 {
			return
		}
		c.timeout = d
		if c.ownHTTP && c.http != nil {
			c.http.Timeout = d
		}
	}
}

// WithHTTPClient подставляет транспорт. Его собственный Timeout не переписывается.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.http = httpClient
		c.ownHTTP = false
	}
}

// Securities возвращает список бумаг.
func (c *Client) Securities(ctx context.Context, q SecuritiesQuery) ([]Security, error) {
	rows, err := c.securities(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("moex securities: %w", err)
	}
	return rows, nil
}

func (c *Client) securities(ctx context.Context, q SecuritiesQuery) ([]Security, error) {
	if q.Instrument.Security != "" {
		return nil, fmt.Errorf("%w: list does not take a security", ErrRequest)
	}
	path, err := q.Instrument.marketPath()
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("iss.meta", "off")
	query.Set("iss.only", "securities,securities.cursor")
	query.Set("securities.columns", "SECID,SHORTNAME")

	pages, err := c.collect(ctx, path+"/securities.json", query, "securities", "SECID", "SHORTNAME")
	if err != nil {
		return nil, err
	}
	rows, err := parseSecurities(pages)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []Security{}
	}
	return rows, nil
}

// Quotes возвращает последние цены.
// Пустой Security — все бумаги рынка, как один запрос marketdata в iOS.
func (c *Client) Quotes(ctx context.Context, ref Instrument) ([]Quote, error) {
	rows, err := c.quotes(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("moex quotes: %w", err)
	}
	return rows, nil
}

func (c *Client) quotes(ctx context.Context, ref Instrument) ([]Quote, error) {
	path, err := ref.quotePath()
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("iss.meta", "off")
	query.Set("iss.only", "marketdata,marketdata.cursor")
	query.Set("marketdata.columns", "SECID,LAST,LASTCHANGEPRCNT")

	pages, err := c.collect(ctx, path, query, "marketdata", "SECID", "LAST", "LASTCHANGEPRCNT")
	if err != nil {
		return nil, err
	}
	rows, err := parseQuotes(pages)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []Quote{}
	}
	return rows, nil
}

// History возвращает свечи и сам листает страницы.
func (c *Client) History(ctx context.Context, q HistoryQuery) ([]HistoryRow, error) {
	rows, err := c.history(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("moex history: %w", err)
	}
	return rows, nil
}

func (c *Client) history(ctx context.Context, q HistoryQuery) ([]HistoryRow, error) {
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

	pages, err := c.collect(ctx, path, query, "candles", "begin", "open", "close")
	if err != nil {
		return nil, err
	}
	rows, err := parseHistory(pages)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []HistoryRow{}
	}
	return rows, nil
}

func (in Instrument) marketPath() (string, error) {
	if err := ident("engine", in.Engine); err != nil {
		return "", err
	}
	if err := ident("market", in.Market); err != nil {
		return "", err
	}
	path := "/iss/engines/" + in.Engine + "/markets/" + in.Market
	if in.Board != "" {
		if err := ident("board", in.Board); err != nil {
			return "", err
		}
		path += "/boards/" + in.Board
	}
	return path, nil
}

func (in Instrument) quotePath() (string, error) {
	base, err := in.marketPath()
	if err != nil {
		return "", err
	}
	if in.Security == "" {
		return base + "/securities.json", nil
	}
	if err := ident("security", in.Security); err != nil {
		return "", err
	}
	return base + "/securities/" + in.Security + ".json", nil
}

func (in Instrument) candlesPath() (string, error) {
	if err := ident("security", in.Security); err != nil {
		return "", err
	}
	base, err := in.marketPath()
	if err != nil {
		return "", err
	}
	return base + "/securities/" + in.Security + "/candles.json", nil
}

func (c *Client) collect(ctx context.Context, path string, query url.Values, block string, columns ...string) ([]table, error) {
	q := cloneValues(query)
	pages := make([]table, 0, 1)
	start := 0
	previousRow := ""
	for page := 0; page < maxPages; page++ {
		q.Set("start", fmt.Sprintf("%d", start))
		body, err := c.get(ctx, path, q)
		if err != nil {
			return nil, err
		}
		tables, err := decodeTables(body)
		if err != nil {
			return nil, err
		}
		current, ok := tables[block]
		if !ok {
			return nil, &ParseError{Reason: "missing block " + block}
		}
		if _, err := current.indexes(columns...); err != nil {
			return nil, err
		}
		if len(current.Data) == 0 {
			return pages, nil
		}
		// Один инструмент ISS отдаёт одну и ту же строку на любой start и без cursor.
		// Повтор первой строки значит, что страниц больше нет.
		key := rowKey(current.Data[0])
		if page > 0 && key == previousRow {
			return pages, nil
		}
		previousRow = key
		pages = append(pages, current)

		cursor, hasCursor := tables[block+".cursor"]
		if hasCursor {
			next, more, err := cursorNext(cursor, start)
			if err != nil {
				return nil, err
			}
			if !more {
				return pages, nil
			}
			start = next
			continue
		}
		start += len(current.Data)
	}
	return nil, &ParseError{Reason: "too many pages"}
}

func (c *Client) get(ctx context.Context, path string, query url.Values) ([]byte, error) {
	if ctx == nil {
		return nil, fmt.Errorf("%w: nil context", ErrRequest)
	}
	if c.http == nil {
		return nil, fmt.Errorf("%w: nil http client", ErrRequest)
	}
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, c.timeout)
	defer cancel()

	endpoint, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("%w: url", ErrRequest)
	}
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrRequest, err.Error())
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNetwork, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNetwork, err)
	}
	if len(body) > maxBodyBytes {
		return nil, &ParseError{Reason: "response too large"}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &StatusError{Code: resp.StatusCode}
	}
	return body, nil
}

func rowKey(row []json.RawMessage) string {
	var b strings.Builder
	for _, cell := range row {
		b.Write(cell)
		b.WriteByte(0)
	}
	return b.String()
}

func cloneValues(query url.Values) url.Values {
	out := make(url.Values, len(query))
	for key, values := range query {
		out[key] = append([]string(nil), values...)
	}
	return out
}
