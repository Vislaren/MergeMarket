package storeconfig

import (
	"os"
	"path/filepath"
	"testing"
)

const validJSON = `{
  "store_id": "shop1",
  "name": "Shop One",
  "base_url": "https://shop1.example",
  "mode": "json_api",
  "search": { "url_template": "https://shop1.example/s?q={query}" },
  "json": { "results_path": "data.products", "title": "name", "price": "price" }
}`

const validYAML = `
store_id: shop2
name: Shop Two
base_url: https://shop2.example
mode: json_api
search:
  url_template: https://shop2.example/s?q={query}&c={location}
json:
  results_path: products
  title: title
  price: price
`

func TestParseJSON(t *testing.T) {
	cfg, err := Parse("shop1.json", []byte(validJSON))
	if err != nil {
		t.Fatalf("Parse JSON error: %v", err)
	}
	if cfg.StoreID != "shop1" || cfg.Mode != ModeJSONAPI {
		t.Errorf("got %+v", cfg)
	}
	if cfg.JSON == nil || cfg.JSON.ResultsPath != "data.products" {
		t.Errorf("json mapping not parsed: %+v", cfg.JSON)
	}
}

func TestParseYAML(t *testing.T) {
	cfg, err := Parse("shop2.yaml", []byte(validYAML))
	if err != nil {
		t.Fatalf("Parse YAML error: %v", err)
	}
	if cfg.StoreID != "shop2" || cfg.JSON.Title != "title" {
		t.Errorf("got %+v / %+v", cfg, cfg.JSON)
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name, file, body string
	}{
		{"unsupported ext", "shop.txt", validJSON},
		{"malformed json", "shop.json", "{not json"},
		{"missing store_id", "x.json", `{"mode":"json_api","search":{"url_template":"u"},"json":{"title":"t","price":"p"}}`},
		{"missing url template", "x.json", `{"store_id":"s","mode":"json_api","json":{"title":"t","price":"p"}}`},
		{"json mode without mapping", "x.json", `{"store_id":"s","mode":"json_api","search":{"url_template":"u"}}`},
		{"json mode missing price", "x.json", `{"store_id":"s","mode":"json_api","search":{"url_template":"u"},"json":{"title":"t"}}`},
		{"unknown mode", "x.json", `{"store_id":"s","mode":"carrier-pigeon","search":{"url_template":"u"}}`},
		{"html mode without selector", "x.json", `{"store_id":"s","mode":"html","search":{"url_template":"u"}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Parse(tt.file, []byte(tt.body)); err == nil {
				t.Fatalf("expected error for %s", tt.name)
			}
		})
	}
}

func TestParseHTMLModeValid(t *testing.T) {
	body := `{"store_id":"h","mode":"html","search":{"url_template":"u"},"html":{"item_selector":".card","title":".t","price":".p"}}`
	cfg, err := Parse("h.json", []byte(body))
	if err != nil {
		t.Fatalf("html parse error: %v", err)
	}
	if cfg.HTML == nil || cfg.HTML.ItemSelector != ".card" {
		t.Errorf("html mapping not parsed: %+v", cfg.HTML)
	}
}

func TestLoadDir(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, dir, "shop1.json", validJSON)
	mustWrite(t, dir, "shop2.yaml", validYAML)
	mustWrite(t, dir, "notes.txt", "ignored, not a config")

	reg, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir error: %v", err)
	}
	if reg.Len() != 2 {
		t.Fatalf("Len = %d, want 2 (ids=%v)", reg.Len(), reg.IDs())
	}
	if _, ok := reg.Get("shop1"); !ok {
		t.Errorf("shop1 not found")
	}
	if _, ok := reg.Get("missing"); ok {
		t.Errorf("missing store unexpectedly found")
	}
}

func TestLoadDirDuplicate(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, dir, "a.json", validJSON)
	mustWrite(t, dir, "b.json", validJSON) // same store_id "shop1"
	if _, err := LoadDir(dir); err == nil {
		t.Fatal("expected duplicate store_id error")
	}
}

func TestLoadDirEmpty(t *testing.T) {
	if _, err := LoadDir(t.TempDir()); err == nil {
		t.Fatal("expected error for empty config dir")
	}
	if _, err := LoadDir(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("expected error for missing config dir")
	}
}

func mustWrite(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}
