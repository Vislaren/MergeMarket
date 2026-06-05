// Package runner drives the history-service's two scheduled jobs:
//
//   - Snapshot: periodically (daily by default) records a price_history row for
//     every priced product, building the long-term time series.
//   - Heartbeat: periodically re-checks every followed (alerted) product's
//     current price, records a snapshot, and fires a price-drop alert when the
//     price crosses *down* through a threshold (so an alert fires once on the
//     drop, not repeatedly while the price stays low).
//
// Both jobs are resilient: a single product or alert failure is logged and the
// cycle continues (NFR-2). The single-cycle methods (Snapshot, Heartbeat) are
// exported so they can be unit-tested directly.
package runner

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/Vislaren/MergeMarket/services/history/internal/alert"
	"github.com/Vislaren/MergeMarket/services/history/internal/pricesource"
	"github.com/Vislaren/MergeMarket/services/history/internal/store"
)

// Stats is a snapshot of the runner's cumulative counters, surfaced via /stats.
type Stats struct {
	SnapshotRuns    int64     `json:"snapshot_runs"`
	RowsSnapshotted int64     `json:"rows_snapshotted"`
	HeartbeatRuns   int64     `json:"heartbeat_runs"`
	ProductsChecked int64     `json:"products_checked"`
	AlertsFired     int64     `json:"alerts_fired"`
	Errors          int64     `json:"errors"`
	LastSnapshotAt  time.Time `json:"last_snapshot_at"`
	LastHeartbeatAt time.Time `json:"last_heartbeat_at"`
}

// Runner orchestrates the snapshot and heartbeat jobs.
type Runner struct {
	repo        store.Repository
	prices      pricesource.Source
	publisher   alert.Publisher
	snapshotInt time.Duration
	heartbeatIn time.Duration
	log         *slog.Logger

	mu    sync.Mutex
	stats Stats
}

// New constructs a Runner.
func New(repo store.Repository, prices pricesource.Source, publisher alert.Publisher, snapshotInterval, heartbeatInterval time.Duration, log *slog.Logger) *Runner {
	if log == nil {
		log = slog.Default()
	}
	return &Runner{
		repo:        repo,
		prices:      prices,
		publisher:   publisher,
		snapshotInt: snapshotInterval,
		heartbeatIn: heartbeatInterval,
		log:         log,
	}
}

// Stats returns a copy of the current counters.
func (r *Runner) Stats() Stats {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stats
}

// Run starts the snapshot and heartbeat tickers and blocks until ctx is
// cancelled. The *OnStart flags trigger one immediate cycle of each job before
// the first tick.
func (r *Runner) Run(ctx context.Context, snapshotOnStart, heartbeatOnStart bool) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); r.loop(ctx, r.snapshotInt, snapshotOnStart, r.runSnapshot) }()
	go func() { defer wg.Done(); r.loop(ctx, r.heartbeatIn, heartbeatOnStart, r.runHeartbeat) }()
	wg.Wait()
	r.log.Info("runner stopped")
}

// loop ticks every interval, invoking job, until ctx is cancelled. When runNow
// is true it invokes job once immediately.
func (r *Runner) loop(ctx context.Context, interval time.Duration, runNow bool, job func(context.Context)) {
	if runNow {
		job(ctx)
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			job(ctx)
		}
	}
}

func (r *Runner) runSnapshot(ctx context.Context) {
	if _, err := r.Snapshot(ctx); err != nil {
		r.log.Error("snapshot cycle failed", "error", err)
	}
}

func (r *Runner) runHeartbeat(ctx context.Context) {
	if err := r.Heartbeat(ctx); err != nil {
		r.log.Error("heartbeat cycle failed", "error", err)
	}
}

