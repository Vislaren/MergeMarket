// Package mockserver_test verifies Agent B's B-02 mock-server as an independent
// QE check. Because the service's logic lives in `internal/` packages that a
// separate module cannot import, this suite uses the same approach as Session-02
// (A-02) and Session-04 (A-04): a subprocess toolchain gate (build/test/vet/
// gofmt) plus structural source assertions of the contract-level invariants
// drawn from project_docs/api/API_CONTRACTS.md and the B-02 task spec.
//
// Run: go test ./docs/testing/session-05/unit/test_suite/...
package mockserver_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// servicePath resolves services/mock-server relative to this test file.
func servicePath(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve caller path")
	}
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "..")
	svc := filepath.Join(root, "services", "mock-server")
	if _, err := os.Stat(filepath.Join(svc, "go.mod")); err != nil {
		t.Fatalf("mock-server not found at %s: %v", svc, err)
	}
	return svc
}

// runGo runs a go subcommand inside the service directory.
func runGo(t *testing.T, svc string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = svc
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// readFile reads a source file under the service directory.
func readFile(t *testing.T, svc, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(svc, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(b)
}

func TestMockServerUnit(t *testing.T) {
	svc := servicePath(t)

	t.Run("TC-05-U-001: service compiles (go build)", func(t *testing.T) {
		if out, err := runGo(t, svc, "build", "./..."); err != nil {
			t.Fatalf("go build failed: %v\n%s", err, out)
		}
	})

	t.Run("TC-05-U-002: developer in-package tests pass (go test)", func(t *testing.T) {
		if out, err := runGo(t, svc, "test", "./..."); err != nil {
			t.Fatalf("go test failed: %v\n%s", err, out)
		}
	})

	t.Run("TC-05-U-003: go vet clean", func(t *testing.T) {
		if out, err := runGo(t, svc, "vet", "./..."); err != nil {
			t.Fatalf("go vet failed: %v\n%s", err, out)
		}
	})

	t.Run("TC-05-U-004: gofmt clean", func(t *testing.T) {
		out, err := runGo(t, svc, "fmt", "-n", "./...") // dry-run lists files needing fmt
		if err != nil {
			t.Fatalf("go fmt check failed: %v\n%s", err, out)
		}
		// Cross-check with gofmt -l for certainty.
		cmd := exec.Command("gofmt", "-l", ".")
		cmd.Dir = svc
		lout, _ := cmd.CombinedOutput()
		if strings.TrimSpace(string(lout)) != "" {
			t.Fatalf("gofmt reports unformatted files:\n%s", lout)
		}
	})

	t.Run("TC-05-U-005: dependency-free (go.mod has no require block)", func(t *testing.T) {
		mod := readFile(t, svc, "go.mod")
		if strings.Contains(mod, "require") {
			t.Fatalf("mock server should be stdlib-only; go.mod has a require block:\n%s", mod)
		}
	})

	t.Run("TC-05-U-006: config is env-driven with default port 8080", func(t *testing.T) {
		cfg := readFile(t, svc, "internal/config/config.go")
		if !strings.Contains(cfg, "MOCK_SERVER_PORT") || !strings.Contains(cfg, "os.Getenv") {
			t.Fatal("config does not read MOCK_SERVER_PORT from the environment")
		}
		if !strings.Contains(cfg, "defaultPort = 8080") {
			t.Fatal("default port is not 8080 (B-02 task spec)")
		}
	})

	t.Run("TC-05-U-007: /health returns {status, service, version}", func(t *testing.T) {
		srv := readFile(t, svc, "internal/server/server.go")
		if !strings.Contains(srv, `"GET /health"`) {
			t.Fatal("no GET /health route registered")
		}
		fx := readFile(t, svc, "internal/fixtures/fixtures.go")
		for _, field := range []string{`json:"status"`, `json:"service"`, `json:"version"`} {
			if !strings.Contains(fx, field) {
				t.Fatalf("HealthResponse missing field tag %s", field)
			}
		}
	})

	t.Run("TC-05-U-008: every API_CONTRACTS endpoint is routed", func(t *testing.T) {
		srv := readFile(t, svc, "internal/server/server.go")
		routes := []string{
			"POST /api/v1/auth/register",
			"POST /api/v1/auth/login",
			"POST /api/v1/auth/refresh",
			"GET /api/v1/search",
			"GET /api/v1/products/{product_id}/history",
			"GET /api/v1/products/{product_id}/truth-score",
			"GET /api/v1/wishlist",
			"POST /api/v1/wishlist",
			"DELETE /api/v1/wishlist/{wishlist_id}",
			"GET /api/v1/alerts",
			"POST /api/v1/alerts",
			"DELETE /api/v1/alerts/{alert_id}",
			"GET /api/v1/savings",
		}
		for _, r := range routes {
			if !strings.Contains(srv, `"`+r+`"`) {
				t.Errorf("missing route registration: %s", r)
			}
		}
	})

	t.Run("TC-05-U-009: search result keeps total_cost = price + shipping", func(t *testing.T) {
		fx := readFile(t, svc, "internal/fixtures/fixtures.go")
		if !strings.Contains(fx, "TotalCost:    price + shipping") {
			t.Fatal("search offer does not compute total_cost as price + shipping")
		}
	})

	t.Run("TC-05-U-010: error shape is {error, message}", func(t *testing.T) {
		fx := readFile(t, svc, "internal/fixtures/fixtures.go")
		if !strings.Contains(fx, `json:"error"`) || !strings.Contains(fx, `json:"message"`) {
			t.Fatal("ErrorResponse does not match the canonical {error, message} shape")
		}
	})

	t.Run("TC-05-U-011: structured logging via log/slog", func(t *testing.T) {
		main := readFile(t, svc, "cmd/mock-server/main.go")
		if !strings.Contains(main, "log/slog") {
			t.Fatal("service does not use log/slog")
		}
	})

	t.Run("TC-05-U-012: every package has a GoDoc package comment", func(t *testing.T) {
		for _, f := range []string{
			"cmd/mock-server/main.go",
			"internal/config/config.go",
			"internal/fixtures/fixtures.go",
			"internal/server/server.go",
		} {
			src := readFile(t, svc, f)
			if !strings.Contains(src, "// Package ") && !strings.Contains(src, "// Command ") {
				t.Errorf("%s missing GoDoc package/command comment", f)
			}
		}
	})
}
