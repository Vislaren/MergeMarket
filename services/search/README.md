# Search Service (A-14)

Serves the public product search endpoint from normalized data already in
PostgreSQL (written by the normalization service, A-06), fronted by a
stale-while-revalidate Redis cache.

## Endpoints

| Method | Path | Notes |
|--------|------|-------|
| `GET`  | `/api/v1/search?q={query}&location={cc}` | API_CONTRACTS.md search contract |
| `GET`  | `/health` | `{status, service, version}` |

`q` and `location` are both required (`400 missing_query` otherwise). The
response carries `cached` and `latency_ms` per the contract.

## How it works

1. Build the cache key `search:{sha256(lower(q)|lower(location))}`
   (DATABASE_SCHEMA.md §3).
2. **Fresh cache hit** → return immediately (`cached: true`).
3. **Stale hit** (older than `SEARCH_CACHE_STALE_AFTER`) → return the stale
   results immediately *and* refresh in the background
   (stale-while-revalidate, ARCHITECTURE.md §10).
4. **Miss** → read the cheapest-first matching products from Postgres, assign a
   `deal_score`, write the cache (`SEARCH_CACHE_TTL`), return (`cached: false`).

### Deal Meter (`deal_score`)

A deterministic 0–100 heuristic: within a result set the cheapest total cost
scores 100, the dearest 0, the rest scale linearly. This is a placeholder until
a richer Deal Meter (price-history- and review-aware) is built — see the
truth-score service (A-15).

## Configuration

All via environment variables (see `.env.example`): `SEARCH_PORT` (8087),
`DB_*` / `DATABASE_URL`, `REDIS_*`, `SEARCH_CACHE_PREFIX`, `SEARCH_CACHE_TTL`,
`SEARCH_CACHE_STALE_AFTER`, `SEARCH_MAX_RESULTS`.

## Known limitations

- **`location` is advisory.** Neither `products` nor `stores` carries a country
  column yet, so geo-filtering is deferred. It is still part of the cache key, so
  enabling real filtering later won't break cached callers.
- Live DB/Redis paths are exercised at runtime; unit tests cover config, the
  cache key, the orchestrator (with fakes), the Deal Meter, and the HTTP layer.
- No `/metrics` endpoint yet — consistent with the other A-04..A-08 services;
  the Prometheus mandate in PORTS_README is an open cross-service follow-up.
