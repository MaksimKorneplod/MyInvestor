package rss

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientRead(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("request method = %q, want %q", r.Method, http.MethodGet)
		}
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = io.WriteString(w, testFeedXML)
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL)
	feed, err := client.Read(context.Background())
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if got, want := len(feed.Channel.Items), 3; got != want {
		t.Fatalf("got %d filtered items, want %d", got, want)
	}
	for _, item := range feed.Channel.Items {
		if item.Category == "" {
			t.Errorf("missing category in response")
		}
	}
}

func TestClientReadUnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL)
	_, err := client.Read(context.Background())
	if !errors.Is(err, ErrSourceUnavailable) || !strings.Contains(err.Error(), "status 502") {
		t.Fatalf("Read() error = %v, want unexpected status code error", err)
	}
}

func TestClientReadInvalidXML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `<rss><channel>`)
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL)
	_, err := client.Read(context.Background())
	if !errors.Is(err, ErrInvalidFeed) || !strings.Contains(err.Error(), "parse RSS response") {
		t.Fatalf("Read() error = %v, want wrapped malformed XML error", err)
	}
}

func TestClientReadInvalidURL(t *testing.T) {
	client := NewClient(http.DefaultClient, "://invalid-url")
	if _, err := client.Read(context.Background()); err == nil || !strings.Contains(err.Error(), "create RSS request") {
		t.Fatalf("Read() error = %v, want wrapped invalid URL error", err)
	}
}

func TestClientReadInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		client *Client
		want   string
	}{
		{
			name:   "nil receiver",
			client: nil,
			want:   "client is nil",
		},
		{
			name:   "nil HTTP client",
			client: NewClient(nil, "https://example.com/feed"),
			want:   "HTTP client is nil",
		},
		{
			name:   "empty feed URL",
			client: NewClient(http.DefaultClient, " "),
			want:   "feed URL is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.client.Read(context.Background())
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Read() error = %v, want error containing %q", err, tt.want)
			}
		})
	}
}
