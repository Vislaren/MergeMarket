package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres is a pgx-backed, read-only Repository over the shared PostgreSQL
// instance (DATABASE_SCHEMA.md). It reads the products and stores relational
// tables the normalization service (A-06) writes; it never mutates them.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres connects to the database at dsn and pings to fail fast.
func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("store: connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("store: ping: %w", err)
	}
	return &Postgres{pool: pool}, nil
}

// Search runs a case-insensitive title match across normalized products, joining
// the store name, and returns the cheapest-by-total-cost results first.
//
// location is accepted for contract compatibility but is currently advisory:
// neither products nor stores carry a country column in DATABASE_SCHEMA.md, so
// geo-filtering is deferred until a store→country mapping exists. The parameter
// is part of the cache key (different locations get distinct cache entries), so
// adding real filtering later will not break cached callers.
func (p *Postgres) Search(ctx context.Context, query, _ string, limit int) ([]Product, error) {
	const q = `
		SELECT
			pr.id::text,
			pr.title,
			COALESCE(pr.last_price, 0),
			pr.currency,
			COALESCE(pr.last_shipping, 0),
			COALESCE(pr.last_price, 0) + COALESCE(pr.last_shipping, 0) AS total_cost,
			COALESCE(pr.image_url, ''),
			s.name,
			COALESCE(pr.affiliate_url, pr.url),
			COALESCE(pr.scraped_at, pr.created_at)
		FROM products pr
		JOIN stores s ON s.id = pr.store_id
		WHERE pr.last_price IS NOT NULL
		  AND pr.title ILIKE '%' || $1 || '%'
		ORDER BY total_cost ASC
		LIMIT $2`
	rows, err := p.pool.Query(ctx, q, query, limit)
	if err != nil {
		return nil, fmt.Errorf("store: search: %w", err)
	}
	defer rows.Close()

	results := make([]Product, 0, limit)
	for rows.Next() {
		var pr Product
		if err := rows.Scan(
			&pr.ProductID, &pr.Title, &pr.Price, &pr.Currency, &pr.Shipping,
			&pr.TotalCost, &pr.ImageURL, &pr.Store, &pr.AffiliateURL, &pr.ScrapedAt,
		); err != nil {
			return nil, fmt.Errorf("store: scan result: %w", err)
		}
		results = append(results, pr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate results: %w", err)
	}
	return results, nil
}

// Close releases the connection pool.
func (p *Postgres) Close() error {
	p.pool.Close()
	return nil
}
