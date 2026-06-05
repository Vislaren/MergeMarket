package store

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestPostgres_UpsertProduct exercises the real pgx path against a live
// database. It is skipped unless DB_TEST_DSN is set, so unit runs and CI without
// a database stay green (mirrors the proxy-validator's live-Redis test gating).
//
// Example:
//
//	DB_TEST_DSN="postgres://postgres:pass@localhost:5432/mergemarket?sslmode=disable" go test ./internal/store/...
func TestPostgres_UpsertProduct(t *testing.T) {
	dsn := os.Getenv("DB_TEST_DSN")
	if dsn == "" {
		t.Skip("DB_TEST_DSN not set; skipping live database test")
	}

	ctx := context.Background()
	repo, err := NewPostgres(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer repo.Close()

	if err := repo.EnsureSchema(ctx); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	p := Product{
		StoreID:   "test-store",
		Store:     "Test Store",
		BaseURL:   "https://test.example",
		Title:     "Widget",
		URL:       "https://test.example/p/" + time.Now().Format("150405.000"),
		Affiliate: "https://test.example/p/1?aff=mm",
		ImageURL:  "https://test.example/img.png",
		Currency:  "USD",
		Price:     12.34,
		Shipping:  1.00,
		ScrapedAt: time.Now().UTC(),
	}

	id1, err := repo.UpsertProduct(ctx, p)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	// Upserting the same (store_id, url) updates rather than duplicating.
	p.Price = 9.99
	id2, err := repo.UpsertProduct(ctx, p)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if id1 != id2 {
		t.Errorf("upsert created a new row: %s != %s", id1, id2)
	}
}
