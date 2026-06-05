// Scraper Service (A-05) — the config-driven worker-queue scraping engine.
// Follows the module/layout convention established by the proxy-validator
// service (A-04). Run `go mod tidy` to populate go.sum before build/test.
module github.com/Vislaren/MergeMarket/services/scraper-service

go 1.22

require (
	github.com/redis/go-redis/v9 v9.6.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
)
