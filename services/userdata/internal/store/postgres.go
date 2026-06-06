package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgreSQL error codes used for mapping driver errors onto domain errors.
const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
	pgInvalidTextRepr     = "22P02" // e.g. a malformed UUID literal
)

// Postgres is a pgx-backed Repository over the shared PostgreSQL instance.
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

// EnsureSchema idempotently creates the purchases table this service owns, so the
// service is self-sufficient against a database initialised before the savings
// dashboard existed (mirrors the canonical infra/db/init/01-schema.sql).
func (p *Postgres) EnsureSchema(ctx context.Context) error {
	const ddl = `
		CREATE TABLE IF NOT EXISTS purchases (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			saved      NUMERIC(12, 2) NOT NULL DEFAULT 0,
			currency   CHAR(3) NOT NULL DEFAULT 'USD',
			bought_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_purchases_user_id ON purchases(user_id);`
	if _, err := p.pool.Exec(ctx, ddl); err != nil {
		return fmt.Errorf("store: ensure schema: %w", err)
	}
	return nil
}

// Close releases the connection pool.
func (p *Postgres) Close() error {
	p.pool.Close()
	return nil
}

// ── Wishlist ────────────────────────────────────────────────────────────────

// AddWishlist inserts a wishlist row for the user. A duplicate yields
// ErrAlreadyExists; an unknown/malformed product_id yields ErrUnknownProduct.
func (p *Postgres) AddWishlist(ctx context.Context, userID, productID string) (WishlistAdded, error) {
	const q = `
		INSERT INTO wishlist_items (user_id, product_id)
		VALUES ($1, $2)
		RETURNING id::text, added_at`
	var out WishlistAdded
	err := p.pool.QueryRow(ctx, q, userID, productID).Scan(&out.WishlistID, &out.AddedAt)
	if err != nil {
		switch pgCode(err) {
		case pgUniqueViolation:
			return WishlistAdded{}, ErrAlreadyExists
		case pgForeignKeyViolation, pgInvalidTextRepr:
			return WishlistAdded{}, ErrUnknownProduct
		}
		return WishlistAdded{}, fmt.Errorf("store: add wishlist: %w", err)
	}
	return out, nil
}

// ListWishlist returns the user's wishlist with a per-store price comparison for
// each item (every store that sells a product with the same title).
func (p *Postgres) ListWishlist(ctx context.Context, userID string) ([]WishlistItem, error) {
	const q = `
		SELECT w.id::text, p.id::text, p.title, COALESCE(p.image_url, ''), w.added_at
		FROM wishlist_items w
		JOIN products p ON p.id = w.product_id
		WHERE w.user_id = $1
		ORDER BY w.added_at DESC`
	rows, err := p.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("store: list wishlist: %w", err)
	}
	defer rows.Close()

	var items []WishlistItem
	for rows.Next() {
		var it WishlistItem
		if err := rows.Scan(&it.WishlistID, &it.ProductID, &it.Title, &it.ImageURL, &it.AddedAt); err != nil {
			return nil, fmt.Errorf("store: scan wishlist: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate wishlist: %w", err)
	}

	// Attach the price comparison per item. Wishlists are small, so a per-item
	// lookup is acceptable and keeps the query simple.
	for i := range items {
		stores, err := p.storesForTitle(ctx, items[i].Title)
		if err != nil {
			return nil, err
		}
		items[i].Stores = stores
	}
	return items, nil
}

func (p *Postgres) storesForTitle(ctx context.Context, title string) ([]StorePrice, error) {
	const q = `
		SELECT s.name,
		       COALESCE(p.last_price, 0),
		       COALESCE(p.last_price, 0) + COALESCE(p.last_shipping, 0)
		FROM products p
		JOIN stores s ON s.id = p.store_id
		WHERE p.title = $1 AND p.last_price IS NOT NULL
		ORDER BY (COALESCE(p.last_price, 0) + COALESCE(p.last_shipping, 0)) ASC`
	rows, err := p.pool.Query(ctx, q, title)
	if err != nil {
		return nil, fmt.Errorf("store: stores for title: %w", err)
	}
	defer rows.Close()

	stores := []StorePrice{}
	for rows.Next() {
		var sp StorePrice
		if err := rows.Scan(&sp.Store, &sp.Price, &sp.TotalCost); err != nil {
			return nil, fmt.Errorf("store: scan store price: %w", err)
		}
		stores = append(stores, sp)
	}
	return stores, rows.Err()
}

