package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Vislaren/MergeMarket/services/proxy-validator/internal/runner"
)

type fakeStats struct{ s runner.Stats }

func (f fakeStats) Stats() runner.Stats { return f.s }

func TestHealthEndpoint(t *testing.T) {
	srv := New(":0", "1.2.3", fakeStats{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body healthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Status != "ok" || body.Service != serviceName || body.Version != "1.2.3" {
		t.Errorf("health body = %+v", body)
	}
}

func TestStatsEndpoint(t *testing.T) {
	now := time.Now()
	srv := New(":0", "1.2.3", fakeStats{s: runner.Stats{
		LastRunAt: now, Fetched: 10, Working: 4, CycleDuration: "1s", HasRun: true,
	}})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var got runner.Stats
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got.Fetched != 10 || got.Working != 4 || !got.HasRun {
		t.Errorf("stats body = %+v", got)
	}
}
