package store

import (
	"context"
	"os"
	"testing"
)

// These tests exercise the real pgx path against a live PostgreSQL/TimescaleDB.
// They are skipped unless DB_TEST_DSN is set, so unit runs and CI without a
// database stay green (mirrors the proxy-validator's live-Redis test gating).
//
//	DB_TEST_DSN="postgres://postgres:pass@localhost:5432/mergemarket?sslmode=disable" go test ./internal/store/...
func liveRepo(t *testing.T) *Postgres {
	t.Helper()
	dsn := os.Getenv("DB_TEST_DSN")
	if dsn == "" {
		t.Skip("DB_TEST_DSN not set; skipping live database test")
	}
	repo, err := NewPostgres(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { repo.Close() })
	return repo
}

func TestPostgres_SnapshotAndHistory(t *testing.T) {
	repo := liveRepo(t)
	ctx := context.Background()

	// SnapshotAll should run without error (count depends on seeded data).
	if _, err := repo.SnapshotAll(ctx); err != nil {
		t.Fatalf("snapshot all: %v", err)
	}

	// FollowedProducts should run without error.
	if _, err := repo.FollowedProducts(ctx); err != nil {
		t.Fatalf("followed products: %v", err)
	}

	// History of a non-existent product is ErrNotFound.
	if _, err := repo.History(ctx, "00000000-0000-0000-0000-000000000000"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	// LatestPrice of a non-existent product is (0,false,nil).
	if _, ok, err := repo.LatestPrice(ctx, "00000000-0000-0000-0000-000000000000"); err != nil || ok {
		t.Fatalf("latest price: ok=%v err=%v", ok, err)
	}
}
