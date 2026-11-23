package rss

import (
	"encoding/xml"
	"errors"
	"io"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
)

var descriptionImageRe = regexp.MustCompile(`(?i)<img[^>]+src\s*=\s*["']([^"']+)["']`)

func parseRss(r io.Reader) (RssFeed, error) {
	var feed RssFeed
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(&feed); err != nil {
		return RssFeed{}, err
	}
	if feed.Channel == nil {
		return RssFeed{}, errors.New("RSS document is missing channel")
	}
	return feed, nil
}

func filterRssFeed(r io.Reader) (RssFeed, error) {
	feed := RssFeed{Channel: &RssChannel{Items: make([]RssItem, 0)}}
	parsedData, err := parseRss(r)
	if err != nil {
		return RssFeed{}, err
	}

	for _, item := range parsedData.Channel.Items {
		item.Title = strings.TrimSpace(item.Title)
		item.Link = strings.TrimSpace(item.Link)
		item.Category = strings.TrimSpace(item.Category)
		item.PubDate = strings.TrimSpace(item.PubDate)
		item.Image = extractItemImage(item)
		if err := validateItem(item); err != nil {
			slog.Warn("skipping invalid RSS item", "guid", item.GUID, "title", item.Title, "error", err)
			continue
		}
		// if item.Category != "Экономика" && item.Category != "Бизнес" {
		// 	continue
		// }
		feed.Channel.Items = append(feed.Channel.Items, item)
	}
	return feed, nil
}

func extractItemImage(item RssItem) string {
	if item.Enclosure != nil &&
		(item.Enclosure.Type == "" || strings.HasPrefix(strings.ToLower(item.Enclosure.Type), "image/")) &&
		validImageURL(item.Enclosure.URL) {
		return strings.TrimSpace(item.Enclosure.URL)
	}
	for _, media := range item.MediaContent {
		if (media.Medium == "" || strings.EqualFold(media.Medium, "image")) &&
			(media.Type == "" || strings.HasPrefix(strings.ToLower(media.Type), "image/")) &&
			validImageURL(media.URL) {
			return strings.TrimSpace(media.URL)
		}
	}
	for _, media := range item.MediaThumbnail {
		if validImageURL(media.URL) {
			return strings.TrimSpace(media.URL)
		}
	}
	if match := descriptionImageRe.FindStringSubmatch(item.Description); len(match) == 2 && validImageURL(match[1]) {
		return strings.TrimSpace(match[1])
	}
	return ""
}

func validImageURL(value string) bool {
	parsed, err := url.Parse(strings.TrimSpace(value))
	return err == nil &&
		(parsed.Scheme == "http" || parsed.Scheme == "https") &&
		parsed.Host != ""
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
