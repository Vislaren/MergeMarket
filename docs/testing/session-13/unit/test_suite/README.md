# Unit test suite — Session 13 (B-11)

The executable B-11 unit tests live **in-tree** so they run under the normal
`flutter test` / `go test` invocations and in CI — they are not duplicated here
to avoid drift. This folder is a pointer to the canonical sources.

| Cases | File | Run |
|-------|------|-----|
| TC-13-U-001…010 | `apps/mobile/test/unit/authenticated_client_test.dart` | `cd apps/mobile && flutter test test/unit/authenticated_client_test.dart` |
| TC-13-U-011 (BFF) | `services/bff/internal/server/server_test.go` → `TestProductDetailForwardsAuthHeaderUpstream` | `cd services/bff && go test ./internal/server/ -run TestProductDetailForwardsAuthHeaderUpstream` |

Full regression: `cd apps/mobile && flutter test` → **151/151 PASS**;
`cd services/bff && go test ./...` → PASS.
