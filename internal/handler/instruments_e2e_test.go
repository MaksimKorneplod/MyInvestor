package handler_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/MaksimKorneplod/MyInvestor/internal/handler"
	"github.com/MaksimKorneplod/MyInvestor/internal/repository/moex"
	"github.com/MaksimKorneplod/MyInvestor/internal/service"
)

type instrumentResponse struct {
	Items  []service.Instrument `json:"items"`
	Total  int                  `json:"total"`
	Limit  int                  `json:"limit"`
	Offset int                  `json:"offset"`
	Error  string               `json:"error"`
}

func newInstrumentEndpoint(t *testing.T, timeout time.Duration, issHandler http.HandlerFunc) http.Handler {
	t.Helper()

	iss := httptest.NewServer(issHandler)
	t.Cleanup(iss.Close)

	client := moex.New(iss.URL, moex.WithTimeout(timeout))
	lister := moex.NewStockLister(client)
	source := service.NewMOEXStockSource(lister)
	catalog := service.NewCatalog(source, 30*time.Minute)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/instruments", handler.InstrumentsHandler(catalog))
	return mux
}

func writeSecuritiesResponse(w http.ResponseWriter) {
	_, _ = io.WriteString(w, `{
		"securities": {
			"columns": ["SECID", "SHORTNAME"],
			"data": [["SBER", "Сбербанк"], ["SBERP", "Сбербанк-п"], ["GAZP", "Газпром"]]
		},
		"securities.cursor": {
			"columns": ["INDEX", "TOTAL", "PAGESIZE"],
			"data": [[0, 3, 3]]
		}
	}`)
}

func TestInstrumentCatalogEndToEnd(t *testing.T) {
	endpoint := newInstrumentEndpoint(t, time.Second, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/iss/engines/stock/markets/shares/boards/TQBR/securities.json" {
			t.Errorf("MOEX path = %q", r.URL.Path)
		}
		writeSecuritiesResponse(w)
	})

	cases := []struct {
		name       string
		query      url.Values
		wantItems  []service.Instrument
		wantTotal  int
		wantLimit  int
		wantOffset int
	}{
		{
			name:  "all instruments with fields",
			query: url.Values{},
			wantItems: []service.Instrument{
				{ID: "GAZP", Ticker: "GAZP", Name: "Газпром"},
				{ID: "SBER", Ticker: "SBER", Name: "Сбербанк"},
				{ID: "SBERP", Ticker: "SBERP", Name: "Сбербанк-п"},
			},
			wantTotal: 3,
			wantLimit: 50,
		},
		{
			name:  "case-insensitive Cyrillic search",
			query: url.Values{"q": []string{"СБеР"}},
			wantItems: []service.Instrument{
				{ID: "SBER", Ticker: "SBER", Name: "Сбербанк"},
				{ID: "SBERP", Ticker: "SBERP", Name: "Сбербанк-п"},
			},
			wantTotal: 2,
			wantLimit: 50,
		},
		{
			name:  "limit and offset",
			query: url.Values{"limit": []string{"1"}, "offset": []string{"1"}},
			wantItems: []service.Instrument{
				{ID: "SBER", Ticker: "SBER", Name: "Сбербанк"},
			},
			wantTotal:  3,
			wantLimit:  1,
			wantOffset: 1,
		},
		{
			name:      "no matches",
			query:     url.Values{"q": []string{"qqqq"}},
			wantItems: []service.Instrument{},
			wantTotal: 0,
			wantLimit: 50,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/instruments"
			if encoded := tt.query.Encode(); encoded != "" {
				path += "?" + encoded
			}
			recorder := httptest.NewRecorder()
			endpoint.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))

			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d; body: %s", recorder.Code, http.StatusOK, recorder.Body.String())
			}
			var got instrumentResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if got.Items == nil {
				t.Fatal("items is null, want an array")
			}
			if len(got.Items) != len(tt.wantItems) {
				t.Fatalf("items = %+v, want %+v", got.Items, tt.wantItems)
			}
			for i := range tt.wantItems {
				if got.Items[i] != tt.wantItems[i] {
					t.Errorf("item %d = %+v, want %+v", i, got.Items[i], tt.wantItems[i])
				}
			}
			if got.Total != tt.wantTotal || got.Limit != tt.wantLimit || got.Offset != tt.wantOffset {
				t.Errorf("pagination = total:%d limit:%d offset:%d, want total:%d limit:%d offset:%d",
					got.Total, got.Limit, got.Offset, tt.wantTotal, tt.wantLimit, tt.wantOffset)
			}
		})
	}
}

func TestInstrumentCatalogEndToEndMOEXFailure(t *testing.T) {
	endpoint := newInstrumentEndpoint(t, time.Second, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "private upstream details", http.StatusInternalServerError)
	})

	recorder := httptest.NewRecorder()
	endpoint.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/instruments", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d; body: %s", recorder.Code, http.StatusServiceUnavailable, recorder.Body.String())
	}
	var got instrumentResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Error != "market data source is temporarily unavailable, please try again later" {
		t.Errorf("error = %q, want generic unavailable message", got.Error)
	}
	if strings.Contains(recorder.Body.String(), "private upstream details") {
		t.Errorf("response leaked internal details: %s", recorder.Body.String())
	}
}

func TestInstrumentCatalogEndToEndMOEXTimeout(t *testing.T) {
	endpoint := newInstrumentEndpoint(t, 20*time.Millisecond, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		writeSecuritiesResponse(w)
	})

	recorder := httptest.NewRecorder()
	endpoint.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/instruments", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d; body: %s", recorder.Code, http.StatusServiceUnavailable, recorder.Body.String())
	}
	var got instrumentResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Error != "market data source is temporarily unavailable, please try again later" {
		t.Errorf("error = %q, want generic unavailable message", got.Error)
	}
	if strings.Contains(recorder.Body.String(), "deadline exceeded") {
		t.Errorf("response leaked timeout details: %s", recorder.Body.String())
	}
	if got.Error == "" || recorder.Code == http.StatusInternalServerError {
		t.Errorf("expected 503 unavailable response, got status %d and body %s", recorder.Code, recorder.Body.String())
	}
}
