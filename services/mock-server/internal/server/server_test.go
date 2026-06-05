package server_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Vislaren/MergeMarket/services/mock-server/internal/config"
	"github.com/Vislaren/MergeMarket/services/mock-server/internal/fixtures"
	"github.com/Vislaren/MergeMarket/services/mock-server/internal/server"
)

// newTestServer returns a routed handler backed by a discarding logger.
func newTestServer() http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return server.New(config.Config{Addr: ":8080", Port: 8080}, logger).Handler()
}

// do executes a request against the handler and returns the recorder.
func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, r)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// decode unmarshals the response body into a generic map.
func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("body is not valid JSON: %v — %s", err, rec.Body.String())
	}
	return m
}

func TestHealth(t *testing.T) {
	h := newTestServer()
	rec := do(t, h, http.MethodGet, "/health", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	m := decode(t, rec)
	if m["status"] != "ok" || m["service"] != "mock-server" {
		t.Fatalf("unexpected health body: %v", m)
	}
}

func TestContentTypeJSON(t *testing.T) {
	h := newTestServer()
	rec := do(t, h, http.MethodGet, "/health", "")
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type = %q, want application/json", ct)
	}
}

func TestCORSHeadersAndPreflight(t *testing.T) {
	h := newTestServer()
	rec := do(t, h, http.MethodOptions, "/api/v1/search", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d, want 204", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("missing CORS allow-origin header")
	}
}

func TestAuthRegister(t *testing.T) {
	h := newTestServer()

	t.Run("happy path 201 with token bundle", func(t *testing.T) {
		rec := do(t, h, http.MethodPost, "/api/v1/auth/register", `{"email":"a@b.com","password":"secret"}`)
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201", rec.Code)
		}
		m := decode(t, rec)
		for _, k := range []string{"token", "refresh_token", "expires_at"} {
			if _, ok := m[k]; !ok {
				t.Fatalf("missing key %q in %v", k, m)
			}
		}
	})

	t.Run("missing fields 400 invalid_input", func(t *testing.T) {
		rec := do(t, h, http.MethodPost, "/api/v1/auth/register", `{"email":""}`)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
		if decode(t, rec)["error"] != "invalid_input" {
			t.Fatalf("wrong error code: %v", decode(t, rec))
		}
	})

	t.Run("duplicate email 409 email_exists", func(t *testing.T) {
		rec := do(t, h, http.MethodPost, "/api/v1/auth/register", `{"email":"taken@mergemarket.app","password":"x"}`)
		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want 409", rec.Code)
		}
		if decode(t, rec)["error"] != "email_exists" {
			t.Fatal("wrong error code")
		}
	})
}

