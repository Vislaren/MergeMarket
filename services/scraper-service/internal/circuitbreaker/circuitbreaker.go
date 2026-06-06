// Package circuitbreaker implements the Circuit Breaker pattern used by the
// scraper to stop hammering a store that is actively blocking us (consecutive
// 403/429 responses). A store's breaker trips open after a threshold of
// consecutive failures, rejects requests for a cooldown window, then allows a
// single half-open trial before deciding to close or re-open.
package circuitbreaker

import (
	"sync"
	"time"
)

// State is the breaker's current disposition.
type State string

const (
	// StateClosed means requests flow normally.
	StateClosed State = "closed"
	// StateOpen means requests are rejected (the store is blocking us).
	StateOpen State = "open"
	// StateHalfOpen means a single trial request is permitted to probe recovery.
	StateHalfOpen State = "half-open"
)

// Breaker is a single store's circuit breaker. It is safe for concurrent use.
type Breaker struct {
	threshold int
	cooldown  time.Duration
	now       func() time.Time // injectable clock for tests

	mu               sync.Mutex
	state            State
	consecutiveFails int
	openedAt         time.Time
	halfOpenInFlight bool
}

// Option customises a Breaker.
type Option func(*Breaker)

// WithClock overrides the time source (used in tests for determinism).
func WithClock(now func() time.Time) Option {
	return func(b *Breaker) { b.now = now }
}

// New returns a closed breaker that trips after threshold consecutive failures
// and stays open for cooldown before permitting a half-open trial. A threshold
// below 1 is clamped to 1.
func New(threshold int, cooldown time.Duration, opts ...Option) *Breaker {
	if threshold < 1 {
		threshold = 1
	}
	b := &Breaker{
		threshold: threshold,
		cooldown:  cooldown,
		now:       time.Now,
		state:     StateClosed,
	}
	for _, o := range opts {
		o(b)
	}
	return b
}

// Allow reports whether a request may proceed right now. When the breaker is
// open and the cooldown has elapsed it transitions to half-open and allows
// exactly one trial request through (subsequent calls are rejected until that
// trial reports its outcome via Success/Failure).
func (b *Breaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		return true
	case StateOpen:
		if b.now().Sub(b.openedAt) >= b.cooldown {
			b.state = StateHalfOpen
			b.halfOpenInFlight = true
			return true
		}
		return false
	case StateHalfOpen:
		// Only one trial request is permitted while half-open.
		if !b.halfOpenInFlight {
			b.halfOpenInFlight = true
			return true
		}
		return false
	default:
		return true
	}
}

// Success records a successful request. It closes the breaker and resets the
// consecutive-failure counter.
func (b *Breaker) Success() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.consecutiveFails = 0
	b.state = StateClosed
	b.halfOpenInFlight = false
}

// Failure records a blocking failure (a 403/429). In closed state it trips the
// breaker open once the threshold is reached; in half-open state it immediately
// re-opens and restarts the cooldown.
func (b *Breaker) Failure() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.state == StateHalfOpen {
		b.trip()
		return
	}
	b.consecutiveFails++
	if b.consecutiveFails >= b.threshold {
		b.trip()
	}
}

// trip moves the breaker to open and stamps the cooldown start. Caller holds mu.
func (b *Breaker) trip() {
	b.state = StateOpen
	b.openedAt = b.now()
	b.halfOpenInFlight = false
}

// State returns the breaker's current state.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

// Group manages one Breaker per store id, created lazily with shared settings.
// It is safe for concurrent use.
type Group struct {
	threshold int
	cooldown  time.Duration
	opts      []Option

	mu       sync.Mutex
	breakers map[string]*Breaker
}

// NewGroup returns a Group that hands out breakers configured with threshold and
// cooldown. opts (e.g. WithClock) are applied to every breaker it creates.
func NewGroup(threshold int, cooldown time.Duration, opts ...Option) *Group {
	return &Group{
		threshold: threshold,
		cooldown:  cooldown,
		opts:      opts,
		breakers:  make(map[string]*Breaker),
	}
}

// For returns the breaker for storeID, creating it on first use.
func (g *Group) For(storeID string) *Breaker {
	g.mu.Lock()
	defer g.mu.Unlock()
	b, ok := g.breakers[storeID]
	if !ok {
		b = New(g.threshold, g.cooldown, g.opts...)
		g.breakers[storeID] = b
	}
	return b
}