// Snapshot runs one full daily-snapshot cycle and returns the rows written.
func (r *Runner) Snapshot(ctx context.Context) (int64, error) {
	n, err := r.repo.SnapshotAll(ctx)
	if err != nil {
		r.mark(func(s *Stats) { s.Errors++ })
		return 0, err
	}
	r.mark(func(s *Stats) {
		s.SnapshotRuns++
		s.RowsSnapshotted += n
		s.LastSnapshotAt = time.Now().UTC()
	})
	r.log.Info("price snapshot recorded", "rows", n)
	return n, nil
}

// Heartbeat runs one heartbeat cycle: for every followed product it resolves the
// current price, records a snapshot, and fires alerts on a downward threshold
// crossing.
func (r *Runner) Heartbeat(ctx context.Context) error {
	followed, err := r.repo.FollowedProducts(ctx)
	if err != nil {
		r.mark(func(s *Stats) { s.Errors++ })
		return err
	}

	now := time.Now().UTC()
	for _, f := range followed {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		r.checkProduct(ctx, f, now)
	}

	r.mark(func(s *Stats) {
		s.HeartbeatRuns++
		s.ProductsChecked += int64(len(followed))
		s.LastHeartbeatAt = now
	})
	r.log.Info("heartbeat cycle complete", "followed", len(followed))
	return nil
}

// checkProduct handles one followed product: read current price, capture the
// previous recorded price (for crossing detection), record a snapshot, then
// evaluate each alert.
func (r *Runner) checkProduct(ctx context.Context, f store.Followed, now time.Time) {
	reading, err := r.prices.CurrentPrice(ctx, f)
	if err != nil {
		r.mark(func(s *Stats) { s.Errors++ })
		r.log.Warn("price check failed", "product_id", f.ProductID, "error", err)
		return
	}
	if !reading.OK {
		r.log.Debug("no current price; skipping", "product_id", f.ProductID)
		return
	}

	// Capture the previous recorded price BEFORE inserting this snapshot so the
	// crossing test compares against the prior observation.
	prev, hadPrev, err := r.repo.LatestPrice(ctx, f.ProductID)
	if err != nil {
		r.mark(func(s *Stats) { s.Errors++ })
		r.log.Warn("latest price lookup failed", "product_id", f.ProductID, "error", err)
		return
	}

	if err := r.repo.InsertSnapshot(ctx, f.ProductID, reading.Price, reading.Shipping, f.Currency, now); err != nil {
		r.mark(func(s *Stats) { s.Errors++ })
		r.log.Warn("insert snapshot failed", "product_id", f.ProductID, "error", err)
		return
	}

	for _, a := range f.Alerts {
		if crossedDown(reading.Price, prev, hadPrev, a.Threshold) {
			r.fire(ctx, f, a, reading.Price, now)
		}
	}
}

// crossedDown reports whether current is at/below threshold while the previous
// observation was strictly above it (or there was no previous observation). This
// makes an alert fire once on the drop rather than every cycle the price stays
// low.
func crossedDown(current, prev float64, hadPrev bool, threshold float64) bool {
	if current > threshold {
		return false
	}
	if !hadPrev {
		return true
	}
	return prev > threshold
}

// fire publishes a price-drop alert event.
func (r *Runner) fire(ctx context.Context, f store.Followed, a store.Alert, price float64, now time.Time) {
	currency := a.Currency
	if currency == "" {
		currency = f.Currency
	}
	ev := alert.Event{
		AlertID:     a.ID,
		ProductID:   f.ProductID,
		Title:       f.Title,
		Store:       f.Store,
		Price:       price,
		Threshold:   a.Threshold,
		Currency:    currency,
		TriggeredAt: now,
	}
	if err := r.publisher.Publish(ctx, ev); err != nil {
		r.mark(func(s *Stats) { s.Errors++ })
		r.log.Error("publish alert failed", "alert_id", a.ID, "product_id", f.ProductID, "error", err)
		return
	}
	r.mark(func(s *Stats) { s.AlertsFired++ })
	r.log.Info("price-drop alert fired",
		"alert_id", a.ID, "product_id", f.ProductID, "price", price, "threshold", a.Threshold)
}

func (r *Runner) mark(fn func(*Stats)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	fn(&r.stats)
}
