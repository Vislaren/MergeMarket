package queue

import (
	"context"
	"sync"
)

// MemoryQueue is an in-memory Queue + Sink used for local dev and tests. It is
// safe for concurrent use. Dequeue returns ErrEmpty immediately when the queue
// is drained (it does not block), which keeps tests fast and deterministic.
type MemoryQueue struct {
	mu        sync.Mutex
	jobs      []Job
	Published []RawResult
}

// NewMemoryQueue returns an empty in-memory queue seeded with the given jobs.
func NewMemoryQueue(jobs ...Job) *MemoryQueue {
	return &MemoryQueue{jobs: append([]Job(nil), jobs...)}
}

// Enqueue appends a job to the in-memory queue.
func (m *MemoryQueue) Enqueue(j Job) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobs = append(m.jobs, j)
}

// Dequeue pops the next job, or returns ErrEmpty when none remain. It also
// honours a cancelled context.
func (m *MemoryQueue) Dequeue(ctx context.Context) (Job, error) {
	if err := ctx.Err(); err != nil {
		return Job{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.jobs) == 0 {
		return Job{}, ErrEmpty
	}
	j := m.jobs[0]
	m.jobs = m.jobs[1:]
	return j, nil
}

// Publish records a raw result.
func (m *MemoryQueue) Publish(_ context.Context, result RawResult) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Published = append(m.Published, result)
	return nil
}

// Len reports the number of jobs still queued.
func (m *MemoryQueue) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.jobs)
}

// Close is a no-op for the in-memory queue.
func (m *MemoryQueue) Close() error { return nil }
