// Package scraper executes a single store scrape from a StoreConfig. It renders
// the store's search URL, performs the request (optionally through a proxy),
// classifies blocking responses (403/429) so the circuit breaker can react, and
// extracts the canonical product fields from the JSON response. HTML (CSS
// selector) extraction is declared in the config schema but not yet implemented.
package scraper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/proxypool"
	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/queue"
	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/storeconfig"
)

// ErrBlocked indicates the store actively rejected us (HTTP 403 or 429). The
// worker feeds this into the store's circuit breaker.
var ErrBlocked = errors.New("scraper: blocked by store (403/429)")

// ErrUnsupportedMode is returned for config modes the engine cannot execute yet
// (currently HTML/CSS-selector scraping).
var ErrUnsupportedMode = errors.New("scraper: unsupported store mode")

// ClientFactory builds the *http.Client used for a scrape. proxyAddr is an
// "ip:port" string, or empty for a direct connection. Making this a field lets
// tests inject a client backed by httptest without real networking.
type ClientFactory func(proxyAddr string, timeout time.Duration) (*http.Client, error)

// Engine performs scrapes. The zero value is not usable; construct with New.
type Engine struct {
	proxies proxypool.Provider
	timeout time.Duration
	factory ClientFactory
}

// New returns an Engine. If proxies is nil, scrapes always go direct. If factory
// is nil, the default proxy-aware factory is used.
func New(proxies proxypool.Provider, timeout time.Duration, factory ClientFactory) *Engine {
	if factory == nil {
		factory = defaultClientFactory
	}
	return &Engine{proxies: proxies, timeout: timeout, factory: factory}
}

// Scrape runs job against cfg and returns the raw products found. A 403/429
// response yields ErrBlocked; any other transport or decode failure yields a
// wrapped named error.
func (e *Engine) Scrape(ctx context.Context, cfg *storeconfig.StoreConfig, job queue.Job) (queue.RawResult, error) {
	if cfg.Mode != storeconfig.ModeJSONAPI {
		return queue.RawResult{}, fmt.Errorf("%w: %q", ErrUnsupportedMode, cfg.Mode)
	}

	reqURL := renderURL(cfg.Search.URLTemplate, job.Query, job.Location)

	proxyAddr := ""
	if e.proxies != nil {
		addr, err := e.proxies.Pick(ctx)
		if err != nil {
			return queue.RawResult{}, fmt.Errorf("scraper: pick proxy: %w", err)
		}
		proxyAddr = addr
	}

	client, err := e.factory(proxyAddr, e.timeout)
	if err != nil {
		return queue.RawResult{}, fmt.Errorf("scraper: build client: %w", err)
	}

	method := cfg.Search.Method
	if method == "" {
		method = http.MethodGet
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return queue.RawResult{}, fmt.Errorf("scraper: build request: %w", err)
	}
	for k, v := range cfg.Search.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return queue.RawResult{}, fmt.Errorf("scraper: request %s: %w", cfg.StoreID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		return queue.RawResult{}, fmt.Errorf("%w: status %d from %s", ErrBlocked, resp.StatusCode, cfg.StoreID)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return queue.RawResult{}, fmt.Errorf("scraper: %s returned status %d", cfg.StoreID, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20)) // cap at 16 MiB
	if err != nil {
		return queue.RawResult{}, fmt.Errorf("scraper: read body: %w", err)
	}

	products, err := extractJSON(body, cfg)
	if err != nil {
		return queue.RawResult{}, err
	}

	return queue.RawResult{
		JobID:     job.JobID,
		StoreID:   cfg.StoreID,
		Store:     cfg.Name,
		Query:     job.Query,
		Location:  job.Location,
		Products:  products,
		ScrapedAt: time.Now().UTC(),
	}, nil
}

// extractJSON decodes the body and maps each result item to a RawProduct using
// the store's JSON field mapping. Items missing a usable title or price are
// skipped rather than failing the whole scrape (NFR-2 resilience).
func extractJSON(body []byte, cfg *storeconfig.StoreConfig) ([]queue.RawProduct, error) {
	var root any
	if err := json.Unmarshal(body, &root); err != nil {
		return nil, fmt.Errorf("scraper: %s: decode JSON: %w", cfg.StoreID, err)
	}

	m := cfg.JSON
	node, ok := lookup(root, m.ResultsPath)
	if !ok {
		return nil, fmt.Errorf("scraper: %s: results_path %q not found", cfg.StoreID, m.ResultsPath)
	}
	items, ok := node.([]any)
	if !ok {
		return nil, fmt.Errorf("scraper: %s: results_path %q is not an array", cfg.StoreID, m.ResultsPath)
	}

	products := make([]queue.RawProduct, 0, len(items))
	for _, item := range items {
		p, ok := mapProduct(item, m, cfg)
		if !ok {
			continue
		}
		products = append(products, p)
	}
	return products, nil
}

// mapProduct builds one RawProduct from a result item. It returns false when the
// item lacks the minimum viable fields (title and a parseable price).
func mapProduct(item any, m *storeconfig.JSONMapping, cfg *storeconfig.StoreConfig) (queue.RawProduct, bool) {
	get := func(path string) (any, bool) {
		if path == "" {
			return nil, false
		}
		return lookup(item, path)
	}

	titleVal, _ := get(m.Title)
	title := strings.TrimSpace(asString(titleVal))
	if title == "" {
		return queue.RawProduct{}, false
	}

	priceVal, ok := get(m.Price)
	if !ok {
		return queue.RawProduct{}, false
	}
	price, ok := asFloat(priceVal)
	if !ok {
		return queue.RawProduct{}, false
	}

	p := queue.RawProduct{Title: title, Price: price}

	if v, ok := get(m.ProductID); ok {
		p.ProductID = asString(v)
	}
	if v, ok := get(m.Shipping); ok {
		if f, ok := asFloat(v); ok {
			p.Shipping = f
		}
	}
	if v, ok := get(m.ImageURL); ok {
		p.ImageURL = asString(v)
	}
	if v, ok := get(m.URL); ok {
		p.URL = resolveURL(cfg.BaseURL, asString(v))
	}

	p.Currency = cfg.Currency
	if v, ok := get(m.Currency); ok {
		if c := strings.TrimSpace(asString(v)); c != "" {
			p.Currency = c
		}
	}

	return p, true
}

// renderURL substitutes {query} and {location} (URL-escaped) into the template.
func renderURL(tmpl, query, location string) string {
	r := strings.NewReplacer(
		"{query}", url.QueryEscape(query),
		"{location}", url.QueryEscape(location),
	)
	return r.Replace(tmpl)
}

// resolveURL turns a possibly-relative product URL into an absolute one using
// the store's base URL. On any parse failure it returns ref unchanged.
func resolveURL(base, ref string) string {
	if ref == "" || base == "" {
		return ref
	}
	b, err := url.Parse(base)
	if err != nil {
		return ref
	}
	r, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return b.ResolveReference(r).String()
}

// defaultClientFactory builds an http.Client, routing through proxyAddr when it
// is non-empty.
func defaultClientFactory(proxyAddr string, timeout time.Duration) (*http.Client, error) {
	transport := &http.Transport{}
	if proxyAddr != "" {
		proxyURL, err := url.Parse("http://" + proxyAddr)
		if err != nil {
			return nil, fmt.Errorf("scraper: parse proxy %q: %w", proxyAddr, err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	return &http.Client{Timeout: timeout, Transport: transport}, nil
}
