package config_test

import (
	"testing"

	"github.com/Vislaren/MergeMarket/services/mock-server/internal/config"
)

func TestLoad(t *testing.T) {
	t.Run("default port when unset", func(t *testing.T) {
		t.Setenv("MOCK_SERVER_PORT", "")
		t.Setenv("PORT", "")
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Addr != ":8080" || cfg.Port != 8080 {
			t.Fatalf("got addr=%q port=%d, want :8080/8080", cfg.Addr, cfg.Port)
		}
	})

	t.Run("MOCK_SERVER_PORT wins over PORT", func(t *testing.T) {
		t.Setenv("MOCK_SERVER_PORT", "9090")
		t.Setenv("PORT", "7000")
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Port != 9090 {
			t.Fatalf("got port=%d, want 9090", cfg.Port)
		}
	})

	t.Run("falls back to PORT when MOCK_SERVER_PORT unset", func(t *testing.T) {
		t.Setenv("MOCK_SERVER_PORT", "")
		t.Setenv("PORT", "7000")
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Port != 7000 {
			t.Fatalf("got port=%d, want 7000", cfg.Port)
		}
	})

	t.Run("invalid port is rejected", func(t *testing.T) {
		t.Setenv("MOCK_SERVER_PORT", "notanumber")
		if _, err := config.Load(); err == nil {
			t.Fatal("expected error for non-numeric port, got nil")
		}
	})

	t.Run("out-of-range port is rejected", func(t *testing.T) {
		t.Setenv("MOCK_SERVER_PORT", "70000")
		if _, err := config.Load(); err == nil {
			t.Fatal("expected error for out-of-range port, got nil")
		}
	})
}
