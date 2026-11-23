package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/MaksimKorneplod/MyInvestor/internal/repository/rss"
)

type NewsReader interface {
	Read(ctx context.Context) (rss.RssFeed, error)
}

const (
	defaultNewsLimit = 50
	maxNewsLimit     = 100
)

type newsItemResponse struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description,omitempty"`
	PubDate     string `json:"pubDate,omitempty"`
	Category    string `json:"category,omitempty"`
	Image       string `json:"image,omitempty"`
}

type newsResponse struct {
	Items  []newsItemResponse `json:"items"`
	Total  int                `json:"total"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}

func NewsHandler(reader NewsReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset, err := newsPagination(r)
		if err != nil {
			writeNewsError(w, http.StatusBadRequest, "invalid pagination parameters")
			return
		}

		if reader == nil {
			slog.Error("news reader is nil")
			writeNewsError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		feed, err := reader.Read(r.Context())
		if err != nil {
			if r.Context().Err() != nil {
				return
			}
			if errors.Is(err, rss.ErrSourceUnavailable) {
				slog.Error("failed to read news source", "error", err)
				writeNewsError(w, http.StatusServiceUnavailable, "news source is temporarily unavailable")
				return
			}
			if errors.Is(err, rss.ErrInvalidFeed) {
				slog.Error("failed to parse news source response", "error", err)
				writeNewsError(w, http.StatusBadGateway, "news source returned invalid data")
				return
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			slog.Error("unexpected news handler error", "error", err)
			writeNewsError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		var feedItems []rss.RssItem
		if feed.Channel != nil {
			feedItems = feed.Channel.Items
		}
		items := orderedNews(feedItems)
		total := len(items)
		start := min(offset, total)
		end := start + min(limit, total-start)
		page := make([]newsItemResponse, 0, end-start)
		for _, item := range items[start:end] {
			page = append(page, newsItemResponse{
				Title:       item.Title,
				Link:        item.Link,
				Description: item.Description,
				PubDate:     item.PubDate,
				Category:    item.Category,
				Image:       item.Image,
			})
		}
		writeNewsJSON(w, http.StatusOK, newsResponse{
			Items:  page,
			Total:  total,
			Limit:  limit,
			Offset: offset,
		})
	}
}

func newsPagination(r *http.Request) (int, int, error) {
	limit := defaultNewsLimit
	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > maxNewsLimit {
			return 0, 0, fmt.Errorf("limit must be between 1 and %d", maxNewsLimit)
		}
		limit = parsed
	}

	offset := 0
	if value := r.URL.Query().Get("offset"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			return 0, 0, fmt.Errorf("offset must be a non-negative integer")
		}
		offset = parsed
	}
	return limit, offset, nil
}

func orderedNews(items []rss.RssItem) []rss.RssItem {
	type datedItem struct {
		item rss.RssItem
		date time.Time
		has  bool
	}

	ordered := make([]datedItem, 0, len(items))
	for _, item := range items {
		date, err := rss.ParsePubDate(item.PubDate)
		if err != nil {
			slog.Warn("skipping news item with invalid publication date", "guid", item.GUID, "error", err)
			continue
		}
		ordered = append(ordered, datedItem{item: item, date: date, has: !date.IsZero()})
	}

	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].has != ordered[j].has {
			return ordered[i].has
		}
		if !ordered[i].has {
			return false
		}
		return ordered[i].date.After(ordered[j].date)
	})

	result := make([]rss.RssItem, len(ordered))
	for i := range ordered {
		result[i] = ordered[i].item
	}
	return result
}

func writeNewsJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("failed to encode news response", "error", err)
	}
}

func writeNewsError(w http.ResponseWriter, status int, message string) {
	writeNewsJSON(w, status, map[string]string{"error": strings.TrimSpace(message)})
}
