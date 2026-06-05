package config

import (
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	for _, k := range []string{
		"HISTORY_PORT", "DATABASE_URL", "HISTORY_ALERT_QUEUE",
		"HISTORY_SNAPSHOT_INTERVAL", "HISTORY_HEARTBEAT_INTERVAL",
		"HISTORY_HEARTBEAT_MODE", "DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD",
	} {
		t.Setenv(k, "")
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != "8085" {
		t.Errorf("port = %q, want 8085", cfg.Port)
	}
	if cfg.SnapshotInterval != 24*time.Hour {
		t.Errorf("snapshot interval = %v", cfg.SnapshotInterval)
	}
	if cfg.HeartbeatInterval != time.Hour {
		t.Errorf("heartbeat interval = %v", cfg.HeartbeatInterval)
	}
	if cfg.HeartbeatMode != ModeDB {
		t.Errorf("mode = %q, want db", cfg.HeartbeatMode)
	}
	if !cfg.HeartbeatOnStart {
		t.Errorf("heartbeat on start should default true")
	}
	if cfg.SnapshotOnStart {
		t.Errorf("snapshot on start should default false")
	}
}

func TestLoad_RejectsBadMode(t *testing.T) {
	t.Setenv("HISTORY_HEARTBEAT_MODE", "carrier-pigeon")
	if _, err := Load(); err == nil {
		t.Errorf("expected error for bad heartbeat mode")
	}
}

func TestLoad_RejectsBadInterval(t *testing.T) {
	t.Setenv("HISTORY_SNAPSHOT_INTERVAL", "soon")
	if _, err := Load(); err == nil {
		t.Errorf("expected error for bad interval")
	}
}

func TestLoad_HTTPMode(t *testing.T) {
	t.Setenv("HISTORY_HEARTBEAT_MODE", "http")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HeartbeatMode != ModeHTTP {
		t.Errorf("mode = %q, want http", cfg.HeartbeatMode)
	}
}
