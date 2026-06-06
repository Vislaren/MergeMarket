# B-11 Contract Audit — Mock vs Real Backend

The B-11 task says "fix any contract mismatches discovered during integration."
Because the real stack cannot be run locally (no Docker; Agent A's services are
not deployed), this is a **static** audit: the B-02 mock fixtures and the Flutter
data layer were compared against Agent A's actual service code (read from
`origin/co-courage`) and the canonical `project_docs/api/API_CONTRACTS.md`.

## Endpoint coverage in the real backend

| Contract endpoint | Real service (Agent A) | Status for the swap |
|-------------------|------------------------|---------------------|
| `POST /api/v1/auth/{register,login,refresh}` | **auth** (A-08), `services/auth/internal/server/server.go` | ✅ Real — identical shape; safe to swap |
| `GET /api/v1/products/{id}/history` | **history** (A-07), `services/history/internal/server/server.go` | ✅ Real — identical shape; safe to swap |
| `GET /api/v1/products/{id}/detail` (aggregate) | **bff** (B-09) | ✅ Real — BFF aggregate; now forwards JWT |
| `GET /api/v1/search` | **none** | ❌ Mock-only — no real read/cache service exists |
| `GET /api/v1/products/{id}/truth-score` | **none** | ❌ Mock-only — no review-analysis service exists |
| `GET/POST/DELETE /api/v1/wishlist` | **none** | ❌ Mock-only — no wishlist service exists |
| `GET/POST/DELETE /api/v1/alerts` | **none** | ❌ Mock-only — history emits drop *events* to Redis but serves no alerts CRUD API |
| `GET /api/v1/savings` | **none** | ❌ Mock-only — no savings service exists |

Verified by grepping `origin/co-courage` for the route patterns across all
`services/**/*.go` (excluding the mock server): only auth and history register
client-facing `/api/v1/...` HTTP routes.

## Shape mismatches found

**None.** Where a real service exists (auth, history) its JSON shape matches both
the mock fixtures the app was built against and `API_CONTRACTS.md`. The Flutter
models decode tolerantly (safe zero-value fallbacks), so no client decode change
was required. The only contract-adjacent fix was a **config** drift: the app's
default base URL pointed at the blocked port `:8080`; corrected to the mock's
assigned `:8089` per `PORTS_README.md`.

## Conclusion / what blocks a full swap

The client- and BFF-side integration work for B-11 is complete and correct:
every protected call now carries a Bearer token, expiry is handled by the
refresh-on-401 interceptor, and the BFF forwards the JWT to upstream. Against the
real backend the app can today fully exercise **auth** and **price history**
(and the BFF **detail** aggregate over them).

The full search → results → wishlist → alert → notification E2E cannot complete
because **the real services behind search, truth-score, wishlist, alerts and
savings have not been built yet** (Agent A backlog), and the local stack is not
runnable (no Docker). Until those exist, `BFF_UPSTREAM_URL` / `API_BASE_URL` for
those routes must remain the B-02 mock server.

### Recommended follow-up tasks (Agent A)
- Search read service: serve `GET /api/v1/search` from normalized Postgres data
  with the `search:{query_hash}` Redis cache (the `cached` / `latency_ms` fields
  in the contract imply it). A-06 explicitly deferred the cache write.
- Truth-Score service (review authenticity).
- Wishlist, Alerts, and Savings persistence services (CRUD behind Kong → BFF).

These are recorded here so the integration gap is not lost; they are out of
Agent B's scope.
