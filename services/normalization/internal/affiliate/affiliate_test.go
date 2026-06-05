package affiliate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInject_NoConfigIsNoOp(t *testing.T) {
	in := New()
	if got := in.Inject("any", "https://shop/p/1"); got != "https://shop/p/1" {
		t.Errorf("no-op injector changed url: %q", got)
	}
	if got := in.Inject("any", ""); got != "" {
		t.Errorf("empty url should stay empty, got %q", got)
	}
}

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "affiliates.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadAndInject(t *testing.T) {
	path := writeConfig(t, `{
		"default_params": { "utm_source": "mergemarket" },
		"stores": {
			"jumia-cm": { "params": { "aff": "mm-21" } },
			"deep":     { "template": "https://go.partner.com/?u={url}&id=9" }
		}
	}`)

	in, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	// Store-specific params override defaults.
	got := in.Inject("jumia-cm", "https://jumia.cm/p/1")
	if got != "https://jumia.cm/p/1?aff=mm-21" {
		t.Errorf("store params: got %q", got)
	}

	// Unknown store falls back to default params.
	got = in.Inject("unknown", "https://other.com/x")
	if got != "https://other.com/x?utm_source=mergemarket" {
		t.Errorf("default params: got %q", got)
	}

	// Template wraps and URL-encodes the product URL.
	got = in.Inject("deep", "https://shop.com/a?b=c")
	want := "https://go.partner.com/?u=https%3A%2F%2Fshop.com%2Fa%3Fb%3Dc&id=9"
	if got != want {
		t.Errorf("template: got %q want %q", got, want)
	}
}

func TestLoad_EmptyPath(t *testing.T) {
	in, err := Load("")
	if err != nil {
		t.Fatalf("empty path should not error: %v", err)
	}
	if got := in.Inject("s", "https://u"); got != "https://u" {
		t.Errorf("expected no-op, got %q", got)
	}
}

func TestLoad_MissingFileErrors(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "nope.json")); err == nil {
		t.Errorf("expected error for missing file")
	}
}
