// Package politeness implements the adaptive random-delay protocol used to
// avoid hammering proxy sources and test endpoints. The delay is randomised
// within a configured window and adapts: it backs off (grows) after failures
// and relaxes (shrinks) after successes, keeping the validator a polite,
// well-behaved client.
package politeness

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

const (
	// maxBackoffFactor caps how far failures can stretch the delay window.
	maxBackoffFactor = 8.0
	// backoffStep is added to the factor on each failure.
	backoffStep = 0.5
	// recoverStep is subtracted from the factor on each success.
	recoverStep = 0.25
)

// Limiter produces adaptive random delays. It is safe for concurrent use.
type Limiter struct {
	min, max time.Duration

	mu     sync.Mutex
	factor float64
	rng    *rand.Rand
}

// New returns a Limiter that waits a random duration within [min, max],
// scaled by an adaptive factor that starts at 1.0. If seed is 0 the current
// time is used, which is the right choice outside of tests.
func New(min, max time.Duration, seed int64) *Limiter {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	if max < min {
		max = min
	}
	return &Limiter{
		min:    min,
		max:    max,
		factor: 1.0,
		rng:    rand.New(rand.NewSource(seed)),
	}
}

// nextDelay computes (without sleeping) the next delay, applying the current
// adaptive factor. Exposed for deterministic testing.
func (l *Limiter) nextDelay() time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()

	span := l.max - l.min
	var base time.Duration
	if span > 0 {
		base = l.min + time.Duration(l.rng.Int63n(int64(span)+1))
	} else {
		base = l.min
	}
	return time.Duration(float64(base) * l.factor)
}

// Wait blocks for the next adaptive random delay, or until ctx is cancelled.
// It returns ctx.Err() if the context ends first.
func (l *Limiter) Wait(ctx context.Context) error {
	d := l.nextDelay()
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// Failure widens the delay window (exponential-style backoff) up to the cap.
func (l *Limiter) Failure() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.factor += backoffStep
	if l.factor > maxBackoffFactor {
		l.factor = maxBackoffFactor
	}
}

// Success narrows the delay window back toward the baseline factor of 1.0.
func (l *Limiter) Success() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.factor -= recoverStep
	if l.factor < 1.0 {
		l.factor = 1.0
	}
}

// Factor reports the current adaptive multiplier (primarily for observability
// and tests).
func (l *Limiter) Factor() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.factor
}
