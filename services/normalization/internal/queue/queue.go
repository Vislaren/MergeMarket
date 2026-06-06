// Package queue defines the input transport for the normalization-service.
// Workers pull RawResult items (produced by the scraper-service A-05) from a
// Source. The transport is abstracted behind an interface so the worker pool can
// be unit-tested without Redis; a Redis-backed implementation (see redis.go) is
// used in production.
//
// The RawResult / RawProduct types intentionally mirror the scraper-service
// contract (services/scraper-service/internal/queue). They are duplicated rather
// than shared because each service is its own Go module; this is the documented
// cross-service wire contract for the `normalize_queue` Redis list.
package queue

import (
	"context"
	"errors"
	"time"
)

// ErrEmpty is returned by Dequeue when no result became available before the
// poll timeout elapsed. It is an expected, non-fatal condition (the worker
// loops).
var ErrEmpty = errors.New("queue: empty")

// RawProduct is a single product extracted from a store, before normalization.
type RawProduct struct {
	ProductID string  `json:"product_id"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`
	Shipping  float64 `json:"shipping"`
	ImageURL  string  `json:"image_url"`
	URL       string  `json:"url"`
}

// RawResult is the output of one scrape job, consumed from the normalization
// queue.
type RawResult struct {
	JobID     string       `json:"job_id,omitempty"`
	StoreID   string       `json:"store_id"`
	Store     string       `json:"store"`
	Query     string       `json:"query"`
	Location  string       `json:"location,omitempty"`
	Products  []RawProduct `json:"products"`
	ScrapedAt time.Time    `json:"scraped_at"`
}

// Source is the stream of raw scrape results awaiting normalization.
type Source interface {
	// Dequeue blocks for up to the source's poll timeout for the next result. It
	// returns ErrEmpty if none arrived, or ctx.Err() if the context is done.
	Dequeue(ctx context.Context) (RawResult, error)
	// Close releases any underlying resources.
	Close() error
}
