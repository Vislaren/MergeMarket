package runner

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/Vislaren/MergeMarket/services/history/internal/alert"
	"github.com/Vislaren/MergeMarket/services/history/internal/pricesource"
	"github.com/Vislaren/MergeMarket/services/history/internal/store"
)

// --- fakes -----------------------------------------------------------------

type fakeRepo struct {
	mu         sync.Mutex
	followed   []store.Followed
	latest     map[string]float64 // product_id → latest recorded price
	snapshots  []snap
	snapshotN  int64
	historyRes store.HistoryResult
	historyErr error
}

type snap struct {
	productID string
	price     float64
}

func (f *fakeRepo) SnapshotAll(context.Context) (int64, error) { return f.snapshotN, nil }

func (f *fakeRepo) FollowedProducts(context.Context) ([]store.Followed, error) {
	return f.followed, nil
}

func (f *fakeRepo) LatestPrice(_ context.Context, productID string) (float64, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.latest[productID]
	return v, ok, nil
}

func (f *fakeRepo) InsertSnapshot(_ context.Context, productID string, price float64, _ *float64, _ string, _ time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.snapshots = append(f.snapshots, snap{productID, price})
	if f.latest == nil {
		f.latest = map[string]float64{}
	}
	f.latest[productID] = price // subsequent cycles see this as the previous price
	return nil
}

func (f *fakeRepo) History(context.Context, string) (store.HistoryResult, error) {
	return f.historyRes, f.historyErr
}

func (f *fakeRepo) Close() error { return nil }

// fixedSource returns the same price for every product.
type fixedSource struct{ price float64 }

func (s fixedSource) CurrentPrice(context.Context, store.Followed) (pricesource.Reading, error) {
	return pricesource.Reading{Price: s.price, OK: true}, nil
}

type capturePublisher struct {
	mu     sync.Mutex
	events []alert.Event
}

func (c *capturePublisher) Publish(_ context.Context, e alert.Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, e)
	return nil
}

func (c *capturePublisher) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.events)
}

func quiet() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

// --- tests -----------------------------------------------------------------

func TestHeartbeat_FiresOnDownwardCrossingOnce(t *testing.T) {
	repo := &fakeRepo{
		followed: []store.Followed{{
			ProductID: "p1", Title: "Phone", Store: "Jumia", Currency: "USD",
			Alerts: []store.Alert{{ID: "a1", Threshold: 100, Currency: "USD"}},
		}},
		latest: map[string]float64{"p1": 120}, // previous recorded price above threshold
	}
	pub := &capturePublisher{}
	// Current price 90 → below threshold 100, previous 120 above → crossing.
	r := New(repo, fixedSource{price: 90}, pub, time.Hour, time.Hour, quiet())

	if err := r.Heartbeat(context.Background()); err != nil {
		t.Fatal(err)
	}
	if pub.count() != 1 {
		t.Fatalf("expected 1 alert on crossing, got %d", pub.count())
	}
	ev := pub.events[0]
	if ev.AlertID != "a1" || ev.Price != 90 || ev.Threshold != 100 {
		t.Errorf("unexpected event: %+v", ev)
	}

	// Second cycle: price still 90 (now the previous recorded price is also 90,
	// i.e. below threshold) → must NOT fire again.
	if err := r.Heartbeat(context.Background()); err != nil {
		t.Fatal(err)
	}
	if pub.count() != 1 {
		t.Errorf("alert re-fired while price stayed low: count=%d", pub.count())
	}
}

func TestHeartbeat_NoFireWhenAboveThreshold(t *testing.T) {
	repo := &fakeRepo{
		followed: []store.Followed{{
			ProductID: "p1", Currency: "USD",
			Alerts: []store.Alert{{ID: "a1", Threshold: 50}},
		}},
		latest: map[string]float64{"p1": 80},
	}
	pub := &capturePublisher{}
	r := New(repo, fixedSource{price: 70}, pub, time.Hour, time.Hour, quiet())

	if err := r.Heartbeat(context.Background()); err != nil {
		t.Fatal(err)
	}
	if pub.count() != 0 {
		t.Errorf("alert fired above threshold: %d", pub.count())
	}
	// A snapshot is still recorded for the followed product.
	if len(repo.snapshots) != 1 || repo.snapshots[0].price != 70 {
		t.Errorf("expected one snapshot at 70, got %+v", repo.snapshots)
	}
}

func TestHeartbeat_FiresWhenNoPreviousPrice(t *testing.T) {
	repo := &fakeRepo{
		followed: []store.Followed{{
			ProductID: "p1", Currency: "USD",
			Alerts: []store.Alert{{ID: "a1", Threshold: 100}},
		}},
		latest: map[string]float64{}, // no prior snapshot
	}
	pub := &capturePublisher{}
	r := New(repo, fixedSource{price: 80}, pub, time.Hour, time.Hour, quiet())

	if err := r.Heartbeat(context.Background()); err != nil {
		t.Fatal(err)
	}
	if pub.count() != 1 {
		t.Errorf("expected alert with no previous price, got %d", pub.count())
	}
}

func TestHeartbeat_SkipsWhenNoPrice(t *testing.T) {
	repo := &fakeRepo{
		followed: []store.Followed{{ProductID: "p1", Alerts: []store.Alert{{ID: "a1", Threshold: 100}}}},
	}
	pub := &capturePublisher{}
	// DBSource with HasPrice=false → no reading.
	r := New(repo, pricesource.DBSource{}, pub, time.Hour, time.Hour, quiet())

	if err := r.Heartbeat(context.Background()); err != nil {
		t.Fatal(err)
	}
	if pub.count() != 0 || len(repo.snapshots) != 0 {
		t.Errorf("expected no alert and no snapshot when price unknown")
	}
	st := r.Stats()
	if st.ProductsChecked != 1 || st.HeartbeatRuns != 1 {
		t.Errorf("unexpected stats: %+v", st)
	}
}

func TestSnapshot_CountsRows(t *testing.T) {
	repo := &fakeRepo{snapshotN: 42}
	r := New(repo, fixedSource{price: 1}, &capturePublisher{}, time.Hour, time.Hour, quiet())
	n, err := r.Snapshot(context.Background())
	if err != nil || n != 42 {
		t.Fatalf("snapshot: n=%d err=%v", n, err)
	}
	if r.Stats().RowsSnapshotted != 42 {
		t.Errorf("rows not counted: %+v", r.Stats())
	}
}

func TestCrossedDown(t *testing.T) {
	cases := []struct {
		cur, prev float64
		had       bool
		thr       float64
		want      bool
	}{
		{90, 120, true, 100, true},   // crossed down
		{90, 95, true, 100, false},   // already below last cycle
		{110, 120, true, 100, false}, // above threshold
		{90, 0, false, 100, true},    // no previous → fire
		{100, 120, true, 100, true},  // exactly at threshold counts as drop
	}
	for i, c := range cases {
		if got := crossedDown(c.cur, c.prev, c.had, c.thr); got != c.want {
			t.Errorf("case %d: crossedDown(%v,%v,%v,%v)=%v want %v", i, c.cur, c.prev, c.had, c.thr, got, c.want)
		}
	}
}
