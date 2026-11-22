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

type recordingFetcher struct {
	got moex.Instrument
}

func (f *recordingFetcher) Quotes(_ context.Context, ref moex.Instrument) ([]moex.Quote, error) {
	f.got = ref
	return []moex.Quote{{SecID: ref.Security}}, nil
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
	t.Run("parallel preserves request order", func(t *testing.T) {
		// ИСПРАВЛЕНО: Используем byTicker, так как 3 тикера пойдут в параллельный запрос,
		// а не в bulk (для bulk нужно >5 тикеров).
		f := &fakeFetcher{byTicker: map[string]moex.Quote{
			"A": {SecID: "A"}, "B": {SecID: "B"}, "C": {SecID: "C"},
		}}
		s := New(f, time.Second)

		// ИСПРАВЛЕНО: Проверяем ошибку (err)
		got, err := s.Batch(context.Background(),
			moex.Instrument{Engine: "stock", Market: "shares"},
			[]string{"C", "A", "B", "C", "A", "B"}) // 6 тикеров, после dedupe 3
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []string{"C", "A", "B"}

		// ИСПРАВЛЕНО: Явно проверяем длину ответа
		if len(got) != len(want) {
			t.Fatalf("got %d quotes, want %d", len(got), len(want))
		}

		// Теперь цикл безопасен, так как мы уверены, что длина совпадает
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
	if !errors.Is(err, ErrInvalidRequest) {
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
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v", err)
	}
}

func TestBatch_EmptyRejected(t *testing.T) {
	f := &fakeFetcher{}
	s := New(f, time.Second)
	_, err := s.Batch(context.Background(),
		moex.Instrument{Engine: "stock", Market: "shares"}, nil)
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("err = %v", err)
	}
}

func TestBatch_BadTickerRejected(t *testing.T) {
	f := &fakeFetcher{}
	s := New(f, time.Second)
	for _, bad := range []string{"SBER/P", "../etc", "SBER P", ""} {
		_, err := s.Batch(context.Background(),
			moex.Instrument{Engine: "stock", Market: "shares"}, []string{bad})
		if !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("%q should fail: %v", bad, err)
		}
	}
}

func TestOne_KeepsMarketPathCase(t *testing.T) {
	f := &recordingFetcher{}
	s := New(f, time.Second)
	_, err := s.One(context.Background(), moex.Instrument{Engine: "stock", Market: "shares", Security: "sber"})
	if err != nil {
		t.Fatal(err)
	}
	if f.got.Engine != "stock" || f.got.Market != "shares" || f.got.Security != "SBER" {
		t.Fatalf("ISS path parts = %+v", f.got)
	}
}

// НОВОЕ: Мок для проверки singleflight, который умеет блокироваться по команде
type blockingFetcher struct {
	started chan struct{}
	release chan struct{}
}

func (b *blockingFetcher) Quotes(ctx context.Context, _ moex.Instrument) ([]moex.Quote, error) {
	close(b.started)
	select {
	case <-b.release:
		price := moex.Decimal{Text: "250.0"}                    // Создаем переменную нужного типа
		return []moex.Quote{{SecID: "SBER", Last: &price}}, nil // Передаем указатель на неё (&price)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// НОВОЕ: Тест на отмену лидера в singleflight
func TestService_Singleflight_LeaderCancellation(t *testing.T) {
	f := &blockingFetcher{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	s := New(f, time.Second)
	ref := moex.Instrument{Engine: "stock", Market: "shares", Security: "SBER"}

	leaderCtx, cancelLeader := context.WithCancel(context.Background())
	defer cancelLeader()

	leaderDone := make(chan error)
	go func() {
		_, err := s.One(leaderCtx, ref)
		leaderDone <- err
	}()

	<-f.started // Ждем, пока лидер начнет запрос

	followerDone := make(chan error)
	go func() {
		_, err := s.One(context.Background(), ref)
		followerDone <- err
	}()

	time.Sleep(10 * time.Millisecond) // Даем догоняющему время встать в очередь singleflight

	cancelLeader() // Отменяем контекст лидера

	// Лидер должен упасть с context.Canceled
	if err := <-leaderDone; !errors.Is(err, context.Canceled) {
		t.Fatalf("expected leader to fail with context.Canceled, got: %v", err)
	}

	close(f.release) // Разрешаем фоновому запросу завершиться

	// Догоняющий должен успешно получить данные, несмотря на отмену лидера
	if err := <-followerDone; err != nil {
		t.Fatalf("follower failed: %v", err)
	}
}
