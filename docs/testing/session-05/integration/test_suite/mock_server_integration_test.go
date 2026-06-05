//go:build integration

// Package mockserver_integration_test drives the REAL mock-server binary over
// HTTP. It builds the service, starts it on a free port, waits for /health, then
// exercises every API contract — asserting status codes, key fields, and the
// total_cost = price + shipping invariant. No mocks, no Redis: the mock server
// is self-contained, so this whole suite runs offline.
//
// Run: go test -tags=integration ./docs/testing/session-05/integration/test_suite/...
package mockserver_integration_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// servicePath returns the absolute path to services/mock-server from this test
// file's location (…/docs/testing/session-05/integration/test_suite).
func servicePath(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve caller path")
	}
	// up 5: test_suite → integration → session-05 → testing → docs → repo root
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "..")
	svc := filepath.Join(root, "services", "mock-server")
	if _, err := os.Stat(filepath.Join(svc, "go.mod")); err != nil {
		t.Fatalf("mock-server not found at %s: %v", svc, err)
	}
	return svc
}

// freePort asks the OS for an unused TCP port and returns it as a string.
func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("cannot find free port: %v", err)
	}
	defer l.Close()
	return fmt.Sprintf("%d", l.Addr().(*net.TCPAddr).Port)
}

// startServer builds and launches the mock-server binary on a free port and
// returns its base URL plus a cleanup function.
func startServer(t *testing.T) (string, func()) {
	t.Helper()
	svc := servicePath(t)

	bin := filepath.Join(t.TempDir(), "mock-server")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	build := exec.Command("go", "build", "-o", bin, "./cmd/mock-server")
	build.Dir = svc
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}

	port := freePort(t)
	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(), "MOCK_SERVER_PORT="+port)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start binary: %v", err)
	}

	base := "http://127.0.0.1:" + port
	cleanup := func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	}

	// Wait for readiness.
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(base + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return base, cleanup
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	cleanup()
	t.Fatal("server did not become ready within 10s")
	return "", nil
}

// getJSON performs a GET and decodes the JSON body into a map.
func getJSON(t *testing.T, url string) (int, map[string]any) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var m map[string]any
	if len(body) > 0 {
		_ = json.Unmarshal(body, &m)
	}
	return resp.StatusCode, m
}

// status performs an arbitrary request and returns just the status code.
func status(t *testing.T, method, url, body string) int {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = io.NopCloser(stringReader(body))
	}
	req, err := http.NewRequest(method, url, r)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return resp.StatusCode
}

type stringReader string

func (s stringReader) Read(p []byte) (int, error) {
	n := copy(p, s)
	if n == len(s) {
		return n, io.EOF
	}
	return n, nil
}

