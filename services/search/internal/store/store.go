// Package store is the search-service's read-only data-access layer over the
// normalized products PostgreSQL tables (written by the normalization service,
// A-06). Access is expressed behind the Repository interface so the search
// orchestrator can be unit-tested with a fake; the pgx-backed implementation
// lives in postgres.go.
package store

import (
	"context"
	"time"
)

// Product is one search result, matching the API_CONTRACTS.md search result
// item. DealScore is populated downstream by the search orchestrator (Deal
// Meter), not by the store.
type Product struct {
	ProductID    string    `json:"product_id"`
	Title        string    `json:"title"`
	Price        float64   `json:"price"`
	Currency     string    `json:"currency"`
	Shipping     float64   `json:"shipping"`
	TotalCost    float64   `json:"total_cost"`
	ImageURL     string    `json:"image_url"`
	Store        string    `json:"store"`
	AffiliateURL string    `json:"affiliate_url"`
	DealScore    int       `json:"deal_score"`
	ScrapedAt    time.Time `json:"scraped_at"`
}

// Repository is the search-service's persistence surface.
type Repository interface {
	// Search returns up to limit normalized products whose title matches the
	// query, cheapest total cost first. location is the ISO 3166-1 alpha-2
	// country code (currently advisory — see postgres.go).
	Search(ctx context.Context, query, location string, limit int) ([]Product, error)
	// Close releases any underlying resources.
	Close() error
}
