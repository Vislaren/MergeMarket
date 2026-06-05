package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Vislaren/MergeMarket/services/normalization/internal/worker"
)

type fakeStats struct{ s worker.Stats }

func (f fakeStats) Stats() worker.Stats { return f.s }

func TestHealth(t *testing.T) {
	srv := New(":0", "9.9.9", fakeStats{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" || body["service"] != serviceName || body["version"] != "9.9.9" {
		t.Errorf("unexpected health body: %v", body)
	}
}

func TestStats(t *testing.T) {
	want := worker.Stats{ResultsProcessed: 3, ProductsWritten: 7, LastResultAt: time.Now().UTC()}
	srv := New(":0", "1", fakeStats{s: want})
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var got worker.Stats
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.ResultsProcessed != 3 || got.ProductsWritten != 7 {
		t.Errorf("unexpected stats body: %+v", got)
	}
}
