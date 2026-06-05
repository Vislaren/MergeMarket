package proxypool

import (
	"context"
	"testing"
)

func TestStaticProvider(t *testing.T) {
	s := Static{Addr: "1.2.3.4:8080"}
	got, err := s.Pick(context.Background())
	if err != nil {
		t.Fatalf("Pick error: %v", err)
	}
	if got != "1.2.3.4:8080" {
		t.Errorf("Pick = %q, want 1.2.3.4:8080", got)
	}
}

func TestStaticProviderEmpty(t *testing.T) {
	got, err := Static{}.Pick(context.Background())
	if err != nil {
		t.Fatalf("Pick error: %v", err)
	}
	if got != "" {
		t.Errorf("Pick = %q, want empty (direct connection)", got)
	}
}