func TestMockServerIntegration(t *testing.T) {
	base, cleanup := startServer(t)
	defer cleanup()

	t.Run("TC-05-I-001: GET /health returns the contract shape", func(t *testing.T) {
		code, m := getJSON(t, base+"/health")
		if code != http.StatusOK {
			t.Fatalf("status = %d, want 200", code)
		}
		if m["status"] != "ok" || m["service"] != "mock-server" || m["version"] == nil {
			t.Fatalf("unexpected health body: %v", m)
		}
	})

	t.Run("TC-05-I-002: GET /search returns multi-store results, total_cost = price + shipping", func(t *testing.T) {
		resp, err := http.Get(base + "/api/v1/search?q=phone&location=CM")
		if err != nil {
			t.Fatalf("GET search: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		var sr struct {
			Query   string `json:"query"`
			Results []struct {
				Price     float64 `json:"price"`
				Shipping  float64 `json:"shipping"`
				TotalCost float64 `json:"total_cost"`
				DealScore int     `json:"deal_score"`
			} `json:"results"`
			Cached bool `json:"cached"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if sr.Query != "phone" {
			t.Fatalf("query echoed = %q", sr.Query)
		}
		if len(sr.Results) < 2 {
			t.Fatalf("want multiple stores, got %d", len(sr.Results))
		}
		for _, r := range sr.Results {
			if r.TotalCost != r.Price+r.Shipping {
				t.Fatalf("total_cost invariant broken: %+v", r)
			}
			if r.DealScore < 0 || r.DealScore > 100 {
				t.Fatalf("deal_score out of range: %d", r.DealScore)
			}
		}
	})

	t.Run("TC-05-I-003: search error paths (400 missing_query, 504 timeout)", func(t *testing.T) {
		if c := status(t, http.MethodGet, base+"/api/v1/search", ""); c != http.StatusBadRequest {
			t.Fatalf("missing q: status = %d, want 400", c)
		}
		if c := status(t, http.MethodGet, base+"/api/v1/search?q=timeout", ""); c != http.StatusGatewayTimeout {
			t.Fatalf("timeout sentinel: status = %d, want 504", c)
		}
	})

	t.Run("TC-05-I-004: product history returns 6 points + aggregates; unknown is 404", func(t *testing.T) {
		resp, err := http.Get(base + "/api/v1/products/prod-001/history")
		if err != nil {
			t.Fatalf("GET history: %v", err)
		}
		defer resp.Body.Close()
		var hr struct {
			History   []any   `json:"history"`
			Average6m float64 `json:"average_6m"`
			Lowest30d float64 `json:"lowest_30d"`
		}
		json.NewDecoder(resp.Body).Decode(&hr)
		if len(hr.History) != 6 || hr.Average6m <= 0 || hr.Lowest30d <= 0 {
			t.Fatalf("unexpected history payload: %+v", hr)
		}
		if c := status(t, http.MethodGet, base+"/api/v1/products/unknown/history", ""); c != http.StatusNotFound {
			t.Fatalf("unknown product: status = %d, want 404", c)
		}
	})

	t.Run("TC-05-I-005: auth flows (201 register, 401 login, 401 refresh)", func(t *testing.T) {
		if c := status(t, http.MethodPost, base+"/api/v1/auth/register", `{"email":"new@x.com","password":"p"}`); c != http.StatusCreated {
			t.Fatalf("register: status = %d, want 201", c)
		}
		if c := status(t, http.MethodPost, base+"/api/v1/auth/login", `{"email":"a@b.com","password":"wrongpassword"}`); c != http.StatusUnauthorized {
			t.Fatalf("bad login: status = %d, want 401", c)
		}
		if c := status(t, http.MethodPost, base+"/api/v1/auth/refresh", `{"refresh_token":"expired"}`); c != http.StatusUnauthorized {
			t.Fatalf("expired refresh: status = %d, want 401", c)
		}
	})

	t.Run("TC-05-I-006: wishlist + alerts CRUD status codes", func(t *testing.T) {
		if c := status(t, http.MethodGet, base+"/api/v1/wishlist", ""); c != http.StatusOK {
			t.Fatalf("wishlist list: %d", c)
		}
		if c := status(t, http.MethodPost, base+"/api/v1/wishlist", `{"product_id":"prod-001"}`); c != http.StatusConflict {
			t.Fatalf("duplicate wishlist: status = %d, want 409", c)
		}
		if c := status(t, http.MethodDelete, base+"/api/v1/wishlist/wl-001", ""); c != http.StatusNoContent {
			t.Fatalf("wishlist delete: status = %d, want 204", c)
		}
		if c := status(t, http.MethodPost, base+"/api/v1/alerts", `{"product_id":"prod-001","threshold_price":200000,"currency":"XAF"}`); c != http.StatusCreated {
			t.Fatalf("alert create: status = %d, want 201", c)
		}
		if c := status(t, http.MethodDelete, base+"/api/v1/alerts/unknown", ""); c != http.StatusNotFound {
			t.Fatalf("alert delete unknown: status = %d, want 404", c)
		}
	})

	t.Run("TC-05-I-007: savings total equals sum of transactions", func(t *testing.T) {
		resp, err := http.Get(base + "/api/v1/savings")
		if err != nil {
			t.Fatalf("GET savings: %v", err)
		}
		defer resp.Body.Close()
		var s struct {
			TotalSaved   float64 `json:"total_saved"`
			Transactions []struct {
				Saved float64 `json:"saved"`
			} `json:"transactions"`
		}
		json.NewDecoder(resp.Body).Decode(&s)
		var sum float64
		for _, tx := range s.Transactions {
			sum += tx.Saved
		}
		if s.TotalSaved != sum {
			t.Fatalf("total_saved %v != sum %v", s.TotalSaved, sum)
		}
	})

	t.Run("TC-05-I-008: CORS preflight returns 204 with allow-origin", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodOptions, base+"/api/v1/search", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("OPTIONS: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("preflight status = %d, want 204", resp.StatusCode)
		}
		if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
			t.Fatal("missing CORS allow-origin")
		}
	})
}
