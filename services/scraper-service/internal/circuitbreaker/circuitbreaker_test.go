package circuitbreaker

import (
	"testing"
	"time"
)

// fakeClock is a manually-advanced clock for deterministic cooldown tests.
type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time      { return c.t }
func (c *fakeClock) add(d time.Duration) { c.t = c.t.Add(d) }

func TestTripsAfterThreshold(t *testing.T) {
	b := New(3, time.Minute)
	for i := 0; i < 2; i++ {
		b.Failure()
		if b.State() != StateClosed {
			t.Fatalf("after %d failures state=%s, want closed", i+1, b.State())
		}
		if !b.Allow() {
			t.Fatalf("Allow() should be true while closed")
		}
	}
	b.Failure() // third consecutive → trip
	if b.State() != StateOpen {
		t.Fatalf("state=%s, want open", b.State())
	}
	if b.Allow() {
		t.Fatalf("Allow() should be false while open within cooldown")
	}
}

func TestSuccessResetsCounter(t *testing.T) {
	b := New(3, time.Minute)
	b.Failure()
	b.Failure()
	b.Success() // resets the consecutive counter
	b.Failure()
	b.Failure()
	if b.State() != StateClosed {
		t.Fatalf("state=%s, want closed (success should have reset count)", b.State())
	}
}

func TestHalfOpenRecovery(t *testing.T) {
	clk := &fakeClock{t: time.Unix(1000, 0)}
	b := New(1, 30*time.Second, WithClock(clk.now))

	b.Failure() // threshold 1 → open immediately
	if b.State() != StateOpen || b.Allow() {
		t.Fatalf("expected open+reject, state=%s", b.State())
	}

	// Within cooldown: still rejected.
	clk.add(29 * time.Second)
	if b.Allow() {
		t.Fatalf("Allow() should still reject before cooldown elapses")
	}

	// Cooldown elapsed: one half-open trial allowed, second rejected.
	clk.add(2 * time.Second)
	if !b.Allow() {
		t.Fatalf("expected half-open trial to be allowed")
	}
	if b.State() != StateHalfOpen {
		t.Fatalf("state=%s, want half-open", b.State())
	}
	if b.Allow() {
		t.Fatalf("second concurrent half-open request should be rejected")
	}

	// Success during half-open closes the breaker.
	b.Success()
	if b.State() != StateClosed {
		t.Fatalf("state=%s, want closed after half-open success", b.State())
	}
}

func TestHalfOpenFailureReopens(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	b := New(1, 10*time.Second, WithClock(clk.now))
	b.Failure() // open
	clk.add(11 * time.Second)
	if !b.Allow() { // half-open trial
		t.Fatal("expected trial allowed")
	}
	b.Failure() // trial fails → reopen, restart cooldown
	if b.State() != StateOpen {
		t.Fatalf("state=%s, want open", b.State())
	}
	if b.Allow() {
		t.Fatal("should reject immediately after reopening")
	}
}

func TestThresholdClamped(t *testing.T) {
	b := New(0, time.Second) // clamps to 1
	b.Failure()
	if b.State() != StateOpen {
		t.Fatalf("state=%s, want open (threshold clamped to 1)", b.State())
	}
}

func TestGroupReturnsStablePerStore(t *testing.T) {
	g := NewGroup(2, time.Second)
	a1 := g.For("storeA")
	a2 := g.For("storeA")
	bStore := g.For("storeB")
	if a1 != a2 {
		t.Fatal("Group.For should return the same breaker for the same store")
	}
	if a1 == bStore {
		t.Fatal("different stores must get different breakers")
	}
}
