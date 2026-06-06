// Package queue defines the job/result transport for the scraper. Workers pull
// Jobs from a Queue, scrape, and publish a RawResult to a Sink (the
// normalization queue). The transport is abstracted behind interfaces so the
// worker pool can be unit-tested without Redis; a Redis-backed implementation
// (see redis.go) is used in production.
package queue

import (
	"context"
	"errors"
	"time"
)

// ErrEmpty is returned by Dequeue when no job became available before the poll
// timeout elapsed. It is an expected, non-fatal condition (the worker loops).
var ErrEmpty = errors.New("queue: empty")

// Job is a single unit of scraping work: search a store for a query.
type Job struct {
	// JobID correlates with the scrape_jobs row (DATABASE_SCHEMA.md §1); may be
	// empty for ad-hoc jobs.
	JobID string `json:"job_id,omitempty"`
	// StoreID selects which StoreConfig the worker loads to perform the scrape.
	StoreID string `json:"store_id"`
	// Query is the user's search term.
	Query string `json:"query"`
	// Location is an ISO 3166-1 alpha-2 country code (e.g. "CM").
	Location string `json:"location,omitempty"`
}

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

// RawResult is the output of one scrape job, pushed to the normalization queue.
type RawResult struct {
	JobID     string       `json:"job_id,omitempty"`
	StoreID   string       `json:"store_id"`
	Store     string       `json:"store"`
	Query     string       `json:"query"`
	Location  string       `json:"location,omitempty"`
	Products  []RawProduct `json:"products"`
	ScrapedAt time.Time    `json:"scraped_at"`
}

// Queue is the source of scrape jobs.
type Queue interface {
	// Dequeue blocks for up to the queue's poll timeout for the next job. It
	// returns ErrEmpty if none arrived, or ctx.Err() if the context is done.
	Dequeue(ctx context.Context) (Job, error)
	// Close releases any underlying resources.
	Close() error
}

// Sink receives raw scrape results for the normalization service.
type Sink interface {
	// Publish enqueues one raw result.
	Publish(ctx context.Context, result RawResult) error
}
