// Module for the Session-04 test suites (unit + integration) covering Agent A's
// A-04 proxy-validator service.
//
// Note on layout: A-04's logic lives in `internal/` packages, which Go forbids
// importing from a different module. So — exactly as the Session-02 suite did
// for A-02 — the UNIT suite validates the service via subprocess (go
// build/test/vet/gofmt) plus structural source assertions, and the INTEGRATION
// suite drives the compiled binary end-to-end over HTTP + Redis. Both are
// independent QE verifications designed from the oracle, not a copy of the
// developer's in-package tests.
//
// Run `go mod tidy` to populate go.sum in CI before `go test`.
module github.com/Vislaren/MergeMarket/docs/testing/session-04

go 1.22

require (
	github.com/redis/go-redis/v9 v9.6.0
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
