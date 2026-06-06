package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Vislaren/MergeMarket/services/bff/internal/config"
)

// fakeUpstream stands in for the mock server / real services. It serves the
// handful of contract endpoints the BFF reads or forwards to. product_id
// "missing" yields a 404 history so the not-found path can be exercised.
func fakeUpstream() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/products/{id}/history", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") == "missing" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"not_found","message":"no such product"}`))
			return
		}
		_, _ = w.Write([]byte(`{
			"product_id":"prod-001","title":"iPhone 15",
			"history":[{"price":700000,"currency":"XAF","recorded_at":"2026-01-01T00:00:00Z"}],
			"average_6m":750000,"lowest_30d":690000}`))
	})

	mux.HandleFunc("GET /api/v1/products/{id}/truth-score", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"product_id":"prod-001","score":82,
			"sentiment":"positive","fake_review_risk":"low","summary":"Solid reviews."}`))
	})

	mux.HandleFunc("GET /api/v1/search", func(w http.ResponseWriter, _ *http.Request) {
		// Two offers, deliberately out of price order so we can assert sorting.
		_, _ = w.Write([]byte(`{"query":"iPhone 15","results":[
			{"product_id":"prod-001","title":"iPhone 15","price":720000,"currency":"XAF",
			 "shipping":5000,"total_cost":725000,"store":"StoreB","deal_score":60},
			{"product_id":"prod-001","title":"iPhone 15","price":700000,"currency":"XAF",
			 "shipping":2000,"total_cost":702000,"store":"StoreA","deal_score":88}
		],"cached":false,"latency_ms":120}`))
	})

	mux.HandleFunc("GET /api/v1/alerts", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"alerts":[{"alert_id":"al-1"}]}`))
	})

	return httptest.NewServer(mux)
}

func newTestServer(t *testing.T, upstreamURL string) http.Handler {
	t.Helper()
	cfg := config.Config{Addr: ":0", Port: 0, UpstreamURL: upstreamURL}
	srv, err := New(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return srv.Handler()
}

func TestHealth(t *testing.T) {
	up := fakeUpstream()
	defer up.Close()
	h := newTestServer(t, up.URL)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["service"] != "bff" || body["status"] != "ok" {
		t.Errorf("health body = %v", body)
	}
}

func TestMetricsCountsRequests(t *testing.T) {
	up := fakeUpstream()
	defer up.Close()
	h := newTestServer(t, up.URL)

	// One health hit, then read metrics.
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health", nil))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "bff_requests_total") {
		t.Errorf("metrics missing counter: %s", rec.Body.String())
	}
}

func TestProductDetailAggregates(t *testing.T) {
	up := fakeUpstream()
	defer up.Close()
	h := newTestServer(t, up.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/prod-001/detail", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%s)", rec.Code, rec.Body.String())
	}
	var detail ProductDetail
	if err := json.Unmarshal(rec.Body.Bytes(), &detail); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if detail.Title != "iPhone 15" {
		t.Errorf("title = %q, want iPhone 15", detail.Title)
	}
	if detail.TruthScore.Score != 82 {
		t.Errorf("truth score = %d, want 82", detail.TruthScore.Score)
	}
	if len(detail.Offers) != 2 {
		t.Fatalf("offers = %d, want 2", len(detail.Offers))
	}
	// Offers sorted by total cost ascending; cheapest is StoreA.
	if detail.Offers[0].Store != "StoreA" {
		t.Errorf("offers not sorted: first = %q, want StoreA", detail.Offers[0].Store)
	}
	if detail.BestOffer == nil || detail.BestOffer.Store != "StoreA" {
		t.Errorf("best offer = %v, want StoreA", detail.BestOffer)
	}
	if detail.DealScore != 88 {
		t.Errorf("deal score = %d, want 88 (cheapest offer)", detail.DealScore)
	}
	if detail.StoreCount != 2 {
		t.Errorf("store count = %d, want 2", detail.StoreCount)
	}
}

func TestProductDetailNotFound(t *testing.T) {
	up := fakeUpstream()
	defer up.Close()
	h := newTestServer(t, up.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/missing/detail", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["error"] != "not_found" {
		t.Errorf("error = %q, want not_found", body["error"])
	}
}

func TestForwardsUnknownRoutesToUpstream(t *testing.T) {
	up := fakeUpstream()
	defer up.Close()
	h := newTestServer(t, up.URL)

	// /api/v1/alerts has no BFF handler — it must be reverse-proxied upstream.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "al-1") {
		t.Errorf("forwarded body = %s, want upstream alerts", rec.Body.String())
	}
}
