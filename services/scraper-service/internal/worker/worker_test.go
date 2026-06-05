package worker

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/circuitbreaker"
	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/queue"
	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/scraper"
	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/storeconfig"
)

// fakeScraper returns a queued outcome per call, keyed by call order.
type fakeScraper struct {
	mu      sync.Mutex
	results map[string]queue.RawResult // by store id
	err     map[string]error           // by store id
}

func (f *fakeScraper) Scrape(_ context.Context, cfg *storeconfig.StoreConfig, job queue.Job) (queue.RawResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if e, ok := f.err[cfg.StoreID]; ok {
		return queue.RawResult{}, e
	}
	return f.results[cfg.StoreID], nil
}

// registryWith writes one json_api config per store id into a temp dir and loads it.
func registryWith(t *testing.T, storeIDs ...string) *storeconfig.Registry {
	t.Helper()
	dir := t.TempDir()
	for _, id := range storeIDs {
		body := `{"store_id":"` + id + `","name":"` + id + `","base_url":"https://x","mode":"json_api",` +
			`"search":{"url_template":"https://x?q={query}"},"json":{"title":"t","price":"p"}}`
		if err := os.WriteFile(filepath.Join(dir, id+".json"), []byte(body), 0o600); err != nil {
			t.Fatalf("write config: %v", err)
		}
	}
	reg, err := storeconfig.LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	return reg
}

// runUntil runs the pool until predicate holds (or the deadline), then cancels.
func runUntil(t *testing.T, p *Pool, predicate func(Stats) bool) Stats {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { p.Run(ctx); close(done) }()

	deadline := time.After(2 * time.Second)
	for {
		if predicate(p.Stats()) {
			cancel()
			<-done
			return p.Stats()
		}
		select {
		case <-deadline:
			cancel()
			<-done
			t.Fatalf("timed out waiting for predicate; stats=%+v", p.Stats())
		case <-time.After(2 * time.Millisecond):
		}
	}
}

func TestProcessSuccessPublishes(t *testing.T) {
	q := queue.NewMemoryQueue(queue.Job{StoreID: "shopA", Query: "x"})
	fake := &fakeScraper{results: map[string]queue.RawResult{
		"shopA": {StoreID: "shopA", Products: []queue.RawProduct{{Title: "t", Price: 1}}},
	}}
	reg := registryWith(t, "shopA")
	breakers := circuitbreaker.NewGroup(3, time.Minute)
	p := New(q, q, reg, fake, breakers, 2, nil)

	stats := runUntil(t, p, func(s Stats) bool { return s.Succeeded >= 1 })
	if stats.Succeeded != 1 || stats.Failed != 0 {
		t.Fatalf("stats = %+v", stats)
	}
	if len(q.Published) != 1 || q.Published[0].StoreID != "shopA" {
		t.Fatalf("published = %+v", q.Published)
	}
}

func TestProcessUnknownStoreSkipped(t *testing.T) {
	q := queue.NewMemoryQueue(queue.Job{StoreID: "ghost", Query: "x"})
	reg := registryWith(t, "shopA")
	p := New(q, q, reg, &fakeScraper{}, circuitbreaker.NewGroup(3, time.Minute), 1, nil)

	stats := runUntil(t, p, func(s Stats) bool { return s.Processed >= 1 })
	if stats.SkippedNoConfig != 1 || stats.Succeeded != 0 {
		t.Fatalf("stats = %+v", stats)
	}
}

func TestBlockedTripsBreakerAndSkips(t *testing.T) {
	// Threshold 2: two blocked jobs trip the breaker, the third is skipped open.
	jobs := []queue.Job{
		{StoreID: "shopA", Query: "1"},
		{StoreID: "shopA", Query: "2"},
		{StoreID: "shopA", Query: "3"},
	}
	q := queue.NewMemoryQueue(jobs...)
	fake := &fakeScraper{err: map[string]error{"shopA": scraper.ErrBlocked}}
	reg := registryWith(t, "shopA")
	// Single worker keeps ordering deterministic so the breaker trips before job 3.
	p := New(q, q, reg, fake, circuitbreaker.NewGroup(2, time.Minute), 1, nil)

	stats := runUntil(t, p, func(s Stats) bool { return s.Processed >= 3 })
	if stats.Blocked != 2 {
		t.Errorf("Blocked = %d, want 2", stats.Blocked)
	}
	if stats.SkippedOpen != 1 {
		t.Errorf("SkippedOpen = %d, want 1 (breaker should be open for job 3)", stats.SkippedOpen)
	}
	if stats.Succeeded != 0 {
		t.Errorf("Succeeded = %d, want 0", stats.Succeeded)
	}
}

func TestNonBlockingFailureCounted(t *testing.T) {
	q := queue.NewMemoryQueue(queue.Job{StoreID: "shopA", Query: "x"})
	fake := &fakeScraper{err: map[string]error{"shopA": context.DeadlineExceeded}}
	reg := registryWith(t, "shopA")
	p := New(q, q, reg, fake, circuitbreaker.NewGroup(2, time.Minute), 1, nil)

	stats := runUntil(t, p, func(s Stats) bool { return s.Processed >= 1 })
	if stats.Failed != 1 || stats.Blocked != 0 {
		t.Fatalf("stats = %+v (non-blocking error should be Failed, not Blocked)", stats)
	}
}
