package worker

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Vislaren/MergeMarket/services/normalization/internal/affiliate"
	"github.com/Vislaren/MergeMarket/services/normalization/internal/queue"
	"github.com/Vislaren/MergeMarket/services/normalization/internal/store"
)

// fakeRepo records upserts and can be told to fail.
type fakeRepo struct {
	mu      sync.Mutex
	upserts []store.Product
	failURL string
}

func (f *fakeRepo) UpsertProduct(_ context.Context, p store.Product) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if p.URL == f.failURL {
		return "", errors.New("boom")
	}
	f.upserts = append(f.upserts, p)
	return "id-" + p.URL, nil
}

func (f *fakeRepo) Close() error { return nil }

func (f *fakeRepo) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.upserts)
}

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func runUntilDrained(t *testing.T, p *Pool, src *queue.MemorySource) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { p.Run(ctx); close(done) }()

	deadline := time.After(2 * time.Second)
	for src.Len() > 0 {
		select {
		case <-deadline:
			cancel()
			t.Fatal("timed out waiting for source to drain")
		case <-time.After(5 * time.Millisecond):
		}
	}
	// Give workers a moment to finish the in-flight result, then stop.
	time.Sleep(20 * time.Millisecond)
	cancel()
	<-done
}

func TestProcess_NormalizesInjectsPersists(t *testing.T) {
	src := queue.NewMemory(queue.RawResult{
		StoreID:   "jumia-cm",
		Store:     "Jumia",
		Query:     "phone",
		ScrapedAt: time.Now().UTC(),
		Products: []queue.RawProduct{
			{Title: "Phone", Price: 100, Shipping: 5, Currency: "USD", URL: "https://jumia.cm/p/1", ImageURL: "https://img/1"},
			{Title: "", Price: 1}, // skipped
		},
	})

	cfgJSON := `{"stores":{"jumia-cm":{"params":{"aff":"mm-21"}}}}`
	inj := mustInjector(t, cfgJSON)
	repo := &fakeRepo{}

	pool := New(src, repo, inj, 1, quietLogger())
	runUntilDrained(t, pool, src)

	if repo.count() != 1 {
		t.Fatalf("expected 1 upsert, got %d", repo.count())
	}
	got := repo.upserts[0]
	if got.Affiliate != "https://jumia.cm/p/1?aff=mm-21" {
		t.Errorf("affiliate not injected: %q", got.Affiliate)
	}
	if got.BaseURL != "https://jumia.cm" {
		t.Errorf("base url not derived: %q", got.BaseURL)
	}
	if got.Store != "Jumia" || got.StoreID != "jumia-cm" {
		t.Errorf("store fields wrong: %+v", got)
	}

	st := pool.Stats()
	if st.ResultsProcessed != 1 || st.ProductsWritten != 1 || st.ProductsSkipped != 1 || st.ProductsIn != 2 {
		t.Errorf("unexpected stats: %+v", st)
	}
}

func TestProcess_PersistErrorCounted(t *testing.T) {
	src := queue.NewMemory(queue.RawResult{
		Store: "S", StoreID: "s",
		Products: []queue.RawProduct{
			{Title: "ok", Price: 1, URL: "https://s/ok"},
			{Title: "bad", Price: 1, URL: "https://s/bad"},
		},
	})
	repo := &fakeRepo{failURL: "https://s/bad"}
	pool := New(src, repo, affiliate.New(), 1, quietLogger())
	runUntilDrained(t, pool, src)

	st := pool.Stats()
	if st.ProductsWritten != 1 {
		t.Errorf("written = %d, want 1", st.ProductsWritten)
	}
	if st.PersistErrors != 1 {
		t.Errorf("persist errors = %d, want 1", st.PersistErrors)
	}
}

func mustInjector(t *testing.T, body string) *affiliate.Injector {
	t.Helper()
	path := filepath.Join(t.TempDir(), "a.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	inj, err := affiliate.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return inj
}
