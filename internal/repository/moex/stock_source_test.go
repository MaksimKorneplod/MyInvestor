package moex

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MaksimKorneplod/MyInvestor/internal/service"
)

func TestStockAdapterListStocksMapsSecuritiesAndFiltersInvalidRows(t *testing.T) {
	client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/iss/engines/stock/markets/shares/boards/TQBR/securities.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"securities":{"columns":["SECID","SHORTNAME"],"data":[["SBER","Сбербанк"],["GAZP",""]]}}`)
	})

	instruments, err := NewStockAdapter(client).ListStocks(context.Background())
	if err != nil {
		t.Fatalf("ListStocks: %v", err)
	}
	want := []service.Instrument{
		{ID: "SBER", Ticker: "SBER", Name: "Сбербанк"},
		{ID: "GAZP", Ticker: "GAZP", Name: "GAZP"},
	}
	if len(instruments) != len(want) {
		t.Fatalf("instruments = %+v, want %+v", instruments, want)
	}
	for i := range want {
		if instruments[i] != want[i] {
			t.Errorf("instrument %d = %+v, want %+v", i, instruments[i], want[i])
		}
	}
}

func TestMapSecuritiesSkipsEmptySecIDAndUsesSecIDForEmptyName(t *testing.T) {
	instruments := mapSecurities([]Security{
		{SecID: "SBER", ShortName: "Сбербанк"},
		{SecID: ""},
		{SecID: "GAZP", ShortName: ""},
	})
	want := []service.Instrument{
		{ID: "SBER", Ticker: "SBER", Name: "Сбербанк"},
		{ID: "GAZP", Ticker: "GAZP", Name: "GAZP"},
	}
	if len(instruments) != len(want) {
		t.Fatalf("instruments = %+v, want %+v", instruments, want)
	}
	for i := range want {
		if instruments[i] != want[i] {
			t.Errorf("instrument %d = %+v, want %+v", i, instruments[i], want[i])
		}
	}
}

func TestStockAdapterListStocksEmptyResponseIsNonNil(t *testing.T) {
	client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"securities":{"columns":["SECID","SHORTNAME"],"data":[]}}`)
	})

	instruments, err := NewStockAdapter(client).ListStocks(context.Background())
	if err != nil {
		t.Fatalf("ListStocks: %v", err)
	}
	if instruments == nil {
		t.Fatal("instruments is nil, want an empty slice")
	}
	if len(instruments) != 0 {
		t.Fatalf("len(instruments) = %d, want 0", len(instruments))
	}
}

func TestStockAdapterListStocksMapsClientErrors(t *testing.T) {
	tests := []struct {
		name      string
		client    func(t *testing.T) *Client
		wantCause error
	}{
		{
			name: "status",
			client: func(t *testing.T) *Client {
				return newClient(t, func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadGateway)
				})
			},
			wantCause: ErrStatus,
		},
		{
			name: "network",
			client: func(t *testing.T) *Client {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
				baseURL := server.URL
				server.Close()
				return New(baseURL)
			},
			wantCause: ErrNetwork,
		},
		{
			name: "invalid JSON",
			client: func(t *testing.T) *Client {
				return newClient(t, func(w http.ResponseWriter, r *http.Request) {
					_, _ = io.WriteString(w, `{"securities":`)
				})
			},
			wantCause: ErrParse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewStockAdapter(tt.client(t)).ListStocks(context.Background())
			if !errors.Is(err, service.ErrSourceUnavailable) {
				t.Fatalf("err = %v, want ErrSourceUnavailable", err)
			}
			if !errors.Is(err, tt.wantCause) {
				t.Fatalf("err = %v, want underlying cause %v", err, tt.wantCause)
			}
		})
	}
}

func TestStockAdapterListStocksDoesNotLogContextCancellation(t *testing.T) {
	var logs strings.Builder
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	defer slog.SetDefault(previousLogger)

	started := make(chan struct{})
	client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	})

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		_, err := NewStockAdapter(client).ListStocks(ctx)
		errCh <- err
	}()
	<-started
	cancel()

	err := <-errCh
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if errors.Is(err, service.ErrSourceUnavailable) {
		t.Fatalf("err = %v, cancellation should remain distinguishable in adapter", err)
	}
	if logs.Len() != 0 {
		t.Fatalf("unexpected logs for context cancellation: %s", logs.String())
	}
}

func TestStockAdapterListStocksDoesNotMapDeadlineExceeded(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	_, err := NewStockAdapter(New("")).ListStocks(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
	if errors.Is(err, service.ErrSourceUnavailable) {
		t.Fatalf("err = %v, deadline should remain distinguishable in adapter", err)
	}
}
