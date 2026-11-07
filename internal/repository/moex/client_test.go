package moex

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSecuritiesListAndPages(t *testing.T) {
	t.Parallel()
	var starts []string
	var path, only, columns string
	client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		starts = append(starts, r.URL.Query().Get("start"))
		path = r.URL.Path
		only = r.URL.Query().Get("iss.only")
		columns = r.URL.Query().Get("securities.columns")
		if r.URL.Query().Get("start") == "0" {
			_, _ = io.WriteString(w, `{
				"securities": {"columns":["SECID","SHORTNAME"],"data":[["SBER","Сбербанк"],["GAZP",null]]},
				"securities.cursor": {"columns":["INDEX","TOTAL","PAGESIZE"],"data":[[0,3,2]]}
			}`)
			return
		}
		_, _ = io.WriteString(w, `{
			"securities": {"columns":["SHORTNAME","SECID"],"data":[["Лукойл","LKOH"]]},
			"securities.cursor": {"columns":["INDEX","TOTAL","PAGESIZE"],"data":[[2,3,2]]}
		}`)
	})

	list, err := client.Securities(context.Background(), SecuritiesQuery{
		Instrument: Instrument{Engine: "stock", Market: "shares"},
	})
	if err != nil {
		t.Fatalf("Securities: %v", err)
	}
	if path != "/iss/engines/stock/markets/shares/securities.json" {
		t.Fatalf("path = %s", path)
	}
	if only != "securities,securities.cursor" || columns != "SECID,SHORTNAME" {
		t.Fatalf("only = %s columns = %s", only, columns)
	}
	if join(starts) != "0,2" {
		t.Fatalf("starts = %v", starts)
	}
	want := []Security{
		{SecID: "SBER", ShortName: "Сбербанк"},
		{SecID: "GAZP", ShortName: "GAZP"},
		{SecID: "LKOH", ShortName: "Лукойл"},
	}
	if len(list) != len(want) {
		t.Fatalf("list = %+v", list)
	}
	for i := range want {
		if list[i] != want[i] {
			t.Fatalf("row %d = %+v", i, list[i])
		}
	}
}

func TestQuotesBoardSecurityAndNulls(t *testing.T) {
	t.Parallel()
	var path string
	client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		if r.URL.Query().Get("marketdata.columns") != "SECID,LAST,LASTCHANGEPRCNT" {
			t.Errorf("columns = %s", r.URL.Query().Get("marketdata.columns"))
		}
		if r.URL.Query().Get("start") == "0" {
			_, _ = io.WriteString(w, `{
				"marketdata": {
					"columns": ["SECID","LAST","LASTCHANGEPRCNT"],
					"data": [["SBER",300.5,1.25],["GAZP",null,2]]
				},
				"marketdata.cursor": {"columns":["INDEX","TOTAL","PAGESIZE"],"data":[[0,3,2]]}
			}`)
			return
		}
		_, _ = io.WriteString(w, `{
			"marketdata": {
				"columns": ["LASTCHANGEPRCNT","LAST","SECID"],
				"data": [[null,7000,"LKOH"]]
			},
			"marketdata.cursor": {"columns":["INDEX","TOTAL","PAGESIZE"],"data":[[2,3,2]]}
		}`)
	})

	quotes, err := client.Quotes(context.Background(), Instrument{
		Engine: "stock", Market: "shares", Board: "TQBR", Security: "SBER",
	})
	if err != nil {
		t.Fatalf("Quotes: %v", err)
	}
	if path != "/iss/engines/stock/markets/shares/boards/TQBR/securities/SBER.json" {
		t.Fatalf("path = %s", path)
	}
	if len(quotes) != 2 {
		t.Fatalf("quotes = %+v", quotes)
	}
	if quotes[0].SecID != "SBER" || quotes[0].Last.String() != "300.5" || quotes[0].LastChangePercent.String() != "1.25" {
		t.Fatalf("sber = %+v last=%s change=%s", quotes[0], quotes[0].Last, quotes[0].LastChangePercent)
	}
	if quotes[1].SecID != "LKOH" || quotes[1].Last.String() != "7000" || quotes[1].LastChangePercent != nil {
		t.Fatalf("lkoh = %+v", quotes[1])
	}
}

