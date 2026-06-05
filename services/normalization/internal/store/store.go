// Package store persists normalized products into PostgreSQL. Each scraped store
// is resolved (upserted) to a stores row by name, and each product is upserted
// into the products table keyed by (store_id, url) so re-scrapes update the
// existing row and refresh last_price/scraped_at rather than duplicating it.
//
// Persistence is expressed behind the Repository interface so the worker pool
// can be unit-tested with a fake; the pgx-backed implementation lives in
// postgres.go.
package store

import (
	"context"
	"time"
)

// Product is the row written to the products table. It is the normalize.Product
// reduced to the fields the relational schema persists (total_cost is computed
// at query time, not stored — DATABASE_SCHEMA.md §1).
type Product struct {
	StoreID   string // scraper store id (config identifier), used to resolve the store row
	Store     string // human-readable store name (stores.name, the unique key)
	BaseURL   string // best-effort store root for stores.base_url
	Title     string
	URL       string
	Affiliate string
	ImageURL  string
	Currency  string
	Price     float64
	Shipping  float64
	ScrapedAt time.Time
}

// Repository persists normalized products.
type Repository interface {
	// UpsertProduct inserts or updates a single product and returns its row id.
	UpsertProduct(ctx context.Context, p Product) (string, error)
	// Close releases any underlying resources.
	Close() error
}
