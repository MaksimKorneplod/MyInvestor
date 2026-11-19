package rss

import (
	"encoding/xml"
	"errors"
	"io"
	"log/slog"
	"net/url"
	"strings"
)

func parseRss(r io.Reader) (RssFeed, error) {
	var feed RssFeed
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(&feed); err != nil {
		return RssFeed{}, err
	}
	return feed, nil
}

func filterRssFeed(r io.Reader) (RssFeed, error) {
	var feed RssFeed
	parsedData, err := parseRss(r)
	if err != nil {
		return RssFeed{}, err
	}

	for _, item := range parsedData.Channel.Items {
		item.Title = strings.TrimSpace(item.Title)
		item.Link = strings.TrimSpace(item.Link)
		item.Category = strings.TrimSpace(item.Category)
		item.PubDate = strings.TrimSpace(item.PubDate)
		if err := validateItem(item); err != nil {
			slog.Warn("skipping invalid RSS item", "guid", item.GUID, "title", item.Title, "error", err)
			continue
		}
		if item.Category != "Экономика" && item.Category != "Бизнес" {
			continue
		}
		feed.Channel.Items = append(feed.Channel.Items, item)
	}
	return feed, nil
}

func validateItem(item RssItem) error {
	if strings.TrimSpace(item.Title) == "" {
		return errors.New("missing title")
	}
	if strings.TrimSpace(item.Category) == "" {
		return errors.New("missing category")
	}

	link, err := url.Parse(strings.TrimSpace(item.Link))
	if err != nil || (link.Scheme != "http" && link.Scheme != "https") || link.Host == "" {
		return errors.New("invalid or missing link")
	}
	if _, err := ParsePubDate(item.PubDate); err != nil {
		return err
	}
	return nil
}
