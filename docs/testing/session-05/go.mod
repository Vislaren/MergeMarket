// Module for the Session-05 test suites (unit + integration) covering Agent B's
// B-02 mock-server service.
//
// Note on layout: the mock server's logic lives in `internal/` packages, which
// Go forbids importing from a different module. So — exactly as the Session-04
// suite did for A-04 — the UNIT suite validates the service via subprocess (go
// build/test/vet/gofmt) plus structural source assertions, and the INTEGRATION
// suite builds and drives the compiled binary end-to-end over HTTP. Both are
// independent QE verifications designed from the oracle (API_CONTRACTS.md), not
// a copy of the developer's in-package tests.
//
// Intentionally dependency-free (standard library only): the mock server needs
// no Redis/network, so the whole suite — unit AND integration — runs offline.
module github.com/Vislaren/MergeMarket/docs/testing/session-05

go 1.22
