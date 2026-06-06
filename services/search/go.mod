// Search Service (A-14) — serves GET /api/v1/search from normalized products in
// PostgreSQL (written by the normalization service, A-06) with the
// search:{query_hash} Redis cache (DATABASE_SCHEMA.md §3) using a
// stale-while-revalidate strategy (ARCHITECTURE.md §10). Follows the
// module/layout convention established by A-04..A-08.
// Run `go mod tidy` to populate go.sum before build/test.
module github.com/Vislaren/MergeMarket/services/search

go 1.22

require (
	github.com/jackc/pgx/v5 v5.6.0
	github.com/redis/go-redis/v9 v9.6.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)
