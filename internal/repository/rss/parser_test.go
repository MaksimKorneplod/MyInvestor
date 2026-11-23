package rss

import (
	"strings"
	"testing"
)

const testFeedXML = `<rss version="2.0">
	<channel>
		<item>
			<title>Экономические новости</title>
			<link>https://example.com/economy</link>
			<description>Описание экономической новости</description>
			<category>Экономика</category>
			<pubDate>Sat, 15 Nov 2026 10:00:00 GMT</pubDate>
			<guid>economy-1</guid>
		</item>
		<item>
			<title>Новости бизнеса</title>
			<link>https://example.com/business</link>
			<description>Описание новости бизнеса</description>
			<category>Бизнес</category>
			<pubDate>Sat, 15 Nov 2026 11:00:00 GMT</pubDate>
			<guid>business-1</guid>
		</item>
		<item>
			<title>Спортивные новости</title>
			<link>https://example.com/sport</link>
			<description>Описание спортивной новости</description>
			<category>Спорт</category>
			<pubDate>Sat, 15 Nov 2026 12:00:00 GMT</pubDate>
			<guid>sport-1</guid>
		</item>
	</channel>
</rss>`

func TestParseRss(t *testing.T) {
	feed, err := parseRss(strings.NewReader(testFeedXML))
	if err != nil {
		t.Fatalf("parseRss() error = %v", err)
	}

	if got, want := len(feed.Channel.Items), 3; got != want {
		t.Fatalf("got %d items, want %d", got, want)
	}

	got := feed.Channel.Items[0]
	if got.Title != "Экономические новости" ||
		got.Link != "https://example.com/economy" ||
		got.Description != "Описание экономической новости" ||
		got.Category != "Экономика" ||
		got.PubDate != "Sat, 15 Nov 2026 10:00:00 GMT" ||
		got.GUID != "economy-1" {
		t.Errorf("first parsed item = %+v, fields do not match XML", got)
	}
}

func TestParseRssInvalidXML(t *testing.T) {
	_, err := parseRss(strings.NewReader(`<rss><channel>`))
	if err == nil {
		t.Fatal("parseRss() error = nil, want malformed XML error")
	}
}

func TestParseRssRejectsNonRSSRoot(t *testing.T) {
	_, err := parseRss(strings.NewReader(`<html><body>upstream error</body></html>`))
	if err == nil {
		t.Fatal("parseRss() error = nil, want non-RSS root error")
	}
}

func TestParseRssRejectsMissingChannel(t *testing.T) {
	_, err := parseRss(strings.NewReader(`<rss version="2.0"></rss>`))
	if err == nil {
		t.Fatal("parseRss() error = nil, want missing channel error")
	}
}

func TestFilterRssFeedExtractsImage(t *testing.T) {
	xmlData := `<rss xmlns:media="http://search.yahoo.com/mrss/"><channel>
		<item>
			<title>Economy image</title>
			<link>https://example.com/economy-image</link>
			<description>&lt;img src="https://example.com/news.jpg" alt="News"&gt;</description>
			<category>Экономика</category>
			<pubDate>Sat, 15 Nov 2026 10:00:00 GMT</pubDate>
			<enclosure url="https://example.com/photo.jpg" type="image/jpeg"/>
		</item>
	</channel></rss>`

	feed, err := filterRssFeed(strings.NewReader(xmlData))
	if err != nil {
		t.Fatalf("filterRssFeed() error = %v", err)
	}
	if len(feed.Channel.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(feed.Channel.Items))
	}
	if got, want := feed.Channel.Items[0].Image, "https://example.com/photo.jpg"; got != want {
		t.Fatalf("image = %q, want %q", got, want)
	}
}

func TestFilterRssFeed(t *testing.T) {
	feed, err := filterRssFeed(strings.NewReader(testFeedXML))
	if err != nil {
		t.Fatalf("filterRssFeed() error = %v", err)
	}

	if got, want := len(feed.Channel.Items), 3; got != want {
		t.Fatalf("got %d filtered items, want %d", got, want)
	}
	for _, item := range feed.Channel.Items {
		if item.Category == "" {
			t.Errorf("missing category in filtered feed")
		}
	}
}

func TestFilterRssFeedSkipsInvalidItems(t *testing.T) {
	xmlData := `<rss><channel>
			<item><title></title><link>https://example.com/no-title</link><category>Экономика</category></item>
			<item><title>Bad link</title><link>/relative</link><category>Бизнес</category></item>
			<item><title>Bad date</title><link>https://example.com/bad-date</link><category>Бизнес</category><pubDate>not a date</pubDate></item>
			<item><title>Valid</title><link>https://example.com/valid</link><category>Бизнес</category></item>
			<item><title>Other category</title><link>https://example.com/other</link><category>Спорт</category></item>
		</channel></rss>`

	feed, err := filterRssFeed(strings.NewReader(xmlData))
	if err != nil {
		t.Fatalf("filterRssFeed() error = %v", err)
	}
	if got, want := len(feed.Channel.Items), 2; got != want {
		t.Fatalf("got %d items, want %d", got, want)
	}
	if got := feed.Channel.Items[0].Title; got != "Valid" {
		t.Errorf("item title = %q, want %q", got, "Valid")
	}
}
