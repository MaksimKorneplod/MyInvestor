package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockSource is a test stub for StockSource.
type mockSource struct {
	items     []Instrument
	err       error
	callCount int
}

func (m *mockSource) ListStocks(ctx context.Context) ([]Instrument, error) {
	m.callCount++
	if m.err != nil {
		return nil, m.err
	}
	return m.items, nil
}

func stockList() []Instrument {
	return []Instrument{
		{ID: "TEST-ID-1", Ticker: "SBER", Name: "Сбербанк"},
		{ID: "TEST-ID-2", Ticker: "SBERP", Name: "Сбербанк-п"},
		{ID: "TEST-ID-3", Ticker: "GAZP", Name: "Газпром"},
		{ID: "TEST-ID-4", Ticker: "LKOH", Name: "ЛУКОЙЛ"},
	}
}

func TestSearchByTickerLowerCase(t *testing.T) {
	c := NewCatalog(&mockSource{items: stockList()}, 30*time.Minute)
	page, err := c.Search(context.Background(), "sber", 10, 0)
	if err != nil {
		t.Fatalf("search by ticker lowercase: %v", err)
	}
	if len(page.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(page.Items))
	}
	for _, item := range page.Items {
		if item.Ticker != "SBER" && item.Ticker != "SBERP" {
			t.Errorf("unexpected ticker: %s", item.Ticker)
		}
	}
}

func TestSearchByNameVariousCases(t *testing.T) {
	c := NewCatalog(&mockSource{items: stockList()}, 30*time.Minute)
	for _, query := range []string{"сбер", "СБЕР", "Сбер"} {
		page, err := c.Search(context.Background(), query, 10, 0)
		if err != nil {
			t.Fatalf("search by name %q: %v", query, err)
		}
		if len(page.Items) != 2 {
			t.Errorf("expected 2 items for query %q, got %d", query, len(page.Items))
		}
	}
}

func TestSearchByNameNotMatchingTicker(t *testing.T) {
	c := NewCatalog(&mockSource{items: stockList()}, 30*time.Minute)
	page, err := c.Search(context.Background(), "газ", 10, 0)
	if err != nil {
		t.Fatalf("search by name gaz: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].Ticker != "GAZP" {
		t.Errorf("expected GAZP, got %v", page.Items)
	}
}

func TestSearchEmptyQueryReturnsAllSorted(t *testing.T) {
	c := NewCatalog(&mockSource{items: stockList()}, 30*time.Minute)
	page, err := c.Search(context.Background(), "", 10, 0)
	if err != nil {
		t.Fatalf("empty query: %v", err)
	}
	if page.Total != 4 {
		t.Errorf("expected total 4, got %d", page.Total)
	}
	if len(page.Items) != 4 {
		t.Errorf("expected 4 items, got %d", len(page.Items))
	}
	expected := []string{"GAZP", "LKOH", "SBER", "SBERP"}
	for i, ticker := range expected {
		if page.Items[i].Ticker != ticker {
			t.Errorf("item %d: expected ticker %s, got %s", i, ticker, page.Items[i].Ticker)
		}
	}
}

func TestSearchNoMatchesEmptyItems(t *testing.T) {
	c := NewCatalog(&mockSource{items: stockList()}, 30*time.Minute)
	page, err := c.Search(context.Background(), "XYZ", 10, 0)
	if err != nil {
		t.Fatalf("no matches: %v", err)
	}
	if page.Items == nil {
		t.Error("items should not be nil for no matches")
	}
	if page.Total != 0 {
		t.Errorf("expected total 0, got %d", page.Total)
	}
}

func TestPagination(t *testing.T) {
	c := NewCatalog(&mockSource{items: stockList()}, 30*time.Minute)
	page, err := c.Search(context.Background(), "", 2, 1)
	if err != nil {
		t.Fatalf("pagination: %v", err)
	}
	if len(page.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(page.Items))
	}
	if page.Total != 4 {
		t.Errorf("expected total 4, got %d", page.Total)
	}
	if page.Items[0].Ticker != "LKOH" {
		t.Errorf("expected first item LKOH, got %s", page.Items[0].Ticker)
	}
}

func TestPaginationOffsetBeyondList(t *testing.T) {
	c := NewCatalog(&mockSource{items: stockList()}, 30*time.Minute)
	page, err := c.Search(context.Background(), "", 10, 100)
	if err != nil {
		t.Fatalf("pagination offset beyond: %v", err)
	}
	if len(page.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(page.Items))
	}
	if page.Total != 4 {
		t.Errorf("expected total 4, got %d", page.Total)
	}
}

func TestSourceErrorReturnsErrSourceUnavailable(t *testing.T) {
	c := NewCatalog(&mockSource{err: errors.New("down")}, 30*time.Minute)
	page, err := c.Search(context.Background(), "", 10, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrSourceUnavailable) {
		t.Errorf("expected ErrSourceUnavailable, got %v", err)
	}
	if page.Items != nil && len(page.Items) == 0 {
		t.Error("on error items should be empty")
	}
}

func TestSourceErrorNotCached(t *testing.T) {
	ms := &mockSource{err: errors.New("down")}
	c := NewCatalog(ms, 30*time.Minute)
	_, err1 := c.Search(context.Background(), "", 10, 0)
	if err1 == nil {
		t.Fatal("expected first error")
	}
	_, err2 := c.Search(context.Background(), "", 10, 0)
	if err2 == nil {
		t.Fatal("expected second error")
	}
	if ms.callCount != 2 {
		t.Errorf("expected source called twice, got %d", ms.callCount)
	}
}

func TestCacheHitsSourceOnce(t *testing.T) {
	ms := &mockSource{items: stockList()}
	c := NewCatalog(ms, 30*time.Minute)
	_, err1 := c.Search(context.Background(), "", 10, 0)
	if err1 != nil {
		t.Fatalf("first search: %v", err1)
	}
	_, err2 := c.Search(context.Background(), "", 10, 0)
	if err2 != nil {
		t.Fatalf("second search: %v", err2)
	}
	if ms.callCount != 1 {
		t.Errorf("expected source called once, got %d", ms.callCount)
	}
}

func TestCacheExpiredCallsSourceAgain(t *testing.T) {
	called := 0
	source := &mockSource{items: stockList()}
	now := time.Now()
	c := NewCatalog(source, time.Minute)
	c.now = func() time.Time { return now }
	_, err := c.Search(context.Background(), "", 10, 0)
	if err != nil {
		t.Fatalf("first search: %v", err)
	}
	called = source.callCount
	c.now = func() time.Time { return now.Add(time.Minute * 2) }
	_, err = c.Search(context.Background(), "", 10, 0)
	if err != nil {
		t.Fatalf("second search after ttl: %v", err)
	}
	if source.callCount != called+1 {
		t.Errorf("expected source called again after TTL, count %d", source.callCount)
	}
}

func TestReturnedSliceModificationDoesNotAffectCache(t *testing.T) {
	c := NewCatalog(&mockSource{items: stockList()}, 30*time.Minute)
	page1, err := c.Search(context.Background(), "", 10, 0)
	if err != nil {
		t.Fatalf("first search: %v", err)
	}
	if len(page1.Items) > 0 {
		page1.Items[0].Ticker = "MODIFIED"
	}
	page2, err := c.Search(context.Background(), "", 10, 0)
	if err != nil {
		t.Fatalf("second search: %v", err)
	}
	if page2.Items[0].Ticker == "MODIFIED" {
		t.Error("modifying returned slice affected cache")
	}
}
