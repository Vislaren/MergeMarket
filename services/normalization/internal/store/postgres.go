package store

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres is a pgx-backed Repository. It caches resolved store ids (name → uuid)
// in memory so a hot scrape stream does not re-hit the stores table for every
// product, and upserts products by (store_id, url).
//
// Upsert relies on a UNIQUE (store_id, url) constraint on products. The
// normalization service creates that index at startup via EnsureSchema; the
// canonical schema (infra/db/init/01-schema.sql) also declares it.
type Postgres struct {
	pool *pgxpool.Pool

	mu       sync.Mutex
	storeIDs map[string]string // stores.name → stores.id (uuid)
}

// NewPostgres connects to the database at dsn and returns a Postgres repository.
// It pings to fail fast on an unreachable database.
func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("store: connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("store: ping: %w", err)
	}
	return &Postgres{pool: pool, storeIDs: map[string]string{}}, nil
}

// EnsureSchema creates the UNIQUE (store_id, url) index the product upsert
// depends on, if it does not already exist. It is idempotent and safe to call on
// every startup.
func (p *Postgres) EnsureSchema(ctx context.Context) error {
	const ddl = `CREATE UNIQUE INDEX IF NOT EXISTS idx_products_store_url ON products(store_id, url)`
	if _, err := p.pool.Exec(ctx, ddl); err != nil {
		return fmt.Errorf("store: ensure schema: %w", err)
	}
	return nil
}

// UpsertProduct resolves the store row, then inserts or updates the product by
// (store_id, url), refreshing price/shipping/affiliate/image/title/scraped_at.
func (p *Postgres) UpsertProduct(ctx context.Context, prod Product) (string, error) {
	storeID, err := p.ensureStore(ctx, prod)
	if err != nil {
		return "", err
	}

	const q = `
		INSERT INTO products
			(store_id, title, url, affiliate_url, image_url, currency, last_price, last_shipping, scraped_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (store_id, url) DO UPDATE SET
			title         = EXCLUDED.title,
			affiliate_url = EXCLUDED.affiliate_url,
			image_url     = EXCLUDED.image_url,
			currency      = EXCLUDED.currency,
			last_price    = EXCLUDED.last_price,
			last_shipping = EXCLUDED.last_shipping,
			scraped_at    = EXCLUDED.scraped_at
		RETURNING id`

	var id string
	err = p.pool.QueryRow(ctx, q,
		storeID, prod.Title, prod.URL, nullable(prod.Affiliate), nullable(prod.ImageURL),
		currencyCode(prod.Currency), prod.Price, prod.Shipping, prod.ScrapedAt,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("store: upsert product %q: %w", prod.URL, err)
	}
	return id, nil
}

// ensureStore returns the stores.id for a product's store, upserting the stores
// row by its unique name and caching the result.
func (p *Postgres) ensureStore(ctx context.Context, prod Product) (string, error) {
	name := strings.TrimSpace(prod.Store)
	if name == "" {
		name = prod.StoreID // fall back to the config id when no display name
	}
	if name == "" {
		return "", fmt.Errorf("store: product has neither store name nor store id")
	}

	p.mu.Lock()
	if id, ok := p.storeIDs[name]; ok {
		p.mu.Unlock()
		return id, nil
	}
	p.mu.Unlock()

	const q = `
		INSERT INTO stores (name, base_url, config_path)
		VALUES ($1, $2, $3)
		ON CONFLICT (name) DO UPDATE SET base_url = EXCLUDED.base_url
		RETURNING id`
	var id string
	if err := p.pool.QueryRow(ctx, q, name, prod.BaseURL, prod.StoreID).Scan(&id); err != nil {
		return "", fmt.Errorf("store: ensure store %q: %w", name, err)
	}

	p.mu.Lock()
	p.storeIDs[name] = id
	p.mu.Unlock()
	return id, nil
}

// Close releases the connection pool.
func (p *Postgres) Close() error {
	p.pool.Close()
	return nil
}

// nullable returns nil for an empty string so the column is stored as NULL
// rather than an empty string (affiliate_url and image_url are nullable).
func nullable(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// currencyCode guarantees a 3-char value for the CHAR(3) currency column.
func currencyCode(c string) string {
	c = strings.ToUpper(strings.TrimSpace(c))
	if len(c) != 3 {
		return "USD"
	}
	return c
}
