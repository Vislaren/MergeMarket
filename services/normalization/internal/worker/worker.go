// Package worker runs the normalization-service worker pool. Each worker pulls a
// RawResult off the normalization queue, normalizes every product to the
// canonical schema, injects retailer-specific affiliate links, and upserts the
// products into PostgreSQL. A single malformed product is skipped rather than
// failing the batch (NFR-2), and a transient persistence error is logged without
// halting the worker.
package worker

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"github.com/Vislaren/MergeMarket/services/normalization/internal/affiliate"
	"github.com/Vislaren/MergeMarket/services/normalization/internal/normalize"
	"github.com/Vislaren/MergeMarket/services/normalization/internal/queue"
	"github.com/Vislaren/MergeMarket/services/normalization/internal/store"
)

// dequeueErrorBackoff is how long a worker pauses after a failed (non-empty,
// non-cancellation) dequeue — typically Redis being unreachable — so the loop
// does not hot-spin while the dependency recovers.
const dequeueErrorBackoff = time.Second

// Stats is a snapshot of the pool's cumulative counters, surfaced via /stats.
type Stats struct {
	ResultsProcessed int64     `json:"results_processed"` // RawResult items dequeued
	ProductsIn       int64     `json:"products_in"`       // raw products seen
	ProductsWritten  int64     `json:"products_written"`  // products upserted to the DB
	ProductsSkipped  int64     `json:"products_skipped"`  // products dropped by normalization
	PersistErrors    int64     `json:"persist_errors"`    // upsert failures
	LastResultAt     time.Time `json:"last_result_at"`
}

// Pool is a fixed-size set of workers consuming from one Source.
type Pool struct {
	source   queue.Source
	repo     store.Repository
	injector *affiliate.Injector
	workers  int
	log      *slog.Logger

	mu    sync.Mutex
	stats Stats
}

// New constructs a Pool. workers is clamped to at least 1.
func New(src queue.Source, repo store.Repository, inj *affiliate.Injector, workers int, log *slog.Logger) *Pool {
	if workers < 1 {
		workers = 1
	}
	if log == nil {
		log = slog.Default()
	}
	if inj == nil {
		inj = affiliate.New()
	}
	return &Pool{source: src, repo: repo, injector: inj, workers: workers, log: log}
}

// Stats returns a copy of the current counters.
func (p *Pool) Stats() Stats {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.stats
}

// Run starts the workers and blocks until ctx is cancelled, then waits for the
// in-flight workers to drain.
func (p *Pool) Run(ctx context.Context) {
	var wg sync.WaitGroup
	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			p.loop(ctx, id)
		}(i)
	}
	wg.Wait()
	p.log.Info("worker pool stopped")
}

// loop is one worker's dequeue→process cycle until the context is cancelled.
func (p *Pool) loop(ctx context.Context, id int) {
	for {
		if ctx.Err() != nil {
			return
		}
		raw, err := p.source.Dequeue(ctx)
		if err != nil {
			if errors.Is(err, queue.ErrEmpty) {
				continue // idle poll timeout — try again
			}
			if ctx.Err() != nil {
				return // shutting down
			}
			p.log.Error("dequeue failed; backing off", "worker", id, "backoff", dequeueErrorBackoff.String(), "error", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(dequeueErrorBackoff):
			}
			continue
		}
		p.process(ctx, raw)
	}
}

// process normalizes one RawResult and persists its products.
func (p *Pool) process(ctx context.Context, raw queue.RawResult) {
	res := normalize.FromRaw(raw)
	skipped := len(raw.Products) - len(res.Products)

	p.mark(func(s *Stats) {
		s.ResultsProcessed++
		s.ProductsIn += int64(len(raw.Products))
		s.ProductsSkipped += int64(skipped)
		s.LastResultAt = time.Now().UTC()
	})

	written := 0
	for _, np := range res.Products {
		affiliateURL := p.injector.Inject(np.StoreID, np.URL)

		_, err := p.repo.UpsertProduct(ctx, store.Product{
			StoreID:   np.StoreID,
			Store:     np.Store,
			BaseURL:   baseURL(np.URL),
			Title:     np.Title,
			URL:       np.URL,
			Affiliate: affiliateURL,
			ImageURL:  np.ImageURL,
			Currency:  np.Currency,
			Price:     np.Price,
			Shipping:  np.Shipping,
			ScrapedAt: np.ScrapedAt,
		})
		if err != nil {
			p.mark(func(s *Stats) { s.PersistErrors++ })
			p.log.Error("persist product failed", "store", np.Store, "url", np.URL, "error", err)
			continue
		}
		written++
	}

	p.mark(func(s *Stats) { s.ProductsWritten += int64(written) })
	p.log.Info("normalized result",
		"store", res.Store, "query", res.Query,
		"in", len(raw.Products), "written", written, "skipped", skipped)
}

// mark applies fn to the stats under the pool mutex.
func (p *Pool) mark(fn func(*Stats)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	fn(&p.stats)
}

// baseURL extracts scheme://host from an absolute product URL for the stores
// row's base_url. It returns "" when the URL is relative or unparseable.
func baseURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}
