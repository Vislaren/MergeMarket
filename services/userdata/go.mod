// User-data Service (A-16..A-18) — JWT-protected per-user CRUD behind Kong→BFF
// for wishlist, price alerts, and the savings dashboard. Consolidates three
// API_CONTRACTS.md resource groups that share the same per-user Postgres CRUD
// pattern (Kong/A-09 already routes all three paths to one upstream). It verifies
// the same HS256 JWT the auth service (A-08) issues, so it shares JWT_SECRET.
// Follows the module/layout convention established by A-04..A-08.
// Run `go mod tidy` to populate go.sum before build/test.
module github.com/Vislaren/MergeMarket/services/userdata

go 1.22

require github.com/jackc/pgx/v5 v5.6.0

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)
