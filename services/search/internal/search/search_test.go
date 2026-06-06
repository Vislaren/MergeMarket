package search

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/Vislaren/MergeMarket/services/search/internal/cache"
	"github.com/Vislaren/MergeMarket/services/search/internal/store"
)

type fakeRepo struct {
	mu      sync.Mutex
	results []store.Product
	calls   int
	err     error
}

func (f *fakeRepo) Search(_ context.Context, _, _ string, _ int) ([]store.Product, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	// Return a copy so the orchestrator's in-place scoring can't mutate the source.
	out := make([]store.Product, len(f.results))
	copy(out, f.results)
	return out, nil
}
func (f *fakeRepo) Close() error { return nil }

func (f *fakeRepo) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

type fakeCache struct {
	mu      sync.Mutex
	entries map[string]cache.Entry
	sets    int
}

func newFakeCache() *fakeCache { return &fakeCache{entries: map[string]cache.Entry{}} }

func (f *fakeCache) Get(_ context.Context, key string) (cache.Entry, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	e, ok := f.entries[key]
	return e, ok, nil
}
func (f *fakeCache) Set(_ context.Context, key string, entry cache.Entry, _ time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.entries[key] = entry
	f.sets++
	return nil
}

func testConfig() Config {
	return Config{CachePrefix: "search:", CacheTTL: 15 * time.Minute, CacheStaleAfter: 5 * time.Minute, MaxResults: 50}
}

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestSearchMissQueriesDBScoresAndCaches(t *testing.T) {
	repo := &fakeRepo{results: []store.Product{
		{ProductID: "a", TotalCost: 100},
		{ProductID: "b", TotalCost: 200},
		{ProductID: "c", TotalCost: 150},
	}}
	c := newFakeCache()
	svc := New(repo, c, testConfig(), quietLogger())

	res, err := svc.Search(context.Background(), "phone", "CM")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if res.Cached {
		t.Error("first call should not be cached")
	}
	if res.Query != "phone" || len(res.Results) != 3 {
		t.Fatalf("unexpected response: %+v", res)
	}
	// Cheapest (100) → 100, dearest (200) → 0, middle (150) → 50.
	got := map[string]int{}
	for _, r := range res.Results {
		got[r.ProductID] = r.DealScore
	}
	if got["a"] != 100 || got["b"] != 0 || got["c"] != 50 {
		t.Errorf("deal scores wrong: %v", got)
	}
	if c.sets != 1 {
		t.Errorf("expected one cache write, got %d", c.sets)
	}
}

func TestSearchFreshCacheHitSkipsDB(t *testing.T) {
	repo := &fakeRepo{results: []store.Product{{ProductID: "a", TotalCost: 100}}}
	c := newFakeCache()
	svc := New(repo, c, testConfig(), quietLogger())
	fixed := time.Now()
	svc.now = func() time.Time { return fixed }

	if _, err := svc.Search(context.Background(), "phone", "CM"); err != nil {
		t.Fatal(err)
	}
	res, err := svc.Search(context.Background(), "phone", "CM")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Cached {
		t.Error("second call should be served from cache")
	}
	if repo.callCount() != 1 {
		t.Errorf("DB should be hit once, got %d", repo.callCount())
	}
}

func TestSearchStaleHitTriggersRevalidation(t *testing.T) {
	repo := &fakeRepo{results: []store.Product{{ProductID: "a", TotalCost: 100}}}
	c := newFakeCache()
	svc := New(repo, c, testConfig(), quietLogger())

	// Seed a stale entry (cached 10m ago, stale threshold is 5m).
	key := cache.Key("search:", "phone", "CM")
	c.entries[key] = cache.Entry{
		Results:  []store.Product{{ProductID: "a", TotalCost: 100, DealScore: 100}},
		CachedAt: time.Now().Add(-10 * time.Minute),
	}

	res, err := svc.Search(context.Background(), "phone", "CM")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Cached {
		t.Error("stale hit should still be reported as cached")
	}
	// Background revalidation should eventually hit the DB.
	deadline := time.Now().Add(2 * time.Second)
	for repo.callCount() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if repo.callCount() == 0 {
		t.Error("expected background revalidation to query the DB")
	}
}

func TestSearchDBErrorPropagates(t *testing.T) {
	repo := &fakeRepo{err: errors.New("boom")}
	svc := New(repo, newFakeCache(), testConfig(), quietLogger())
	if _, err := svc.Search(context.Background(), "phone", "CM"); err == nil {
		t.Fatal("expected error from DB failure")
	}
}

func TestScoreDealsAllEqual(t *testing.T) {
	results := []store.Product{{TotalCost: 50}, {TotalCost: 50}}
	scoreDeals(results)
	for _, r := range results {
		if r.DealScore != 100 {
			t.Errorf("equal prices should all score 100, got %d", r.DealScore)
		}
	}
}
