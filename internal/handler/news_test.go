package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MaksimKorneplod/MyInvestor/internal/repository/rss"
)

type newsReaderStub struct {
	feed rss.RssFeed
	err  error
}

func (s newsReaderStub) Read(context.Context) (rss.RssFeed, error) {
	return s.feed, s.err
}

func TestNewsHandlerOrdersAndPaginates(t *testing.T) {
	reader := newsReaderStub{feed: rss.RssFeed{Channel: rss.RssChannel{Items: []rss.RssItem{
		{Title: "Old", Link: "https://example.com/old", PubDate: "Mon, 02 Jan 2006 10:00:00 GMT"},
		{Title: "New", Link: "https://example.com/new", Description: "Latest", PubDate: "Wed, 04 Jan 2006 10:00:00 GMT"},
		{Title: "Undated", Link: "https://example.com/undated"},
	}}}}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/news?limit=1&offset=1", nil)
	rec := httptest.NewRecorder()

	NewsHandler(reader)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got newsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Total != 3 || got.Limit != 1 || got.Offset != 1 {
		t.Fatalf("pagination = %+v", got)
	}
	if len(got.Items) != 1 || got.Items[0].Title != "Old" {
		t.Fatalf("items = %+v, want the second newest item", got.Items)
	}
}

func TestNewsHandlerDefaultPaginationAndEmptyFeed(t *testing.T) {
	rec := httptest.NewRecorder()
	NewsHandler(newsReaderStub{})(rec, httptest.NewRequest(http.MethodGet, "/api/v1/news", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got newsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Items == nil || len(got.Items) != 0 || got.Limit != defaultNewsLimit || got.Offset != 0 {
		t.Fatalf("response = %+v, want empty array and default pagination", got)
	}
}

func TestNewsHandlerReturnsNewsFields(t *testing.T) {
	reader := newsReaderStub{feed: rss.RssFeed{Channel: rss.RssChannel{Items: []rss.RssItem{
		{
			Title:       "Latest",
			Link:        "https://example.com/latest",
			Description: "News summary",
			PubDate:     "Wed, 04 Jan 2006 10:00:00 GMT",
		},
	}}}}
	rec := httptest.NewRecorder()
	NewsHandler(reader)(rec, httptest.NewRequest(http.MethodGet, "/api/v1/news?limit=1", nil))

	var got newsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if rec.Code != http.StatusOK || len(got.Items) != 1 {
		t.Fatalf("status = %d, response = %+v", rec.Code, got)
	}
	item := got.Items[0]
	if item.Title != "Latest" || item.Link != "https://example.com/latest" ||
		item.Description != "News summary" || item.PubDate == "" {
		t.Errorf("news item fields = %+v", item)
	}
}

func TestNewsHandlerRejectsInvalidPagination(t *testing.T) {
	for _, path := range []string{
		"/api/v1/news?limit=0",
		"/api/v1/news?limit=101",
		"/api/v1/news?limit=abc",
		"/api/v1/news?offset=-1",
		"/api/v1/news?offset=abc",
	} {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			NewsHandler(newsReaderStub{})(rec, httptest.NewRequest(http.MethodGet, path, nil))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestNewsHandlerMapsSourceErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "upstream unavailable", err: rss.ErrSourceUnavailable, want: http.StatusServiceUnavailable},
		{name: "invalid feed", err: rss.ErrInvalidFeed, want: http.StatusBadGateway},
		{name: "unexpected error", err: errors.New("internal"), want: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			NewsHandler(newsReaderStub{err: tt.err})(rec, httptest.NewRequest(http.MethodGet, "/api/v1/news", nil))
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}