func TestHistoryCandlesPagesAndPeriod(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 3, 15, 12, 0, 0, 0, moscow)
	query, err := PeriodMonth.HistoryQuery("SBER", now)
	if err != nil {
		t.Fatalf("HistoryQuery: %v", err)
	}
	if query.Interval != 60 || query.Instrument.Board != "" {
		t.Fatalf("query = %+v", query)
	}

	var starts []string
	var from, till, interval, path string
	client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		starts = append(starts, r.URL.Query().Get("start"))
		from = r.URL.Query().Get("from")
		till = r.URL.Query().Get("till")
		interval = r.URL.Query().Get("interval")
		path = r.URL.Path
		switch r.URL.Query().Get("start") {
		case "0":
			_, _ = io.WriteString(w, `{
				"candles": {"columns":["open","close","begin"],"data":[[101,102.5,"2026-03-10 12:00:00"]]}
			}`)
		case "1":
			_, _ = io.WriteString(w, `{
				"candles": {"columns":["begin","open","close"],"data":[["2026-03-01 09:30:00",9,8]]}
			}`)
		default:
			_, _ = io.WriteString(w, `{"candles":{"columns":["begin","open","close"],"data":[]}}`)
		}
	})

	rows, err := client.History(context.Background(), query)
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	if path != "/iss/engines/stock/markets/shares/securities/SBER/candles.json" {
		t.Fatalf("path = %s", path)
	}
	if from != "2026-02-08" || till != "2026-03-15" || interval != "60" {
		t.Fatalf("from=%s till=%s interval=%s", from, till, interval)
	}
	if join(starts) != "0,1,2" {
		t.Fatalf("starts = %v", starts)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %+v", rows)
	}
	if rows[0].Begin.Format(timeLayout) != "2026-03-01 09:30:00" || rows[0].Open.String() != "9" || rows[0].Close.String() != "8" {
		t.Fatalf("first = %+v %s %s", rows[0].Begin, rows[0].Open, rows[0].Close)
	}
	if rows[1].Open.String() != "101" || rows[1].Close.String() != "102.5" {
		t.Fatalf("second = %+v", rows[1])
	}
	if _, err := rows[1].Close.Float64(); err != nil {
		t.Fatal(err)
	}
}

func TestChartPeriodOffsets(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 3, 15, 12, 0, 0, 0, moscow)
	cases := []struct {
		period   ChartPeriod
		interval int
		from     string
	}{
		{PeriodDay, 1, "2026-03-13"},
		{PeriodWeek, 60, "2026-03-01"},
		{PeriodMonth, 60, "2026-02-08"},
		{PeriodHalfYear, 24, "2025-09-06"},
		{PeriodYear, 7, "2025-03-10"},
		{PeriodFiveYears, 7, "2020-12-31"},
	}
	for _, tc := range cases {
		t.Run(string(tc.period), func(t *testing.T) {
			query, err := tc.period.HistoryQuery("SBER", now)
			if err != nil {
				t.Fatal(err)
			}
			if query.Interval != tc.interval {
				t.Fatalf("interval = %d", query.Interval)
			}
			if query.From.In(moscow).Format(dateLayout) != tc.from {
				t.Fatalf("from = %s", query.From.In(moscow).Format(dateLayout))
			}
			if query.Till.In(moscow).Format(dateLayout) != "2026-03-15" {
				t.Fatalf("till = %s", query.Till.In(moscow).Format(dateLayout))
			}
			if query.Instrument.Engine != "stock" || query.Instrument.Market != "shares" || query.Instrument.Security != "SBER" {
				t.Fatalf("instrument = %+v", query.Instrument)
			}
		})
	}
}

