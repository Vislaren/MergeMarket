package queue

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryQueueFIFO(t *testing.T) {
	q := NewMemoryQueue(
		Job{StoreID: "a", Query: "x"},
		Job{StoreID: "b", Query: "y"},
	)
	q.Enqueue(Job{StoreID: "c", Query: "z"})

	if q.Len() != 3 {
		t.Fatalf("Len = %d, want 3", q.Len())
	}

	ctx := context.Background()
	var got []string
	for {
		j, err := q.Dequeue(ctx)
		if errors.Is(err, ErrEmpty) {
			break
		}
		if err != nil {
			t.Fatalf("Dequeue error: %v", err)
		}
		got = append(got, j.StoreID)
	}
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("dequeued %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestMemoryQueueEmpty(t *testing.T) {
	q := NewMemoryQueue()
	if _, err := q.Dequeue(context.Background()); !errors.Is(err, ErrEmpty) {
		t.Fatalf("Dequeue on empty = %v, want ErrEmpty", err)
	}
}

func TestMemoryQueueContextCancelled(t *testing.T) {
	q := NewMemoryQueue(Job{StoreID: "a"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := q.Dequeue(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Dequeue with cancelled ctx = %v, want context.Canceled", err)
	}
}

func TestMemoryQueuePublish(t *testing.T) {
	q := NewMemoryQueue()
	r := RawResult{StoreID: "a", ScrapedAt: time.Now(), Products: []RawProduct{{Title: "t", Price: 1}}}
	if err := q.Publish(context.Background(), r); err != nil {
		t.Fatalf("Publish error: %v", err)
	}
	if len(q.Published) != 1 || q.Published[0].StoreID != "a" {
		t.Fatalf("Published = %+v", q.Published)
	}
}
