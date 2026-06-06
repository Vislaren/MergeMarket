// Package runner orchestrates the proxy-validator's core loop: scrape proxy
// lists, validate each candidate concurrently (bounded, with politeness
// delays between dispatches), and write the working set to Redis. It exposes
// last-cycle statistics for the /health endpoint.
package runner

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/Vislaren/MergeMarket/services/proxy-validator/internal/politeness"
	"github.com/Vislaren/MergeMarket/services/proxy-validator/internal/proxy"
)

// Fetcher supplies candidate proxies (implemented by the fetcher package).
type Fetcher interface {
	FetchAll(ctx context.Context) ([]proxy.Addr, error)
}

// Checker validates a single proxy (implemented by the validator package).
type Checker interface {
	Validate(ctx context.Context, addr proxy.Addr) error
}

// Writer persists the working pool (implemented by the store package).
type Writer interface {
	Replace(ctx context.Context, members []string, ttl time.Duration) error
}

// Stats is a snapshot of the most recently completed validation cycle.
type Stats struct {
	LastRunAt     time.Time `json:"last_run_at"`
	Fetched       int       `json:"fetched"`
	Working       int       `json:"working"`
	CycleDuration string    `json:"cycle_duration"`
	HasRun        bool      `json:"has_run"`
}

// Runner ties the pipeline together and tracks cycle statistics.
type Runner struct {
	fetcher     Fetcher
	checker     Checker
	writer      Writer
	limiter     *politeness.Limiter
	concurrency int
	poolTTL     time.Duration
	log         *slog.Logger

	mu    sync.RWMutex
	stats Stats
}

// New constructs a Runner. concurrency must be >= 1.
func New(f Fetcher, c Checker, w Writer, limiter *politeness.Limiter, concurrency int, poolTTL time.Duration, log *slog.Logger) *Runner {
	if concurrency < 1 {
		concurrency = 1
	}
	if log == nil {
		log = slog.Default()
	}
	return &Runner{
		fetcher:     f,
		checker:     c,
		writer:      w,
		limiter:     limiter,
		concurrency: concurrency,
		poolTTL:     poolTTL,
		log:         log,
	}
}

// Stats returns a copy of the latest cycle statistics.
func (r *Runner) Stats() Stats {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.stats
}

// RunOnce executes a single scrape→validate→write cycle and records its stats.
func (r *Runner) RunOnce(ctx context.Context) error {
	start := time.Now()

	candidates, err := r.fetcher.FetchAll(ctx)
	if err != nil {
		return err
	}
	r.log.InfoContext(ctx, "fetched proxy candidates", "count", len(candidates))

	working := r.validateAll(ctx, candidates)

	members := make([]string, len(working))
	for i, a := range working {
		members[i] = a.String()
	}
	if err := r.writer.Replace(ctx, members, r.poolTTL); err != nil {
		return err
	}

	r.mu.Lock()
	r.stats = Stats{
		LastRunAt:     start,
		Fetched:       len(candidates),
		Working:       len(working),
		CycleDuration: time.Since(start).String(),
		HasRun:        true,
	}
	r.mu.Unlock()

	r.log.InfoContext(ctx, "validation cycle complete",
		"fetched", len(candidates), "working", len(working), "took", time.Since(start))
	return nil
}

// validateAll validates candidates with bounded concurrency, inserting a
// politeness delay before each dispatch and adapting the delay to outcomes.
func (r *Runner) validateAll(ctx context.Context, candidates []proxy.Addr) []proxy.Addr {
	sem := make(chan struct{}, r.concurrency)
	results := make(chan proxy.Addr, len(candidates))
	var wg sync.WaitGroup

dispatch:
	for _, addr := range candidates {
		// Politeness: pace dispatch so we never burst the test endpoint.
		if err := r.limiter.Wait(ctx); err != nil {
			break
		}

		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			break dispatch
		}

		wg.Add(1)
		go func(a proxy.Addr) {
			defer wg.Done()
			defer func() { <-sem }()

			if err := r.checker.Validate(ctx, a); err != nil {
				r.limiter.Failure()
				return
			}
			r.limiter.Success()
			results <- a
		}(addr)
	}

	wg.Wait()
	close(results)

	working := make([]proxy.Addr, 0, len(candidates))
	for a := range results {
		working = append(working, a)
	}
	return working
}

// Loop runs RunOnce immediately and then every interval until ctx is cancelled.
// Cycle errors are logged but never stop the loop (NFR-2, resilience).
func (r *Runner) Loop(ctx context.Context, interval time.Duration) {
	r.runAndLog(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			r.log.Info("runner loop stopping")
			return
		case <-ticker.C:
			r.runAndLog(ctx)
		}
	}
}

func (r *Runner) runAndLog(ctx context.Context) {
	if err := r.RunOnce(ctx); err != nil && ctx.Err() == nil {
		r.log.ErrorContext(ctx, "validation cycle failed", "error", err)
	}
}