func TestExternalErrors(t *testing.T) {
	t.Parallel()
	t.Run("status", func(t *testing.T) {
		client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "down", http.StatusBadGateway)
		})
		ctx := context.Background()
		ref := Instrument{Engine: "stock", Market: "shares"}
		errs := []error{
			call(client.Securities(ctx, SecuritiesQuery{Instrument: ref})),
			call(client.Quotes(ctx, ref)),
			call(client.History(ctx, sampleHistory())),
		}
		for _, err := range errs {
			var status *StatusError
			if !errors.Is(err, ErrStatus) || !errors.As(err, &status) || status.Code != http.StatusBadGateway {
				t.Fatalf("status = %v", err)
			}
		}
	})

	t.Run("parse", func(t *testing.T) {
		cases := []struct {
			name string
			body string
			call func(*Client) error
		}{
			{
				name: "broken json",
				body: `{`,
				call: func(c *Client) error {
					return call(c.Securities(context.Background(), SecuritiesQuery{Instrument: shares()}))
				},
			},
			{
				name: "missing block",
				body: `{"other":{"columns":["SECID"],"data":[]}}`,
				call: func(c *Client) error {
					return call(c.Quotes(context.Background(), shares()))
				},
			},
			{
				name: "missing column",
				body: `{"securities":{"columns":["SECID"],"data":[["SBER"]]}}`,
				call: func(c *Client) error {
					return call(c.Securities(context.Background(), SecuritiesQuery{Instrument: shares()}))
				},
			},
			{
				name: "wrong type",
				body: `{"marketdata":{"columns":["SECID","LAST","LASTCHANGEPRCNT"],"data":[["SBER","nope",1]]}}`,
				call: func(c *Client) error { return call(c.Quotes(context.Background(), shares())) },
			},
			{
				name: "short row",
				body: `{"marketdata":{"columns":["SECID","LAST","LASTCHANGEPRCNT"],"data":[["SBER"]]}}`,
				call: func(c *Client) error { return call(c.Quotes(context.Background(), shares())) },
			},
			{
				name: "null candle",
				body: `{"candles":{"columns":["begin","open","close"],"data":[["2026-03-01 10:00:00",null,1]]}}`,
				call: func(c *Client) error { return call(c.History(context.Background(), sampleHistory())) },
			},
						{
				name: "missing data",
				body: `{"securities":{"columns":["SECID","SHORTNAME"]}}`,
				call: func(c *Client) error {
					return call(c.Securities(context.Background(), SecuritiesQuery{Instrument: shares()}))
				},
			},
			{
				name: "null data",
				body: `{"securities":{"columns":["SECID","SHORTNAME"],"data":null}}`,
				call: func(c *Client) error {
					return call(c.Securities(context.Background(), SecuritiesQuery{Instrument: shares()}))
				},
			},
			{
				name: "data is not array",
				body: `{"securities":{"columns":["SECID","SHORTNAME"],"data":{}}}`,
				call: func(c *Client) error {
					return call(c.Securities(context.Background(), SecuritiesQuery{Instrument: shares()}))
				},
			},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
					_, _ = io.WriteString(w, tc.body)
				})
				err := tc.call(client)
				if !errors.Is(err, ErrParse) {
					t.Fatalf("err = %v", err)
				}
			})
		}
	})

	t.Run("network", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		base := server.URL
		server.Close()
		_, err := New(base).Securities(context.Background(), SecuritiesQuery{Instrument: shares()})
		if !errors.Is(err, ErrNetwork) {
			t.Fatalf("err = %v", err)
		}
		var status *StatusError
		if errors.As(err, &status) {
			t.Fatalf("network reported as status %d", status.Code)
		}
	})

	t.Run("cancel", func(t *testing.T) {
		started := make(chan struct{})
		client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
			close(started)
			<-r.Context().Done()
		})
		ctx, cancel := context.WithCancel(context.Background())
		errCh := make(chan error, 1)
		go func() {
			_, err := client.Quotes(ctx, shares())
			errCh <- err
		}()
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatal("request did not start")
		}
		cancel()
		err := <-errCh
		if !errors.Is(err, context.Canceled) || !errors.Is(err, ErrNetwork) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(400 * time.Millisecond)
			_, _ = io.WriteString(w, `{"securities":{"columns":["SECID","SHORTNAME"],"data":[]}}`)
		}, WithTimeout(40*time.Millisecond))
		_, err := client.Securities(context.Background(), SecuritiesQuery{Instrument: shares()})
		if !errors.Is(err, context.DeadlineExceeded) && !timedOut(err) {
			t.Fatalf("err = %v", err)
		}
		if !errors.Is(err, ErrNetwork) {
			t.Fatalf("timeout is not a network error: %v", err)
		}
	})
}

