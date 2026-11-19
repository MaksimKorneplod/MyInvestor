package rss

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

// Распарщеная новость
type RssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Category    string `xml:"category"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

// Слайс распарщеных новостей
type RssChannel struct {
	Items []RssItem `xml:"item"`
}

// Лента новостей (могут быть разные RSS-каналы)
type RssFeed struct {
	XMLName xml.Name    `xml:"rss"`
	Channel *RssChannel `xml:"channel"`
}

func ParsePubDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}

	for _, layout := range []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC3339,
		time.RFC3339Nano,
	} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid pubDate %q", value)
}
