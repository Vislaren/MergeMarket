## Test Plan — Unit — Session 05 — B-02 Mock Server for All API Contracts

**Scope:** Static and subprocess verification of `services/mock-server/` against
the contract-level facts in `project_docs/api/API_CONTRACTS.md` and the B-02 task
spec: the service compiles and its own tests/vet/gofmt are clean; configuration is
environment-driven with the mandated default port 8080; `/health` returns the
contract shape; every API_CONTRACTS endpoint is routed; the search
`total_cost = price + shipping` invariant holds in fixture code; the canonical
`{error, message}` error shape is used; logging is via `log/slog`; the module is
dependency-free; and every package carries a GoDoc comment.

**Out of scope:** Live HTTP behaviour (covered by the integration suite); the
Flutter/BFF consumers (B-03+/B-09); SonarQube gate execution (no CI/VPS scan yet).

**Approach:** Unit tests isolate the service via subprocess (`go build/test/vet`,
`gofmt -l`) and structural source assertions. A separate test module is used
because Go forbids importing another module's `internal/` packages — the same
constraint that shaped the Session-02 and Session-04 unit suites.

**Entry criteria:** `services/mock-server/` exists and a Go 1.22+ toolchain is on
PATH.

**Exit criteria:** All 12 unit cases pass; toolchain gates (build/test/vet/gofmt)
are clean.

**Tools:** Go `testing` (stdlib), `os/exec`, `gofmt`. Dependency-free.

**Assumptions:** The mock server is intentionally stdlib-only and stateless;
error paths are driven by deterministic sentinel inputs rather than stored state.

**Risk:** Structural (string-match) assertions can drift if the source is
reformatted in a way that changes the matched literal (e.g. the `total_cost`
expression). Mitigated by also gating on `gofmt`, which fixes the canonical form,
and by the live integration suite that verifies behaviour independently.
