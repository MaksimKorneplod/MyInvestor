package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/MaksimKorneplod/MyInvestor/internal/repository/moex"
)

// mockSource is a test stub for StockSource.
type mockSource struct {
	items     []Instrument
	err       error
	callCount int
	mu        sync.Mutex
	started   chan struct{}
	release   <-chan struct{}
}

func (m *mockSource) ListStocks(ctx context.Context) ([]Instrument, error) {
	m.mu.Lock()
	m.callCount++
	items, err := m.items, m.err
	started, release := m.started, m.release
	m.mu.Unlock()

	if started != nil {
		select {
		case started <- struct{}{}:
		default:
		}
	}
	if release != nil {
		<-release
	}
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (m *mockSource) calls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

func stockList() []Instrument {
	return []Instrument{
		{ID: "TEST-ID-1", Ticker: "SBER", Name: "Сбербанк"},
		{ID: "TEST-ID-2", Ticker: "SBERP", Name: "Сбербанк-п"},
		{ID: "TEST-ID-3", Ticker: "GAZP", Name: "Газпром"},
		{ID: "TEST-ID-4", Ticker: "LKOH", Name: "ЛУКОЙЛ"},
	}
}

type mockMOEXLister struct {
	securities []moex.Security
	err        error
}

func (m mockMOEXLister) List(context.Context) ([]moex.Security, error) {
	return m.securities, m.err
}

func TestMOEXStockSourceMapsSecurities(t *testing.T) {
	source := NewMOEXStockSource(mockMOEXLister{
		securities: []moex.Security{
			{SecID: "SBER", ShortName: "Сбербанк"},
			{SecID: ""},
			{SecID: "GAZP"},
		},
	})

	instruments, err := source.ListStocks(context.Background())
	if err != nil {
		t.Fatalf("ListStocks: %v", err)
	}
	want := []Instrument{
		{ID: "SBER", Ticker: "SBER", Name: "Сбербанк"},
		{ID: "GAZP", Ticker: "GAZP", Name: "GAZP"},
	}
	if len(instruments) != len(want) {
		t.Fatalf("instruments = %+v, want %+v", instruments, want)
	}
	for i := range want {
		if instruments[i] != want[i] {
			t.Errorf("instrument %d = %+v, want %+v", i, instruments[i], want[i])
		}
	}
}

func TestMOEXStockSourceReturnsListerError(t *testing.T) {
	sourceErr := errors.New("MOEX unavailable")
	source := NewMOEXStockSource(mockMOEXLister{err: sourceErr})

	_, err := source.ListStocks(context.Background())
	if err != sourceErr {
		t.Fatalf("err = %v, want original error %v", err, sourceErr)
	}
	if !errors.Is(err, sourceErr) {
		t.Fatalf("errors.Is(%v, sourceErr) = false", err)
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
	sourceErr := errors.New("down")
	c := NewCatalog(&mockSource{err: sourceErr}, 30*time.Minute)
	page, err := c.Search(context.Background(), "", 10, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrSourceUnavailable) {
		t.Errorf("expected ErrSourceUnavailable, got %v", err)
	}
	if !errors.Is(err, sourceErr) {
		t.Errorf("expected source error to be preserved, got %v", err)
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
	if ms.calls() != 2 {
		t.Errorf("expected source called twice, got %d", ms.calls())
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
	if ms.calls() != 1 {
		t.Errorf("expected source called once, got %d", ms.calls())
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
	called = source.calls()
	c.now = func() time.Time { return now.Add(time.Minute * 2) }
	_, err = c.Search(context.Background(), "", 10, 0)
	if err != nil {
		t.Fatalf("second search after ttl: %v", err)
	}
	if source.calls() != called+1 {
		t.Errorf("expected source called again after TTL, count %d", source.calls())
	}
}

func TestEmptyListIsCached(t *testing.T) {
	source := &mockSource{}
	c := NewCatalog(source, 30*time.Minute)

	for i := 0; i < 2; i++ {
		page, err := c.Search(context.Background(), "", 10, 0)
		if err != nil {
			t.Fatalf("Search %d: %v", i+1, err)
		}
		if page.Items == nil || page.Total != 0 {
			t.Errorf("Search %d returned page %+v, want non-nil empty items and total 0", i+1, page)
		}
	}
	if got := source.calls(); got != 1 {
		t.Fatalf("source calls = %d, want 1", got)
	}
}

type observedContext struct {
	context.Context
	entered chan<- struct{}
	once    sync.Once
}

func (c *observedContext) Done() <-chan struct{} {
	c.once.Do(func() { c.entered <- struct{}{} })
	return c.Context.Done()
}

func TestParallelSearchesShareOneLoad(t *testing.T) {
	const searchCount = 20
	started := make(chan struct{}, searchCount)
	release := make(chan struct{})
	source := &mockSource{
		items:   stockList(),
		started: started,
		release: release,
	}
	catalog := NewCatalog(source, 30*time.Minute)
	entered := make(chan struct{}, searchCount)
	errs := make(chan error, searchCount)

	for i := 0; i < searchCount; i++ {
		ctx := &observedContext{Context: context.Background(), entered: entered}
		go func(ctx context.Context) {
			_, err := catalog.Search(ctx, "", 10, 0)
			errs <- err
		}(ctx)
	}

	<-started
	for i := 0; i < searchCount; i++ {
		<-entered
	}
	if got := source.calls(); got != 1 {
		close(release)
		t.Fatalf("source calls while load blocked = %d, want 1", got)
	}
	close(release)

	for i := 0; i < searchCount; i++ {
		if err := <-errs; err != nil {
			t.Errorf("Search %d: %v", i+1, err)
		}
	}
}

func TestCanceledWaiterDoesNotCancelSharedLoad(t *testing.T) {
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	source := &mockSource{
		items:   stockList(),
		started: started,
		release: release,
	}
	catalog := NewCatalog(source, 30*time.Minute)

	leaderErr := make(chan error, 1)
	go func() {
		_, err := catalog.Search(context.Background(), "", 10, 0)
		leaderErr <- err
	}()
	<-started

	waiterCtx, cancelWaiter := context.WithCancel(context.Background())
	entered := make(chan struct{}, 1)
	waiterErr := make(chan error, 1)
	go func() {
		_, err := catalog.Search(&observedContext{Context: waiterCtx, entered: entered}, "", 10, 0)
		waiterErr <- err
	}()
	<-entered
	cancelWaiter()

	err := <-waiterErr
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled waiter error = %v, want context.Canceled", err)
	}
	if errors.Is(err, ErrSourceUnavailable) {
		t.Fatalf("canceled waiter error = %v, must not be ErrSourceUnavailable", err)
	}

	close(release)
	if err := <-leaderErr; err != nil {
		t.Fatalf("leader Search: %v", err)
	}
	if _, err := catalog.Search(context.Background(), "", 10, 0); err != nil {
		t.Fatalf("Search from cache: %v", err)
	}
	if got := source.calls(); got != 1 {
		t.Fatalf("source calls = %d, want 1 after shared load and cache hit", got)
	}
}

func TestCanceledLoadInitiatorDoesNotCancelSharedLoad(t *testing.T) {
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	source := &mockSource{
		items:   stockList(),
		started: started,
		release: release,
	}
	catalog := NewCatalog(source, 30*time.Minute)

	initiatorCtx, cancelInitiator := context.WithCancel(context.Background())
	initiatorEntered := make(chan struct{}, 1)
	initiatorErr := make(chan error, 1)
	go func() {
		_, err := catalog.Search(&observedContext{Context: initiatorCtx, entered: initiatorEntered}, "", 10, 0)
		initiatorErr <- err
	}()
	<-started
	<-initiatorEntered

	followerEntered := make(chan struct{}, 1)
	followerErr := make(chan error, 1)
	go func() {
		_, err := catalog.Search(&observedContext{Context: context.Background(), entered: followerEntered}, "", 10, 0)
		followerErr <- err
	}()
	<-followerEntered
	cancelInitiator()

	if err := <-initiatorErr; !errors.Is(err, context.Canceled) {
		close(release)
		t.Fatalf("initiator error = %v, want context.Canceled", err)
	}

	close(release)
	if err := <-followerErr; err != nil {
		t.Fatalf("follower Search: %v", err)
	}
	if _, err := catalog.Search(context.Background(), "", 10, 0); err != nil {
		t.Fatalf("Search from cache: %v", err)
	}
	if got := source.calls(); got != 1 {
		t.Fatalf("source calls = %d, want 1 after initiator cancellation", got)
	}
}

func TestDeadlineWaiterReturnsContextDeadlineExceeded(t *testing.T) {
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	source := &mockSource{
		items:   stockList(),
		started: started,
		release: release,
	}
	catalog := NewCatalog(source, 30*time.Minute)

	leaderErr := make(chan error, 1)
	go func() {
		_, err := catalog.Search(context.Background(), "", 10, 0)
		leaderErr <- err
	}()
	<-started

	waiterCtx, cancelWaiter := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancelWaiter()
	entered := make(chan struct{}, 1)
	waiterErr := make(chan error, 1)
	go func() {
		_, err := catalog.Search(&observedContext{Context: waiterCtx, entered: entered}, "", 10, 0)
		waiterErr <- err
	}()
	<-entered
	err := <-waiterErr
	if !errors.Is(err, context.DeadlineExceeded) {
		close(release)
		t.Fatalf("deadline waiter error = %v, want context.DeadlineExceeded", err)
	}
	if errors.Is(err, ErrSourceUnavailable) {
		close(release)
		t.Fatalf("deadline waiter error = %v, must not be ErrSourceUnavailable", err)
	}

	close(release)
	if err := <-leaderErr; err != nil {
		t.Fatalf("leader Search: %v", err)
	}
	if got := source.calls(); got != 1 {
		t.Fatalf("source calls = %d, want 1", got)
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
