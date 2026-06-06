package queue

import (
	"context"
	"sync"
)

// MemorySource is an in-memory Source used for tests. Results are served in the
// order they were added; once drained, Dequeue returns ErrEmpty (unless the
// context is cancelled first), mirroring the Redis poll-timeout behaviour.
type MemorySource struct {
	mu      sync.Mutex
	results []RawResult
	closed  bool
}

// NewMemory builds a MemorySource pre-seeded with results.
func NewMemory(results ...RawResult) *MemorySource {
	return &MemorySource{results: append([]RawResult(nil), results...)}
}

// Dequeue pops the next result, or returns ErrEmpty when none remain. It honours
// context cancellation.
func (m *MemorySource) Dequeue(ctx context.Context) (RawResult, error) {
	if err := ctx.Err(); err != nil {
		return RawResult{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.results) == 0 {
		return RawResult{}, ErrEmpty
	}
	r := m.results[0]
	m.results = m.results[1:]
	return r, nil
}

// Len reports how many results remain.
func (m *MemorySource) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.results)
}

// Close marks the source closed.
func (m *MemorySource) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}
