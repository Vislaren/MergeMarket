package queue

import (
	"context"
	"errors"
	"testing"
)

func TestMemorySource_FIFOThenEmpty(t *testing.T) {
	src := NewMemory(
		RawResult{StoreID: "a"},
		RawResult{StoreID: "b"},
	)
	if src.Len() != 2 {
		t.Fatalf("len = %d", src.Len())
	}

	r1, err := src.Dequeue(context.Background())
	if err != nil || r1.StoreID != "a" {
		t.Fatalf("first dequeue: %v %q", err, r1.StoreID)
	}
	r2, _ := src.Dequeue(context.Background())
	if r2.StoreID != "b" {
		t.Fatalf("second dequeue: %q", r2.StoreID)
	}

	if _, err := src.Dequeue(context.Background()); !errors.Is(err, ErrEmpty) {
		t.Fatalf("expected ErrEmpty, got %v", err)
	}
}

func TestMemorySource_ContextCancelled(t *testing.T) {
	src := NewMemory(RawResult{StoreID: "a"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := src.Dequeue(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
