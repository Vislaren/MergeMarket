// Package unit_test holds Session 14's QE unit suite: independent, executable
// verification of Agent A's search (A-14) and userdata (A-16..A-18) services via
// the Go toolchain (subprocess build/test/vet/gofmt) plus structural source
// assertions of the API_CONTRACTS.md / DATABASE_SCHEMA.md invariants. It does not
// import the services' internal/ packages (Go forbids that across modules); the
// services ship their own in-package unit tests, which TC-14-U-002/-015 run.
package unit_test

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// repoRoot walks up from the test's working directory until it finds the dir that
// contains a "services" directory — the MergeMarket repo root.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if fi, err := os.Stat(filepath.Join(dir, "services")); err == nil && fi.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		require.NotEqual(t, parent, dir, "repo root (a dir containing services/) not found")
		dir = parent
	}
}

// serviceDir returns services/<name>, skipping the test if the service is absent
// (e.g. when run from a branch without Agent A's code).
func serviceDir(t *testing.T, name string) string {
	t.Helper()
	d := filepath.Join(repoRoot(t), "services", name)
	if _, err := os.Stat(d); err != nil {
		t.Skipf("service %q not present on this checkout: %v", name, err)
	}
	return d
}

// runTool runs an external tool in dir and returns combined output + error.
func runTool(t *testing.T, dir, name string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// sourceOf concatenates every non-test .go file under dir (recursively).
func sourceOf(t *testing.T, dir string) string {
	t.Helper()
	var sb strings.Builder
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return nil
		}
		b, rerr := os.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		sb.Write(b)
		sb.WriteString("\n")
		return nil
	})
	require.NoError(t, err)
	return sb.String()
}

// readFile returns a repo-relative file's contents.
func readFile(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	require.NoError(t, err)
	return string(b)
}

// eachInternalPackageHasGoDoc asserts every internal/<pkg> dir has at least one
// file beginning a "// Package <pkg>" doc comment (MergeMarket Go standard).
func eachInternalPackageHasGoDoc(t *testing.T, svcDir string) {
	t.Helper()
	internal := filepath.Join(svcDir, "internal")
	entries, err := os.ReadDir(internal)
	require.NoError(t, err)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pkgDir := filepath.Join(internal, e.Name())
		require.Contains(t, sourceOf(t, pkgDir), "// Package ",
			"internal/%s is missing a package GoDoc comment", e.Name())
	}
}
