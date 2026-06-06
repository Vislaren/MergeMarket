//go:build integration

// Package integration_test holds Session 14's QE integration suite. It builds and
// runs the REAL search (A-14) and userdata (A-16..A-18) binaries.
//
// Two classes of case:
//   - Executable here: fail-fast / config-guard behaviour that needs no database
//     (a binary must refuse to run or exit non-zero when its hard dependencies or
//     required secrets are absent).
//   - Gated (PENDING locally): cases needing a live Postgres/Redis are skipped
//     unless DB_TEST_DSN (and, for search, REDIS_TEST_ADDR) are set. Run with:
//       DB_TEST_DSN=postgres://user:pass@localhost:5432/mergemarket?sslmode=disable \
//       REDIS_TEST_ADDR=localhost:6379 \
//       go test -tags=integration ./docs/testing/session-14/integration/test_suite/...
package integration_test

import (
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if fi, err := os.Stat(filepath.Join(dir, "services")); err == nil && fi.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		require.NotEqual(t, parent, dir, "repo root not found")
		dir = parent
	}
}

// buildBinary compiles services/<svc>/cmd/<cmd> to a temp file and returns its
// path. The service is skipped if absent on this checkout.
func buildBinary(t *testing.T, svc, cmd string) string {
	t.Helper()
	svcDir := filepath.Join(repoRoot(t), "services", svc)
	if _, err := os.Stat(svcDir); err != nil {
		t.Skipf("service %q not present on this checkout", svc)
	}
	out := filepath.Join(t.TempDir(), svc)
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	build := exec.Command("go", "build", "-o", out, "./cmd/"+cmd)
	build.Dir = svcDir
	if b, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building %s failed: %v\n%s", svc, err, b)
	}
	return out
}

// runExpectNonZero runs bin with the given env (PATH is added), and asserts the
// process exited with a non-zero status (i.e. it did not start and stay up).
// Returns the combined output for message assertions.
func runExpectNonZero(t *testing.T, bin string, env []string) string {
	t.Helper()
	cmd := exec.Command(bin)
	cmd.Env = append(env, "PATH="+os.Getenv("PATH"))
	out, err := cmd.CombinedOutput()
	require.Error(t, err, "binary was expected to exit non-zero but exited 0:\n%s", out)
	if _, ok := err.(*exec.ExitError); !ok {
		t.Fatalf("binary did not exit cleanly non-zero: %v\n%s", err, out)
	}
	return string(out)
}

// requireTestDB skips the calling test unless DB_TEST_DSN is set.
func requireTestDB(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("DB_TEST_DSN")
	if dsn == "" {
		t.Skip("PENDING: set DB_TEST_DSN to run this live-backend case (see package doc)")
	}
	return dsn
}

// freePort returns a currently-free localhost TCP port as a string.
func freePort(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()
	_, port, err := net.SplitHostPort(ln.Addr().String())
	require.NoError(t, err)
	return port
}

// waitHealthy polls http://localhost:port/health until it returns 200 or the
// deadline passes.
func waitHealthy(port string, timeout time.Duration) bool {
	client := &http.Client{Timeout: time.Second}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := client.Get("http://localhost:" + port + "/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return true
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}