func TestAuthLogin(t *testing.T) {
	h := newTestServer()

	t.Run("valid credentials 200", func(t *testing.T) {
		rec := do(t, h, http.MethodPost, "/api/v1/auth/login", `{"email":"a@b.com","password":"secret"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	})

	t.Run("wrong password 401 invalid_credentials", func(t *testing.T) {
		rec := do(t, h, http.MethodPost, "/api/v1/auth/login", `{"email":"a@b.com","password":"wrongpassword"}`)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
		if decode(t, rec)["error"] != "invalid_credentials" {
			t.Fatal("wrong error code")
		}
	})
}

func TestAuthRefresh(t *testing.T) {
	h := newTestServer()

	t.Run("valid refresh 200", func(t *testing.T) {
		rec := do(t, h, http.MethodPost, "/api/v1/auth/refresh", `{"refresh_token":"good"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	})

	t.Run("expired refresh 401 token_expired", func(t *testing.T) {
		rec := do(t, h, http.MethodPost, "/api/v1/auth/refresh", `{"refresh_token":"expired"}`)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
		if decode(t, rec)["error"] != "token_expired" {
			t.Fatal("wrong error code")
		}
	})
}

func TestSearch(t *testing.T) {
	h := newTestServer()

	t.Run("happy path returns multi-store results with total_cost = price + shipping", func(t *testing.T) {
		rec := do(t, h, http.MethodGet, "/api/v1/search?q=phone&location=CM", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		var resp fixtures.SearchResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.Query != "phone" {
			t.Fatalf("query echoed = %q, want phone", resp.Query)
		}
		if len(resp.Results) < 2 {
			t.Fatalf("expected multiple store results, got %d", len(resp.Results))
		}
		for _, r := range resp.Results {
			if r.TotalCost != r.Price+r.Shipping {
				t.Fatalf("total_cost invariant broken: %+v", r)
			}
			if r.DealScore < 0 || r.DealScore > 100 {
				t.Fatalf("deal_score out of range: %d", r.DealScore)
			}
		}
	})

	t.Run("missing q 400 missing_query", func(t *testing.T) {
		rec := do(t, h, http.MethodGet, "/api/v1/search?location=CM", "")
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
		if decode(t, rec)["error"] != "missing_query" {
			t.Fatal("wrong error code")
		}
	})

	t.Run("timeout sentinel 504", func(t *testing.T) {
		rec := do(t, h, http.MethodGet, "/api/v1/search?q=timeout", "")
		if rec.Code != http.StatusGatewayTimeout {
			t.Fatalf("status = %d, want 504", rec.Code)
		}
	})
}

func TestProductHistory(t *testing.T) {
	h := newTestServer()

	t.Run("happy path 200 with 6 points and aggregates", func(t *testing.T) {
		rec := do(t, h, http.MethodGet, "/api/v1/products/prod-001/history", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		var resp fixtures.HistoryResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.ProductID != "prod-001" {
			t.Fatalf("product_id = %q", resp.ProductID)
		}
		if len(resp.History) != 6 {
			t.Fatalf("history points = %d, want 6", len(resp.History))
		}
		if resp.Average6m <= 0 || resp.Lowest30d <= 0 {
			t.Fatalf("aggregates not populated: %+v", resp)
		}
	})

	t.Run("unknown product 404", func(t *testing.T) {
		rec := do(t, h, http.MethodGet, "/api/v1/products/unknown/history", "")
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", rec.Code)
		}
	})
}

func TestTruthScore(t *testing.T) {
	h := newTestServer()
	rec := do(t, h, http.MethodGet, "/api/v1/products/prod-001/truth-score", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	m := decode(t, rec)
	if m["sentiment"] == nil || m["fake_review_risk"] == nil {
		t.Fatalf("missing truth-score fields: %v", m)
	}
}

func TestWishlist(t *testing.T) {
	h := newTestServer()

	t.Run("list 200", func(t *testing.T) {
		rec := do(t, h, http.MethodGet, "/api/v1/wishlist", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	})

	t.Run("add new 201", func(t *testing.T) {
		rec := do(t, h, http.MethodPost, "/api/v1/wishlist", `{"product_id":"prod-999"}`)
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201", rec.Code)
		}
		if decode(t, rec)["wishlist_id"] == nil {
			t.Fatal("missing wishlist_id")
		}
	})

	t.Run("add duplicate 409", func(t *testing.T) {
		rec := do(t, h, http.MethodPost, "/api/v1/wishlist", `{"product_id":"prod-001"}`)
		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want 409", rec.Code)
		}
	})

	t.Run("delete existing 204 no body", func(t *testing.T) {
		rec := do(t, h, http.MethodDelete, "/api/v1/wishlist/wl-001", "")
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want 204", rec.Code)
		}
		if rec.Body.Len() != 0 {
			t.Fatalf("204 should have no body, got %q", rec.Body.String())
		}
	})

	t.Run("delete unknown 404", func(t *testing.T) {
		rec := do(t, h, http.MethodDelete, "/api/v1/wishlist/unknown", "")
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", rec.Code)
		}
	})
}

func TestAlerts(t *testing.T) {
	h := newTestServer()

	t.Run("list 200", func(t *testing.T) {
		rec := do(t, h, http.MethodGet, "/api/v1/alerts", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	})

	t.Run("create 201", func(t *testing.T) {
		rec := do(t, h, http.MethodPost, "/api/v1/alerts", `{"product_id":"prod-001","threshold_price":200000,"currency":"XAF"}`)
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201", rec.Code)
		}
	})

	t.Run("create with bad threshold 400", func(t *testing.T) {
		rec := do(t, h, http.MethodPost, "/api/v1/alerts", `{"product_id":"prod-001","threshold_price":0}`)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("delete existing 204", func(t *testing.T) {
		rec := do(t, h, http.MethodDelete, "/api/v1/alerts/al-001", "")
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want 204", rec.Code)
		}
	})

	t.Run("delete unknown 404", func(t *testing.T) {
		rec := do(t, h, http.MethodDelete, "/api/v1/alerts/unknown", "")
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", rec.Code)
		}
	})
}

func TestSavings(t *testing.T) {
	h := newTestServer()
	rec := do(t, h, http.MethodGet, "/api/v1/savings", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp fixtures.SavingsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	var sum float64
	for _, txn := range resp.Transactions {
		sum += txn.Saved
	}
	if resp.TotalSaved != sum {
		t.Fatalf("total_saved %v != sum of transactions %v", resp.TotalSaved, sum)
	}
}

func TestMalformedJSONBody(t *testing.T) {
	h := newTestServer()
	rec := do(t, h, http.MethodPost, "/api/v1/auth/login", `{not json`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 for malformed JSON", rec.Code)
	}
}
