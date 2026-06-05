// Package store is the history-service's data-access layer over
// PostgreSQL/TimescaleDB. It writes price snapshots into the price_history
// hypertable, reads followed (alerted) products for the heartbeat, and serves
// the product price-history query (API_CONTRACTS.md). Access is expressed behind
// the Repository interface so the runner and HTTP handlers can be unit-tested
// with a fake; the pgx-backed implementation lives in postgres.go.
package store

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned by History when the product has no row.
var ErrNotFound = errors.New("store: product not found")

// Alert is one active price alert on a product.
type Alert struct {
	ID        string
	Threshold float64
	Currency  string
}

// Followed is a product that has at least one active price alert, with the
// alerts attached. LastPrice/LastShipping come from the products table (the
// freshest scraped values) and may be unset (HasPrice=false).
type Followed struct {
	ProductID    string
	Title        string
	URL          string
	Store        string
	Currency     string
	LastPrice    float64
	LastShipping float64
	HasPrice     bool
	Alerts       []Alert
}

// HistoryPoint is one recorded price observation.
type HistoryPoint struct {
	Price      float64   `json:"price"`
	Currency   string    `json:"currency"`
	RecordedAt time.Time `json:"recorded_at"`
}

// HistoryResult is the full price-history view for one product
// (API_CONTRACTS.md GET /products/{id}/history).
type HistoryResult struct {
	ProductID string         `json:"product_id"`
	Title     string         `json:"title"`
	History   []HistoryPoint `json:"history"`
	Average6m float64        `json:"average_6m"`
	Lowest30d float64        `json:"lowest_30d"`
}

// Repository is the history-service's persistence surface.
type Repository interface {
	// SnapshotAll records a price_history row for every product that currently
	// has a price, stamped now. Returns the number of rows written.
	SnapshotAll(ctx context.Context) (int64, error)
	// FollowedProducts returns every product with at least one active alert.
	FollowedProducts(ctx context.Context) ([]Followed, error)
	// LatestPrice returns the most recent recorded price for a product, and
	// whether any snapshot exists.
	LatestPrice(ctx context.Context, productID string) (float64, bool, error)
	// InsertSnapshot records one price observation for a product.
	InsertSnapshot(ctx context.Context, productID string, price float64, shipping *float64, currency string, at time.Time) error
	// History returns the price history for a product, or ErrNotFound.
	History(ctx context.Context, productID string) (HistoryResult, error)
	// Close releases any underlying resources.
	Close() error
}
