// Package worker runs the scraper's worker pool. Each worker pulls a job from
// the queue, loads the relevant StoreConfig, consults the store's circuit
// breaker, performs the scrape, and publishes the raw result to the
// normalization queue. Blocking (403/429) responses trip the breaker so a store
// that is rate-limiting us is left alone until it cools down (ARCHITECTURE §2).
package worker

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/circuitbreaker"
	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/queue"
	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/scraper"
	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/storeconfig"
)

// dequeueErrorBackoff is how long a worker pauses after a failed (non-empty,
// non-cancellation) dequeue — typically Redis being unreachable. Without it the
// loop would hot-spin, flooding logs and burning CPU until Redis returns.
const dequeueErrorBackoff = time.Second

// Scraper is the scrape engine the pool depends on (implemented by the scraper
// package). Kept as an interface so the pool is unit-testable with a fake.
type Scraper interface {
	Scrape(ctx context.Context, cfg *storeconfig.StoreConfig, job queue.Job) (queue.RawResult, error)
}

// Stats is a snapshot of the pool's cumulative counters, surfaced via /stats.
type Stats struct {
	Processed       int64     `json:"processed"`         // jobs dequeued and acted on
	Succeeded       int64     `json:"succeeded"`         // scrapes that published a result
	Blocked         int64     `json:"blocked"`           // 403/429 responses (fed to breakers)
	Failed          int64     `json:"failed"`            // other scrape errors
	SkippedNoConfig int64     `json:"skipped_no_config"` // jobs for an unknown store id
	SkippedOpen     int64     `json:"skipped_open"`      // jobs skipped because the breaker was open
	LastJobAt       time.Time `json:"last_job_at"`
}

// Pool is a fixed-size set of workers consuming from one queue.
type Pool struct {
	queue    queue.Queue
	sink     queue.Sink
	registry *storeconfig.Registry
	scraper  Scraper
	breakers *circuitbreaker.Group
	workers  int
	log      *slog.Logger

	mu    sync.Mutex
	stats Stats
}

// New constructs a Pool. workers is clamped to at least 1.
func New(q queue.Queue, sink queue.Sink, reg *storeconfig.Registry, s Scraper, breakers *circuitbreaker.Group, workers int, log *slog.Logger) *Pool {
	if workers < 1 {
		workers = 1
	}
	if log == nil {
		log = slog.Default()
	}
	return &Pool{
		queue:    q,
		sink:     sink,
		registry: reg,
		scraper:  s,
		breakers: breakers,
		workers:  workers,
		log:      log,
	}
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
		job, err := p.queue.Dequeue(ctx)
		if err != nil {
			if errors.Is(err, queue.ErrEmpty) {
				continue // idle poll timeout — try again
			}
			if ctx.Err() != nil {
				return // shutting down
			}
			// A real error (e.g. Redis unreachable): back off so we don't
			// hot-spin while the dependency recovers.
			p.log.Error("dequeue failed; backing off", "worker", id, "backoff", dequeueErrorBackoff.String(), "error", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(dequeueErrorBackoff):
			}
			continue
		}
		p.process(ctx, job)
	}
}

// process handles a single job end to end and updates stats.
func (p *Pool) process(ctx context.Context, job queue.Job) {
	p.mark(func(s *Stats) { s.Processed++; s.LastJobAt = time.Now().UTC() })

	cfg, ok := p.registry.Get(job.StoreID)
	if !ok {
		p.mark(func(s *Stats) { s.SkippedNoConfig++ })
		p.log.Warn("no config for store; skipping job", "store_id", job.StoreID, "job_id", job.JobID)
		return
	}

	breaker := p.breakers.For(job.StoreID)
	if !breaker.Allow() {
		p.mark(func(s *Stats) { s.SkippedOpen++ })
		p.log.Info("circuit open; skipping job", "store_id", job.StoreID, "state", string(breaker.State()))
		return
	}

	result, err := p.scraper.Scrape(ctx, cfg, job)
	if err != nil {
		if errors.Is(err, scraper.ErrBlocked) {
			breaker.Failure()
			p.mark(func(s *Stats) { s.Blocked++ })
			p.log.Warn("store blocked request; breaker updated",
				"store_id", job.StoreID, "state", string(breaker.State()), "error", err)
			return
		}
		// Non-blocking failures don't trip the breaker but are still failures.
		p.mark(func(s *Stats) { s.Failed++ })
		p.log.Error("scrape failed", "store_id", job.StoreID, "job_id", job.JobID, "error", err)
		return
	}

	breaker.Success()
	if err := p.sink.Publish(ctx, result); err != nil {
		p.mark(func(s *Stats) { s.Failed++ })
		p.log.Error("publish result failed", "store_id", job.StoreID, "error", err)
		return
	}

	p.mark(func(s *Stats) { s.Succeeded++ })
	p.log.Info("scrape published",
		"store_id", job.StoreID, "query", job.Query, "products", len(result.Products))
}

// mark applies fn to the stats under the pool mutex.
func (p *Pool) mark(fn func(*Stats)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	fn(&p.stats)
}
