// History Service (A-07) — records daily price snapshots into the TimescaleDB
// price_history hypertable, runs a scheduled heartbeat that re-checks followed
// (alerted) products and fires price-drop alerts on a downward threshold
// crossing, and serves the product price-history API. Follows the
// module/layout convention established by the scraper-service (A-05) and
// normalization-service (A-06).
// Run `go mod tidy` to populate go.sum before build/test.
module github.com/Vislaren/MergeMarket/services/history

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
