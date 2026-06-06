// Package normalize converts a raw scraped product into the canonical
// MergeMarket product schema (API_CONTRACTS.md "Search" result item). It trims
// and sanitises text, enforces a sensible currency, drops items that cannot form
// a valid offer (no title or non-positive price), and computes total_cost. It
// does not perform I/O — affiliate injection and persistence live in their own
// packages so this stays a pure, table-testable transform.
package normalize

import (
	"strings"
	"time"

	"github.com/Vislaren/MergeMarket/services/normalization/internal/queue"
)

// defaultCurrency is used when neither the product nor the store supplied one.
const defaultCurrency = "USD"

// Product is the canonical, normalized representation of one offer. Field names
// and JSON tags match the Search API result item (minus deal_score, which is
// assigned downstream by the Deal Meter).
type Product struct {
	ProductID string    `json:"product_id"`
	Title     string    `json:"title"`
	Price     float64   `json:"price"`
	Currency  string    `json:"currency"`
	Shipping  float64   `json:"shipping"`
	TotalCost float64   `json:"total_cost"`
	ImageURL  string    `json:"image_url"`
	URL       string    `json:"url"`
	Store     string    `json:"store"`
	StoreID   string    `json:"store_id"`
	ScrapedAt time.Time `json:"scraped_at"`
}

// Result groups the normalized products from one RawResult.
type Result struct {
	StoreID   string
	Store     string
	Query     string
	Location  string
	ScrapedAt time.Time
	Products  []Product
}

// FromRaw normalizes every product in a RawResult, skipping any that fail
// validation. The returned Result preserves the store/query context.
func FromRaw(raw queue.RawResult) Result {
	scrapedAt := raw.ScrapedAt
	if scrapedAt.IsZero() {
		scrapedAt = time.Now().UTC()
	}

	out := Result{
		StoreID:   raw.StoreID,
		Store:     strings.TrimSpace(raw.Store),
		Query:     raw.Query,
		Location:  raw.Location,
		ScrapedAt: scrapedAt,
		Products:  make([]Product, 0, len(raw.Products)),
	}

	for _, rp := range raw.Products {
		p, ok := normalizeProduct(rp, raw, scrapedAt)
		if !ok {
			continue
		}
		out.Products = append(out.Products, p)
	}
	return out
}

// normalizeProduct sanitises a single raw product. It returns false when the
// item lacks a title or a positive price, which the caller skips (NFR-2: one bad
// row never fails the batch).
func normalizeProduct(rp queue.RawProduct, raw queue.RawResult, scrapedAt time.Time) (Product, bool) {
	title := collapseSpaces(rp.Title)
	if title == "" {
		return Product{}, false
	}
	if rp.Price <= 0 {
		return Product{}, false
	}

	shipping := rp.Shipping
	if shipping < 0 {
		shipping = 0
	}

	currency := normalizeCurrency(rp.Currency)

	return Product{
		ProductID: strings.TrimSpace(rp.ProductID),
		Title:     title,
		Price:     round2(rp.Price),
		Currency:  currency,
		Shipping:  round2(shipping),
		TotalCost: round2(rp.Price + shipping),
		ImageURL:  strings.TrimSpace(rp.ImageURL),
		URL:       strings.TrimSpace(rp.URL),
		Store:     strings.TrimSpace(raw.Store),
		StoreID:   raw.StoreID,
		ScrapedAt: scrapedAt,
	}, true
}

// normalizeCurrency upper-cases a 3-letter ISO 4217 code, falling back to USD
// when the value is absent or not a 3-letter code.
func normalizeCurrency(c string) string {
	c = strings.ToUpper(strings.TrimSpace(c))
	if len(c) != 3 {
		return defaultCurrency
	}
	for _, r := range c {
		if r < 'A' || r > 'Z' {
			return defaultCurrency
		}
	}
	return c
}

// collapseSpaces trims and collapses internal runs of whitespace to a single
// space so titles persist cleanly.
func collapseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// round2 rounds to 2 decimal places (currency precision), avoiding float noise
// from upstream JSON like 19.989999999.
func round2(f float64) float64 {
	return float64(int64(f*100+sign(f)*0.5)) / 100
}

func sign(f float64) float64 {
	if f < 0 {
		return -1
	}
	return 1
}
