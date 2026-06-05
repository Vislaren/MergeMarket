package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/worker"
)

type fakeStats struct{ s worker.Stats }

func (f fakeStats) Stats() worker.Stats { return f.s }

func TestHealthEndpoint(t *testing.T) {
	srv := New(":0", "1.2.3", fakeStats{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body healthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Status != "ok" || body.Service != "scraper-service" || body.Version != "1.2.3" {
		t.Errorf("health body = %+v", body)
	}
}

func TestStatsEndpoint(t *testing.T) {
	srv := New(":0", "1.2.3", fakeStats{s: worker.Stats{Processed: 7, Succeeded: 5, Blocked: 2}})
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var got worker.Stats
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Processed != 7 || got.Succeeded != 5 || got.Blocked != 2 {
		t.Errorf("stats body = %+v", got)
	}
}