// RemoveWishlist deletes a wishlist row owned by the user, or ErrNotFound.
func (p *Postgres) RemoveWishlist(ctx context.Context, userID, wishlistID string) error {
	const q = `DELETE FROM wishlist_items WHERE id = $1 AND user_id = $2`
	tag, err := p.pool.Exec(ctx, q, wishlistID, userID)
	if err != nil {
		if pgCode(err) == pgInvalidTextRepr {
			return ErrNotFound
		}
		return fmt.Errorf("store: remove wishlist: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ── Alerts ──────────────────────────────────────────────────────────────────

// CreateAlert inserts a price alert for the user. An unknown/malformed
// product_id yields ErrUnknownProduct.
func (p *Postgres) CreateAlert(ctx context.Context, userID, productID string, threshold float64, currency string) (AlertCreated, error) {
	const q = `
		INSERT INTO price_alerts (user_id, product_id, threshold_price, currency)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, created_at`
	var out AlertCreated
	err := p.pool.QueryRow(ctx, q, userID, productID, threshold, currencyCode(currency)).Scan(&out.AlertID, &out.CreatedAt)
	if err != nil {
		switch pgCode(err) {
		case pgForeignKeyViolation, pgInvalidTextRepr:
			return AlertCreated{}, ErrUnknownProduct
		}
		return AlertCreated{}, fmt.Errorf("store: create alert: %w", err)
	}
	return out, nil
}

// ListAlerts returns the user's alerts, newest first.
func (p *Postgres) ListAlerts(ctx context.Context, userID string) ([]Alert, error) {
	const q = `
		SELECT a.id::text, p.id::text, p.title, a.threshold_price, a.currency, a.is_active, a.created_at
		FROM price_alerts a
		JOIN products p ON p.id = a.product_id
		WHERE a.user_id = $1
		ORDER BY a.created_at DESC`
	rows, err := p.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("store: list alerts: %w", err)
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var a Alert
		if err := rows.Scan(&a.AlertID, &a.ProductID, &a.Title, &a.ThresholdPrice, &a.Currency, &a.IsActive, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("store: scan alert: %w", err)
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// DeleteAlert deletes an alert owned by the user, or ErrNotFound.
func (p *Postgres) DeleteAlert(ctx context.Context, userID, alertID string) error {
	const q = `DELETE FROM price_alerts WHERE id = $1 AND user_id = $2`
	tag, err := p.pool.Exec(ctx, q, alertID, userID)
	if err != nil {
		if pgCode(err) == pgInvalidTextRepr {
			return ErrNotFound
		}
		return fmt.Errorf("store: delete alert: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ── Savings ─────────────────────────────────────────────────────────────────

// Savings returns the user's total saved and per-purchase transactions.
func (p *Postgres) Savings(ctx context.Context, userID string) (Savings, error) {
	const q = `
		SELECT pr.id::text, pr.title, pu.saved, pu.currency, pu.bought_at
		FROM purchases pu
		JOIN products pr ON pr.id = pu.product_id
		WHERE pu.user_id = $1
		ORDER BY pu.bought_at DESC`
	rows, err := p.pool.Query(ctx, q, userID)
	if err != nil {
		return Savings{}, fmt.Errorf("store: savings: %w", err)
	}
	defer rows.Close()

	out := Savings{Currency: "USD", Transactions: []Transaction{}}
	for rows.Next() {
		var (
			tx       Transaction
			currency string
		)
		if err := rows.Scan(&tx.ProductID, &tx.Title, &tx.Saved, &currency, &tx.BoughtAt); err != nil {
			return Savings{}, fmt.Errorf("store: scan saving: %w", err)
		}
		out.TotalSaved += tx.Saved
		out.Currency = currency // most recent purchase's currency (rows are newest-first)
		out.Transactions = append(out.Transactions, tx)
	}
	if err := rows.Err(); err != nil {
		return Savings{}, fmt.Errorf("store: iterate savings: %w", err)
	}
	out.TotalSaved = round2(out.TotalSaved)
	return out, nil
}

// ── helpers ─────────────────────────────────────────────────────────────────

func pgCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ""
	}
	return ""
}

func currencyCode(c string) string {
	if len(c) != 3 {
		return "USD"
	}
	return c
}

func round2(f float64) float64 {
	return float64(int64(f*100+0.5)) / 100
}
