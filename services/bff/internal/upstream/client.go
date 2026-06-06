// Package upstream is the BFF's typed client for the handful of upstream
// endpoints it needs to build the aggregate product-detail view. It knows the
// upstream JSON shapes (project_docs/api/API_CONTRACTS.md) so the aggregator and
// server do not.
package upstream

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client calls the upstream API (mock server locally, Kong/real services in prod).
type Client struct {
	baseURL string
	http    *http.Client
}

// New returns a Client targeting baseURL with a sane request timeout.
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// PricePoint is one observation in a price-history series.
type PricePoint struct {
	Price      float64 `json:"price"`
	Currency   string  `json:"currency"`
	RecordedAt string  `json:"recorded_at"`
}

// History is the GET /products/{id}/history response.
type History struct {
	ProductID string       `json:"product_id"`
	Title     string       `json:"title"`
	History   []PricePoint `json:"history"`
	Average6m float64      `json:"average_6m"`
	Lowest30d float64      `json:"lowest_30d"`
}

// TruthScore is the GET /products/{id}/truth-score response.
type TruthScore struct {
	ProductID      string `json:"product_id"`
	Score          int    `json:"score"`
	Sentiment      string `json:"sentiment"`
	FakeReviewRisk string `json:"fake_review_risk"`
	Summary        string `json:"summary"`
}

// Offer is one store result within a search response.
type Offer struct {
	ProductID    string  `json:"product_id"`
	Title        string  `json:"title"`
	Price        float64 `json:"price"`
	Currency     string  `json:"currency"`
	Shipping     float64 `json:"shipping"`
	TotalCost    float64 `json:"total_cost"`
	ImageURL     string  `json:"image_url"`
	Store        string  `json:"store"`
	AffiliateURL string  `json:"affiliate_url"`
	DealScore    int     `json:"deal_score"`
	ScrapedAt    string  `json:"scraped_at"`
}

// SearchResponse is the GET /search response.
type SearchResponse struct {
	Query     string  `json:"query"`
	Results   []Offer `json:"results"`
	Cached    bool    `json:"cached"`
	LatencyMs int     `json:"latency_ms"`
}

// NotFoundError signals the upstream returned 404 for the requested resource.
type NotFoundError struct{ Path string }

func (e *NotFoundError) Error() string { return "upstream: not found: " + e.Path }

// History fetches a product's price history. Returns *NotFoundError on 404.
func (c *Client) History(ctx context.Context, productID string) (History, error) {
	var out History
	err := c.getJSON(ctx, "/api/v1/products/"+url.PathEscape(productID)+"/history", nil, &out)
	return out, err
}

// TruthScore fetches a product's review truth score.
func (c *Client) TruthScore(ctx context.Context, productID string) (TruthScore, error) {
	var out TruthScore
	err := c.getJSON(ctx, "/api/v1/products/"+url.PathEscape(productID)+"/truth-score", nil, &out)
	return out, err
}

// Search runs a store search for query in location.
func (c *Client) Search(ctx context.Context, query, location string) (SearchResponse, error) {
	var out SearchResponse
	q := url.Values{"q": {query}, "location": {location}}
	err := c.getJSON(ctx, "/api/v1/search", q, &out)
	return out, err
}

// getJSON performs a GET and decodes a 200 body into out. A 404 becomes a
// *NotFoundError; any other non-200 becomes a generic error.
func (c *Client) getJSON(ctx context.Context, path string, query url.Values, out any) error {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("upstream: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("upstream: %s: %w", path, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("upstream: decode %s: %w", path, err)
		}
		return nil
	case http.StatusNotFound:
		io.Copy(io.Discard, resp.Body)
		return &NotFoundError{Path: path}
	default:
		io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("upstream: %s returned %d", path, resp.StatusCode)
	}
}