func TestBadArgumentsDoNotCallISS(t *testing.T) {
	t.Parallel()
	called := 0
	client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		called++
	})
	ctx := context.Background()
	cases := []error{
		call(client.Securities(ctx, SecuritiesQuery{Instrument: Instrument{Engine: "../stock", Market: "shares"}})),
		call(client.Securities(ctx, SecuritiesQuery{Instrument: Instrument{Engine: "stock", Market: "shares", Security: "SBER"}})),
		call(client.Quotes(ctx, Instrument{Engine: "stock", Market: "shares", Security: "SBER/P"})),
		call(client.History(ctx, HistoryQuery{})),
		call(client.History(nil, sampleHistory())),
		callErr(PeriodDay.HistoryQuery("", time.Now())),
		callErr(ChartPeriod("decade").HistoryQuery("SBER", time.Now())),
	}
	for i, err := range cases {
		if !errors.Is(err, ErrRequest) {
			t.Fatalf("case %d err = %v", i, err)
		}
	}
	if called != 0 {
		t.Fatalf("upstream calls = %d", called)
	}
}

func TestSamePageDoesNotLoop(t *testing.T) {
	t.Parallel()
	calls := 0
	client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = io.WriteString(w, `{
			"marketdata": {"columns":["SECID","LAST","LASTCHANGEPRCNT"],"data":[["SBER",293.99,-0.06]]}
		}`)
	})

	quotes, err := client.Quotes(context.Background(), Instrument{
		Engine: "stock", Market: "shares", Board: "TQBR", Security: "SBER",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 1 || quotes[0].Last.String() != "293.99" {
		t.Fatalf("quotes = %+v", quotes)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestEmptyHistory(t *testing.T) {
	t.Parallel()
	client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"candles":{"columns":["begin","open","close"],"data":[]}}`)
	})
	rows, err := client.History(context.Background(), sampleHistory())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestDefaultBaseURL(t *testing.T) {
	t.Parallel()
	client := New("")
	if client.baseURL != defaultBaseURL {
		t.Fatalf("base = %s", client.baseURL)
	}
	if client.timeout != defaultTimeout {
		t.Fatalf("timeout = %s", client.timeout)
	}
}

func newClient(t *testing.T, handler http.HandlerFunc, opts ...Option) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	opts = append([]Option{WithTimeout(time.Second)}, opts...)
	// Поздний WithTimeout заменяет дефолт теста, если он передан последним.
	return New(server.URL, opts...)
}

func shares() Instrument {
	return Instrument{Engine: "stock", Market: "shares"}
}

func sampleHistory() HistoryQuery {
	return HistoryQuery{
		Instrument: Instrument{Engine: "stock", Market: "shares", Security: "SBER"},
		From:       time.Date(2026, 2, 8, 0, 0, 0, 0, moscow),
		Till:       time.Date(2026, 3, 15, 0, 0, 0, 0, moscow),
		Interval:   60,
	}
}

func call[T any](rows []T, err error) error { return err }

func callErr(_ HistoryQuery, err error) error { return err }

func join(parts []string) string {
	out := ""
	for i, part := range parts {
		if i > 0 {
			out += ","
		}
		out += part
	}
	return out
}

func timedOut(err error) bool {
	var netErr interface{ Timeout() bool }
	return errors.As(err, &netErr) && netErr.Timeout()
}
