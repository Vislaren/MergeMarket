package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vislaren/MergeMarket/services/history/internal/runner"
	"github.com/Vislaren/MergeMarket/services/history/internal/store"
)

type fakeHist struct {
	res store.HistoryResult
	err error
}

func (f fakeHist) History(context.Context, string) (store.HistoryResult, error) {
	return f.res, f.err
}

type fakeStats struct{ s runner.Stats }

func (f fakeStats) Stats() runner.Stats { return f.s }

func TestHistoryHandler_OK(t *testing.T) {
	h := fakeHist{res: store.HistoryResult{
		ProductID: "p1", Title: "Phone",
		History:   []store.HistoryPoint{{Price: 10, Currency: "USD"}},
		Average6m: 12.5, Lowest30d: 9,
	}}
	srv := New(":0", "1", h, fakeStats{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/p1/history", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var got store.HistoryResult
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.ProductID != "p1" || got.Average6m != 12.5 || len(got.History) != 1 {
		t.Errorf("unexpected body: %+v", got)
	}
}

func TestHistoryHandler_NotFound(t *testing.T) {
	srv := New(":0", "1", fakeHist{err: store.ErrNotFound}, fakeStats{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/missing/history", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["error"] != "not_found" {
		t.Errorf("error code = %q", body["error"])
	}
}

func TestHealth(t *testing.T) {
	srv := New(":0", "2.0", fakeHist{}, fakeStats{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["status"] != "ok" || body["service"] != serviceName || body["version"] != "2.0" {
		t.Errorf("unexpected health body: %v", body)
	}
}
