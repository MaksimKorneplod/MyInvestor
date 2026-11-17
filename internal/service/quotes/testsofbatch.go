package quotes

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MaksimKorneplod/MyInvestor/internal/repository/moex"
)

type fakeFetcher struct {
	calls    int32
	bulk     []moex.Quote
	byTicker map[string]moex.Quote
	fail     map[string]error
}

func (f *fakeFetcher) Quotes(_ context.Context, ref moex.Instrument) ([]moex.Quote, error) {
	atomic.AddInt32(&f.calls, 1)
	if ref.Security == "" {
		return f.bulk, nil
	}
	if err, ok := f.fail[ref.Security]; ok {
		return nil, err
	}
	if q, ok := f.byTicker[ref.Security]; ok {
		return []moex.Quote{q}, nil
	}
	return []moex.Quote{}, nil
}

func TestBatch_Strategies(t *testing.T) {
	t.Run("one ticker → per-security", func(t *testing.T) {
		f := &fakeFetcher{byTicker: map[string]moex.Quote{"SBER": {SecID: "SBER"}}}
		s := New(f, time.Second)
		got, err := s.Batch(context.Background(), moex.Instrument{Engine: "stock", Market: "shares"}, []string{"sber"})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[0].SecID != "SBER" {
			t.Fatalf("got %+v", got)
		}
		if f.calls != 1 {
			t.Fatalf("calls = %d", f.calls)
		}
	})

	t.Run("2..5 tickers → parallel per-security", func(t *testing.T) {
		f := &fakeFetcher{byTicker: map[string]moex.Quote{
			"SBER": {SecID: "SBER"}, "GAZP": {SecID: "GAZP"}, "LKOH": {SecID: "LKOH"},
		}}
		s := New(f, time.Second)
		got, err := s.Batch(context.Background(),
			moex.Instrument{Engine: "stock", Market: "shares"},
			[]string{"SBER", "GAZP", "LKOH"})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 3 {
			t.Fatalf("got %+v", got)
		}
		if f.calls != 3 {
			t.Fatalf("calls = %d (want 3 per-security)", f.calls)
		}
	})

	t.Run(">5 tickers → bulk", func(t *testing.T) {
		bulk := []moex.Quote{
			{SecID: "SBER"}, {SecID: "GAZP"}, {SecID: "LKOH"},
			{SecID: "GMKN"}, {SecID: "YNDX"}, {SecID: "ROSN"},
			{SecID: "NOT_WANTED"},
		}
		f := &fakeFetcher{bulk: bulk}
		s := New(f, time.Second)
		got, err := s.Batch(context.Background(),
			moex.Instrument{Engine: "stock", Market: "shares"},
			[]string{"SBER", "GAZP", "LKOH", "GMKN", "YNDX", "ROSN"})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 6 {
			t.Fatalf("got %+v", got)
		}
		if f.calls != 1 {
			t.Fatalf("calls = %d (want 1 bulk)", f.calls)
		}
	})

	t.Run("bulk preserves request order", func(t *testing.T) {
		f := &fakeFetcher{bulk: []moex.Quote{
			{SecID: "A"}, {SecID: "B"}, {SecID: "C"},
		}}
		s := New(f, time.Second)
		got, _ := s.Batch(context.Background(),
			moex.Instrument{Engine: "stock", Market: "shares"},
			[]string{"C", "A", "B", "C", "A", "B"}) // 6 тикеров, после dedupe 3
		// 3 ≤ 5 → parallel; проверяем порядок
		want := []string{"C", "A", "B"}
		for i, q := range got {
			if q.SecID != want[i] {
				t.Fatalf("pos %d = %s, want %s", i, q.SecID, want[i])
			}
		}
	})
}

func TestBatch_DedupeBeforeLimit(t *testing.T) {
	f := &fakeFetcher{}
	s := New(f, time.Second)
	in := make([]string, 1000)
	for i := range in {
		in[i] = "SBER"
	}
	got, err := s.Batch(context.Background(),
		moex.Instrument{Engine: "stock", Market: "shares"}, in)
	if err != nil {
		t.Fatalf("1000 copies of SBER should pass: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestBatch_TooManyUnique(t *testing.T) {
	f := &fakeFetcher{}
	s := New(f, time.Second)
	in := make([]string, 201)
	for i := range in {
		in[i] = "T" + string(rune('A'+i%26)) + string(rune('0'+i/26))
	}
	_, err := s.Batch(context.Background(),
		moex.Instrument{Engine: "stock", Market: "shares"}, in)
	if !errors.Is(err, moex.ErrRequest) {
		t.Fatalf("err = %v", err)
	}
}

func TestBatch_PartialFailureInParallel(t *testing.T) {
	f := &fakeFetcher{
		byTicker: map[string]moex.Quote{
			"SBER": {SecID: "SBER"}, "LKOH": {SecID: "LKOH"},
		},
		fail: map[string]error{"GAZP": moex.ErrNetwork},
	}
	s := New(f, time.Second)
	_, err := s.Batch(context.Background(),
		moex.Instrument{Engine: "stock", Market: "shares"},
		[]string{"SBER", "GAZP", "LKOH"})
	if !errors.Is(err, moex.ErrNetwork) {
		t.Fatalf("err = %v", err)
	}
}

func TestBatch_EmptyRejected(t *testing.T) {
	f := &fakeFetcher{}
	s := New(f, time.Second)
	_, err := s.Batch(context.Background(),
		moex.Instrument{Engine: "stock", Market: "shares"}, nil)
	if !errors.Is(err, moex.ErrRequest) {
		t.Fatalf("err = %v", err)
	}
}

func TestBatch_BadTickerRejected(t *testing.T) {
	f := &fakeFetcher{}
	s := New(f, time.Second)
	for _, bad := range []string{"SBER/P", "../etc", "SBER P", ""} {
		_, err := s.Batch(context.Background(),
			moex.Instrument{Engine: "stock", Market: "shares"}, []string{bad})
		if !errors.Is(err, moex.ErrRequest) {
			t.Fatalf("%q should fail: %v", bad, err)
		}
	}
}
