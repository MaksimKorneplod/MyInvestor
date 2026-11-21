package rss

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const DefaultFeedURL = "https://www.interfax.ru/rss"

var (
	ErrSourceUnavailable = errors.New("RSS source unavailable")
	ErrInvalidFeed       = errors.New("invalid RSS feed")
)

type Client struct {
	httpClient *http.Client
	feedURL    string
}

func NewClient(httpClient *http.Client, feedURL string) *Client {
	return &Client{
		httpClient: httpClient,
		feedURL:    feedURL,
	}
}

func (c *Client) Read(ctx context.Context) (RssFeed, error) {
	if c == nil {
		return RssFeed{}, fmt.Errorf("read RSS feed: client is nil")
	}
	if c.httpClient == nil {
		return RssFeed{}, fmt.Errorf("read RSS feed: HTTP client is nil")
	}
	if strings.TrimSpace(c.feedURL) == "" {
		return RssFeed{}, fmt.Errorf("read RSS feed: feed URL is empty")
	}
	if ctx == nil {
		return RssFeed{}, fmt.Errorf("read RSS feed: context is nil")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.feedURL, nil)
	if err != nil {
		return RssFeed{}, fmt.Errorf("create RSS request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return RssFeed{}, fmt.Errorf("%w: send RSS request: %w", ErrSourceUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return RssFeed{}, fmt.Errorf("%w: unexpected HTTP status %d", ErrSourceUnavailable, resp.StatusCode)
	}

	feed, err := filterRssFeed(resp.Body)
	if err != nil {
		return RssFeed{}, fmt.Errorf("%w: parse RSS response: %w", ErrInvalidFeed, err)
	}
	return feed, nil
}
