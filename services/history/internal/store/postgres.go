package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres is a pgx-backed Repository over the shared PostgreSQL/TimescaleDB
// instance (DATABASE_SCHEMA.md). The history-service only reads relational
// tables (products, stores, price_alerts) and reads/writes the price_history
// hypertable; it never mutates the relational rows it does not own.
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

// SnapshotAll copies the current price of every priced product into
// price_history in a single statement, stamped NOW().
func (p *Postgres) SnapshotAll(ctx context.Context) (int64, error) {
	const q = `
		INSERT INTO price_history (product_id, price, shipping, currency, recorded_at)
		SELECT id, last_price, last_shipping, currency, NOW()
		FROM products
		WHERE last_price IS NOT NULL`
	tag, err := p.pool.Exec(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("store: snapshot all: %w", err)
	}
	return tag.RowsAffected(), nil
}

// FollowedProducts returns products that have at least one active alert, with
// their alerts grouped. One row per (product, alert) is read and grouped in Go.
func (p *Postgres) FollowedProducts(ctx context.Context) ([]Followed, error) {
	const q = `
		SELECT p.id, p.title, p.url, s.name, p.currency, p.last_price, p.last_shipping,
		       pa.id, pa.threshold_price, pa.currency
		FROM price_alerts pa
		JOIN products p ON p.id = pa.product_id
		JOIN stores   s ON s.id = p.store_id
		WHERE pa.is_active = TRUE
		ORDER BY p.id`
	rows, err := p.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("store: followed products: %w", err)
	}
	defer rows.Close()

	var (
		out  []Followed
		byID = map[string]int{} // product id → index in out
	)
	for rows.Next() {
		var (
			pid, title, url, storeName, currency string
			lastPrice, lastShipping              *float64
			alertID, alertCurrency               string
			threshold                            float64
		)
		if err := rows.Scan(&pid, &title, &url, &storeName, &currency,
			&lastPrice, &lastShipping, &alertID, &threshold, &alertCurrency); err != nil {
			return nil, fmt.Errorf("store: scan followed: %w", err)
		}

		idx, ok := byID[pid]
		if !ok {
			f := Followed{
				ProductID: pid,
				Title:     title,
				URL:       url,
				Store:     storeName,
				Currency:  currency,
			}
			if lastPrice != nil {
				f.LastPrice = *lastPrice
				f.HasPrice = true
			}
			if lastShipping != nil {
				f.LastShipping = *lastShipping
			}
			out = append(out, f)
			idx = len(out) - 1
			byID[pid] = idx
		}
		out[idx].Alerts = append(out[idx].Alerts, Alert{
			ID:        alertID,
			Threshold: threshold,
			Currency:  alertCurrency,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate followed: %w", err)
	}
	return out, nil
}

// LatestPrice returns the most recent recorded price for a product.
func (p *Postgres) LatestPrice(ctx context.Context, productID string) (float64, bool, error) {
	const q = `
		SELECT price FROM price_history
		WHERE product_id = $1
		ORDER BY recorded_at DESC
		LIMIT 1`
	var price float64
	err := p.pool.QueryRow(ctx, q, productID).Scan(&price)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("store: latest price: %w", err)
	}
	return price, true, nil
}

// InsertSnapshot records one price observation.
func (p *Postgres) InsertSnapshot(ctx context.Context, productID string, price float64, shipping *float64, currency string, at time.Time) error {
	const q = `
		INSERT INTO price_history (product_id, price, shipping, currency, recorded_at)
		VALUES ($1, $2, $3, $4, $5)`
	if _, err := p.pool.Exec(ctx, q, productID, price, shipping, currencyCode(currency), at); err != nil {
		return fmt.Errorf("store: insert snapshot %s: %w", productID, err)
	}
	return nil
}

// History returns the product's title, its price points over the last 6 months,
// and the 6-month average / 30-day low. A product with no products row yields
// ErrNotFound.
func (p *Postgres) History(ctx context.Context, productID string) (HistoryResult, error) {
	var title string
	if err := p.pool.QueryRow(ctx, `SELECT title FROM products WHERE id = $1`, productID).Scan(&title); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return HistoryResult{}, ErrNotFound
		}
		return HistoryResult{}, fmt.Errorf("store: history title: %w", err)
	}

	res := HistoryResult{ProductID: productID, Title: title, History: []HistoryPoint{}}

	const pointsQ = `
		SELECT price, currency, recorded_at
		FROM price_history
		WHERE product_id = $1 AND recorded_at >= NOW() - INTERVAL '6 months'
		ORDER BY recorded_at`
	rows, err := p.pool.Query(ctx, pointsQ, productID)
	if err != nil {
		return HistoryResult{}, fmt.Errorf("store: history points: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var pt HistoryPoint
		if err := rows.Scan(&pt.Price, &pt.Currency, &pt.RecordedAt); err != nil {
			return HistoryResult{}, fmt.Errorf("store: scan history point: %w", err)
		}
		res.History = append(res.History, pt)
	}
	if err := rows.Err(); err != nil {
		return HistoryResult{}, fmt.Errorf("store: iterate history: %w", err)
	}

	const aggQ = `
		SELECT
			COALESCE(AVG(price) FILTER (WHERE recorded_at >= NOW() - INTERVAL '6 months'), 0),
			COALESCE(MIN(price) FILTER (WHERE recorded_at >= NOW() - INTERVAL '30 days'), 0)
		FROM price_history
		WHERE product_id = $1`
	if err := p.pool.QueryRow(ctx, aggQ, productID).Scan(&res.Average6m, &res.Lowest30d); err != nil {
		return HistoryResult{}, fmt.Errorf("store: history aggregates: %w", err)
	}
	res.Average6m = round2(res.Average6m)
	res.Lowest30d = round2(res.Lowest30d)
	return res, nil
}

// Close releases the connection pool.
func (p *Postgres) Close() error {
	p.pool.Close()
	return nil
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
