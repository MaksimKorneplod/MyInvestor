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

func TestStockListerListTQBRAndReturnsAllPages(t *testing.T) {
	var starts []string
	var path string
	client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		starts = append(starts, r.URL.Query().Get("start"))
		path = r.URL.Path
		if r.URL.Query().Get("start") == "0" {
			_, _ = io.WriteString(w, `{
				"securities": {"columns":["SECID","SHORTNAME"],"data":[["SBER","Сбербанк"]]},
				"securities.cursor": {"columns":["INDEX","TOTAL","PAGESIZE"],"data":[[0,2,1]]}
			}`)
			return
		}
		_, _ = io.WriteString(w, `{
			"securities": {"columns":["SECID","SHORTNAME"],"data":[["GAZP","Газпром"]]},
			"securities.cursor": {"columns":["INDEX","TOTAL","PAGESIZE"],"data":[[1,2,1]]}
		}`)
	})

	securities, err := NewStockLister(client).List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if path != "/iss/engines/stock/markets/shares/boards/TQBR/securities.json" {
		t.Errorf("path = %q", path)
	}
	if got := join(starts); got != "0,1" {
		t.Errorf("starts = %q, want %q", got, "0,1")
	}
	want := []Security{
		{SecID: "SBER", ShortName: "Сбербанк"},
		{SecID: "GAZP", ShortName: "Газпром"},
	}
	if len(securities) != len(want) {
		t.Fatalf("securities = %+v, want %+v", securities, want)
	}
	for i := range want {
		if securities[i] != want[i] {
			t.Errorf("security %d = %+v, want %+v", i, securities[i], want[i])
		}
	}
}

func TestStockListerListReturnsClientErrors(t *testing.T) {
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
		{
			name: "timeout",
			client: func(t *testing.T) *Client {
				return newClient(t, func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(100 * time.Millisecond)
					_, _ = io.WriteString(w, `{"securities":{"columns":["SECID","SHORTNAME"],"data":[]}}`)
				}, WithTimeout(20*time.Millisecond))
			},
			wantCause: ErrNetwork,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewStockLister(tt.client(t)).List(context.Background())
			if err == nil {
				t.Fatal("List returned nil error")
			}
			if !errors.Is(err, tt.wantCause) {
				t.Errorf("err = %v, want cause %v", err, tt.wantCause)
			}
			if tt.name == "timeout" && !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("timeout error = %v, want context.DeadlineExceeded", err)
			}
		})
	}
}

func TestStockListerListNilClient(t *testing.T) {
	_, err := NewStockLister(nil).List(context.Background())
	if !errors.Is(err, ErrRequest) {
		t.Fatalf("err = %v, want ErrRequest", err)
	}
}
