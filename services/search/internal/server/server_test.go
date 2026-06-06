package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vislaren/MergeMarket/services/search/internal/search"
	"github.com/Vislaren/MergeMarket/services/search/internal/store"
)

type fakeSearcher struct {
	resp search.Response
	err  error
}

func (f fakeSearcher) Search(_ context.Context, query, _ string) (search.Response, error) {
	if f.err != nil {
		return search.Response{}, f.err
	}
	r := f.resp
	r.Query = query
	return r, nil
}

func do(t *testing.T, h http.Handler, target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestSearchMissingQuery(t *testing.T) {
	srv := New(":0", "test", fakeSearcher{})
	rec := do(t, srv.Handler, "/api/v1/search?location=CM")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", rec.Code)
	}
	var body errorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Error != "missing_query" {
		t.Errorf("error = %q, want missing_query", body.Error)
	}
}

func TestSearchMissingLocation(t *testing.T) {
	srv := New(":0", "test", fakeSearcher{})
	rec := do(t, srv.Handler, "/api/v1/search?q=phone")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", rec.Code)
	}
}

func TestSearchSuccess(t *testing.T) {
	svc := fakeSearcher{resp: search.Response{
		Results:   []store.Product{{ProductID: "a", Title: "Phone", Price: 100, DealScore: 100}},
		Cached:    true,
		LatencyMs: 3,
	}}
	srv := New(":0", "test", svc)
	rec := do(t, srv.Handler, "/api/v1/search?q=phone&location=CM")
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", rec.Code)
	}
	var body search.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Query != "phone" || len(body.Results) != 1 || !body.Cached {
		t.Errorf("unexpected body: %+v", body)
	}
}

func TestHealth(t *testing.T) {
	srv := New(":0", "9.9.9", fakeSearcher{})
	rec := do(t, srv.Handler, "/health")
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", rec.Code)
	}
	var body healthResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Status != "ok" || body.Service != serviceName || body.Version != "9.9.9" {
		t.Errorf("unexpected health body: %+v", body)
	}
}
