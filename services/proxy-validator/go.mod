// Proxy Validator service (A-04) — first Go service in MergeMarket; sets the
// module/layout convention for the remaining backend services (A-05..A-08).
// Run `go mod tidy` to populate go.sum before `go build` / `go test`.
module github.com/Vislaren/MergeMarket/services/proxy-validator

go 1.22

require github.com/redis/go-redis/v9 v9.6.0

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
)
