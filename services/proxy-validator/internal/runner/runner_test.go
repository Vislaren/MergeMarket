package runner

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Vislaren/MergeMarket/services/proxy-validator/internal/politeness"
	"github.com/Vislaren/MergeMarket/services/proxy-validator/internal/proxy"
)

type fakeFetcher struct {
	addrs []proxy.Addr
	err   error
}

func (f fakeFetcher) FetchAll(context.Context) ([]proxy.Addr, error) { return f.addrs, f.err }

// fakeChecker marks proxies whose IP is in `good` as valid.
type fakeChecker struct{ good map[string]bool }

func (c fakeChecker) Validate(_ context.Context, a proxy.Addr) error {
	if c.good[a.IP] {
		return nil
	}
	return errors.New("bad proxy")
}

type fakeWriter struct {
	mu      sync.Mutex
	members []string
	ttl     time.Duration
	calls   int
}

func (w *fakeWriter) Replace(_ context.Context, members []string, ttl time.Duration) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.members = members
	w.ttl = ttl
	w.calls++
	return nil
}

func newFastLimiter() *politeness.Limiter {
	// Zero-length window => no real sleeping, keeps the test fast.
	return politeness.New(0, 0, 1)
}

func TestRunOnceWritesOnlyWorkingProxies(t *testing.T) {
	fetch := fakeFetcher{addrs: []proxy.Addr{
		{IP: "1.1.1.1", Port: 80},
		{IP: "2.2.2.2", Port: 80},
		{IP: "3.3.3.3", Port: 80},
	}}
	check := fakeChecker{good: map[string]bool{"1.1.1.1": true, "3.3.3.3": true}}
	writer := &fakeWriter{}

	r := New(fetch, check, writer, newFastLimiter(), 4, 5*time.Minute, nil)
	if err := r.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce error = %v", err)
	}

	if len(writer.members) != 2 {
		t.Fatalf("wrote %d members, want 2: %#v", len(writer.members), writer.members)
	}
	if writer.ttl != 5*time.Minute {
		t.Errorf("ttl = %s, want 5m", writer.ttl)
	}

	st := r.Stats()
	if !st.HasRun || st.Fetched != 3 || st.Working != 2 {
		t.Errorf("stats = %+v, want fetched=3 working=2 hasRun=true", st)
	}
}

func TestRunOnceFetchErrorPropagates(t *testing.T) {
	r := New(fakeFetcher{err: errors.New("network down")}, fakeChecker{}, &fakeWriter{}, newFastLimiter(), 2, time.Minute, nil)
	if err := r.RunOnce(context.Background()); err == nil {
		t.Fatal("RunOnce should propagate a fetch error")
	}
}

func TestRunOnceAllBadClearsPool(t *testing.T) {
	fetch := fakeFetcher{addrs: []proxy.Addr{{IP: "9.9.9.9", Port: 80}}}
	writer := &fakeWriter{}
	r := New(fetch, fakeChecker{good: map[string]bool{}}, writer, newFastLimiter(), 2, time.Minute, nil)
	if err := r.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce error = %v", err)
	}
	if len(writer.members) != 0 {
		t.Fatalf("expected empty member set, got %#v", writer.members)
	}
	if writer.calls != 1 {
		t.Errorf("writer called %d times, want 1", writer.calls)
	}
}

func TestLoopStopsOnContextCancel(t *testing.T) {
	fetch := fakeFetcher{addrs: []proxy.Addr{{IP: "1.1.1.1", Port: 80}}}
	check := fakeChecker{good: map[string]bool{"1.1.1.1": true}}
	r := New(fetch, check, &fakeWriter{}, newFastLimiter(), 2, time.Minute, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		r.Loop(ctx, time.Hour)
		close(done)
	}()

	// Give the initial RunOnce a moment, then cancel.
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Loop did not stop within 1s of context cancellation")
	}
}
