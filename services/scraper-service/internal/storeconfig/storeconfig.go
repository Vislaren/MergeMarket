// Package storeconfig defines the StoreConfig schema and a loader that reads
// per-store config files (JSON or YAML) from a directory. Each file describes
// how to scrape one store: where to send the search request and how to pull the
// canonical product fields out of the response. This is the heart of the
// "config-driven" scraper — adding a store is a config file, not code.
package storeconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Mode selects how a store's search response is parsed.
type Mode string

const (
	// ModeJSONAPI scrapes an internal/public JSON search API. This is preferred
	// per ARCHITECTURE.md §2 ("Targets internal JSON APIs where available") and
	// is the mode the scrape engine implements.
	ModeJSONAPI Mode = "json_api"
	// ModeHTML is declared for stores that only expose HTML and must be parsed
	// with CSS selectors. The selector fields below carry that configuration;
	// the engine does not yet implement HTML extraction (see scraper package).
	ModeHTML Mode = "html"
)

// Search describes how to issue the search request for a store.
type Search struct {
	// URLTemplate is the request URL with {query} and {location} placeholders,
	// each URL-escaped at render time. Example:
	//   https://api.shop.example/search?q={query}&country={location}
	URLTemplate string `json:"url_template" yaml:"url_template"`
	// Method defaults to GET when empty.
	Method string `json:"method,omitempty" yaml:"method,omitempty"`
	// Headers are sent with every request (e.g. a User-Agent or Accept header).
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
}

// JSONMapping maps canonical product fields to dotted paths inside a single
// result item of the store's JSON response (see fieldpath in the scraper pkg).
type JSONMapping struct {
	// ResultsPath is the dotted path to the array of result items in the
	// response body, e.g. "data.products". Empty means the body is itself the
	// array.
	ResultsPath string `json:"results_path" yaml:"results_path"`

	ProductID string `json:"product_id" yaml:"product_id"`
	Title     string `json:"title" yaml:"title"`
	Price     string `json:"price" yaml:"price"`
	Currency  string `json:"currency,omitempty" yaml:"currency,omitempty"`
	Shipping  string `json:"shipping,omitempty" yaml:"shipping,omitempty"`
	ImageURL  string `json:"image_url,omitempty" yaml:"image_url,omitempty"`
	URL       string `json:"url,omitempty" yaml:"url,omitempty"`
}

// HTMLMapping carries CSS selectors for HTML stores (ModeHTML). It is parsed and
// validated for completeness but extraction is not yet implemented.
type HTMLMapping struct {
	ItemSelector string `json:"item_selector" yaml:"item_selector"`
	Title        string `json:"title" yaml:"title"`
	Price        string `json:"price" yaml:"price"`
	ImageURL     string `json:"image_url,omitempty" yaml:"image_url,omitempty"`
	URL          string `json:"url,omitempty" yaml:"url,omitempty"`
}

// StoreConfig is the complete scrape definition for one store.
type StoreConfig struct {
	// StoreID is the stable identifier used to route jobs to this config and to
	// key the per-store circuit breaker. Must be unique across the config dir.
	StoreID string `json:"store_id" yaml:"store_id"`
	// Name is a human-readable store name.
	Name string `json:"name" yaml:"name"`
	// BaseURL is the store's site root; used to resolve relative product URLs.
	BaseURL string `json:"base_url" yaml:"base_url"`
	// Currency is the default ISO 4217 code used when an item carries no
	// per-result currency field.
	Currency string `json:"currency,omitempty" yaml:"currency,omitempty"`
	// AffiliateTemplate, when set, is how the normalization service (A-06) will
	// wrap outbound URLs. Carried here so it lives with the store definition;
	// the scraper itself does not inject affiliate parameters.
	AffiliateTemplate string `json:"affiliate_template,omitempty" yaml:"affiliate_template,omitempty"`

	Mode   Mode   `json:"mode" yaml:"mode"`
	Search Search `json:"search" yaml:"search"`

	JSON *JSONMapping `json:"json,omitempty" yaml:"json,omitempty"`
	HTML *HTMLMapping `json:"html,omitempty" yaml:"html,omitempty"`
}

// Validate checks a single config for the invariants the engine relies on.
func (c *StoreConfig) Validate() error {
	if strings.TrimSpace(c.StoreID) == "" {
		return fmt.Errorf("storeconfig: store_id must not be empty")
	}
	if strings.TrimSpace(c.Search.URLTemplate) == "" {
		return fmt.Errorf("storeconfig: %s: search.url_template must not be empty", c.StoreID)
	}
	switch c.Mode {
	case ModeJSONAPI:
		if c.JSON == nil {
			return fmt.Errorf("storeconfig: %s: mode json_api requires a json mapping", c.StoreID)
		}
		if strings.TrimSpace(c.JSON.Title) == "" || strings.TrimSpace(c.JSON.Price) == "" {
			return fmt.Errorf("storeconfig: %s: json mapping requires at least title and price", c.StoreID)
		}
	case ModeHTML:
		if c.HTML == nil || strings.TrimSpace(c.HTML.ItemSelector) == "" {
			return fmt.Errorf("storeconfig: %s: mode html requires an html mapping with item_selector", c.StoreID)
		}
	default:
		return fmt.Errorf("storeconfig: %s: unknown mode %q (want %q or %q)", c.StoreID, c.Mode, ModeJSONAPI, ModeHTML)
	}
	return nil
}

// Parse decodes a single config from bytes, choosing the decoder by extension
// (".json" -> JSON, ".yaml"/".yml" -> YAML). The config is validated before it
// is returned.
func Parse(name string, data []byte) (*StoreConfig, error) {
	var cfg StoreConfig
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("storeconfig: parse JSON %s: %w", name, err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("storeconfig: parse YAML %s: %w", name, err)
		}
	default:
		return nil, fmt.Errorf("storeconfig: %s: unsupported extension %q (want .json, .yaml, .yml)", name, ext)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Registry is an immutable lookup of StoreConfig keyed by store id.
type Registry struct {
	byID map[string]*StoreConfig
}

// Get returns the config for a store id and whether it was found.
func (r *Registry) Get(storeID string) (*StoreConfig, bool) {
	c, ok := r.byID[storeID]
	return c, ok
}

// Len reports how many store configs are loaded.
func (r *Registry) Len() int { return len(r.byID) }

// IDs returns the loaded store ids (order is unspecified).
func (r *Registry) IDs() []string {
	ids := make([]string, 0, len(r.byID))
	for id := range r.byID {
		ids = append(ids, id)
	}
	return ids
}

// LoadDir reads every *.json/*.yaml/*.yml file in dir, parses and validates each
// into a Registry. It fails if the directory cannot be read, if any file is
// invalid, or if two configs declare the same store_id. A directory with no
// config files is an error — the service would have nothing to scrape.
func LoadDir(dir string) (*Registry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("storeconfig: read dir %s: %w", dir, err)
	}

	byID := make(map[string]*StoreConfig)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext != ".json" && ext != ".yaml" && ext != ".yml" {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("storeconfig: read %s: %w", path, err)
		}
		cfg, err := Parse(e.Name(), data)
		if err != nil {
			return nil, err
		}
		if _, dup := byID[cfg.StoreID]; dup {
			return nil, fmt.Errorf("storeconfig: duplicate store_id %q (in %s)", cfg.StoreID, e.Name())
		}
		byID[cfg.StoreID] = cfg
	}

	if len(byID) == 0 {
		return nil, fmt.Errorf("storeconfig: no store config files found in %s", dir)
	}
	return &Registry{byID: byID}, nil
}
