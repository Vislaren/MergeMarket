package politeness

import (
	"context"
	"testing"
	"time"
)

func TestNextDelayWithinWindow(t *testing.T) {
	l := New(10*time.Millisecond, 30*time.Millisecond, 42)
	for i := 0; i < 100; i++ {
		d := l.nextDelay()
		if d < 10*time.Millisecond || d > 30*time.Millisecond {
			t.Fatalf("delay %s outside [10ms,30ms] at baseline factor", d)
		}
	}
}

func TestFailureBacksOffSuccessRecovers(t *testing.T) {
	l := New(time.Millisecond, time.Millisecond, 1)
	if got := l.Factor(); got != 1.0 {
		t.Fatalf("initial factor = %v, want 1.0", got)
	}

	l.Failure()
	l.Failure()
	if l.Factor() <= 1.0 {
		t.Fatalf("factor should grow after failures, got %v", l.Factor())
	}

	// Backoff is capped.
	for i := 0; i < 100; i++ {
		l.Failure()
	}
	if l.Factor() > maxBackoffFactor {
		t.Fatalf("factor %v exceeded cap %v", l.Factor(), maxBackoffFactor)
	}

	// Successes drive it back toward 1.0 but never below.
	for i := 0; i < 1000; i++ {
		l.Success()
	}
	if l.Factor() != 1.0 {
		t.Fatalf("factor after many successes = %v, want 1.0", l.Factor())
	}
}

func TestWaitRespectsContextCancellation(t *testing.T) {
	l := New(time.Hour, time.Hour, 1) // would block ~1h if not cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := l.Wait(ctx); err == nil {
		t.Fatal("Wait should return ctx error when context is already cancelled")
	}
}

func TestWaitReturnsAfterDelay(t *testing.T) {
	l := New(time.Millisecond, time.Millisecond, 1)
	start := time.Now()
	if err := l.Wait(context.Background()); err != nil {
		t.Fatalf("Wait error = %v", err)
	}
	if time.Since(start) < time.Millisecond {
		t.Fatal("Wait returned before the minimum delay elapsed")
	}
}
