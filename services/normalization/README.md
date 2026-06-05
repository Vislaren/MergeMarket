# Normalization Service (Go) — A-06

Consumes raw scrape results from the `normalize_queue` Redis list (produced by
the scraper-service, A-05), converts each product into the canonical MergeMarket
schema, injects retailer-specific **affiliate links** (FR-6), and upserts the
results into PostgreSQL.

- **Port:** `8084` (`NORMALIZATION_PORT`)
- **Endpoints:** `GET /health` (API contract shape), `GET /stats` (worker-pool counters)
- **Module:** `github.com/Vislaren/MergeMarket/services/normalization`

## Pipeline

```
Redis list (normalize_queue)            ← RawResult JSON from scraper-service (A-05)
        │  BLPOP
        ▼
   worker pool
        │  normalize.FromRaw  (trim/collapse title, validate price>0,
        │                       clamp shipping, ISO currency, total_cost)
        ▼
   affiliate.Inject  (deep-link template and/or query params per store_id)
        │
        ▼
   PostgreSQL  (upsert stores by name → upsert products by (store_id, url),
                refreshing last_price / last_shipping / scraped_at)
```

| Package | Responsibility |
|---|---|
| `internal/config` | Env-driven service configuration (`Load`). |
| `internal/queue` | `RawResult`/`RawProduct` wire types, `Source` interface, Redis + in-memory impls. |
| `internal/normalize` | Pure transform: raw product → canonical `Product` (drops invalid items, NFR-2). |
| `internal/affiliate` | Affiliate Link Injection: data-driven per-store templates/params (FR-6). |
| `internal/store` | `Repository` interface + pgx-backed upsert of stores and products. |
| `internal/worker` | Worker pool: dequeue → normalize → inject → persist, with stats. |
| `internal/server` | `GET /health` + `GET /stats`, graceful shutdown. |

## Canonical product schema

Each raw product becomes the Search API result item shape
(`API_CONTRACTS.md`): `product_id, title, price, currency, shipping,
total_cost, image_url, store, affiliate_url, scraped_at`. `total_cost = price +
shipping`. Items without a title or with a non-positive price are skipped so a
single malformed row never fails the batch (NFR-2).

## Persistence

- A scraped store is resolved to a `stores` row by its unique `name` (upsert),
  with `config_path` set to the scraper store id and `base_url` derived from the
  product URL host. Resolved ids are cached in memory.
- Each product is upserted into `products` keyed by **`(store_id, url)`** so a
  re-scrape updates the existing row (price/shipping/affiliate/title/image/
  scraped_at) rather than creating duplicates.
- The required `UNIQUE (store_id, url)` index is declared in the canonical schema
  (`infra/db/init/01-schema.sql`) and also created idempotently at startup
  (`EnsureSchema`) so the service is self-sufficient against an existing DB.

## Affiliate Link Injection (FR-6)

Driven by a JSON config (point `NORMALIZATION_AFFILIATE_CONFIG` at it; see
[`configs/affiliates.example.json`](configs/affiliates.example.json)):

```json
{
  "default_params": { "utm_source": "mergemarket" },
  "stores": {
    "jumia-cm": { "params": { "aff": "mergemarket-21" } },
    "dummyjson": { "template": "https://go.partner.com/redirect?url={url}" }
  }
}
```

- `template` — deep-link wrapper; `{url}` (URL-encoded) / `{url_raw}` placeholders.
- `params` — query parameters appended to the link.
- A configured store fully specifies its own behaviour; `default_params` applies
  only to stores **without** an entry. No config = links pass through unchanged.

## Configuration (environment)

| Var | Default | Meaning |
|---|---|---|
| `NORMALIZATION_PORT` | `8084` | HTTP port for `/health` + `/stats`. |
| `NORMALIZATION_WORKERS` | `5` | Concurrent worker goroutines. |
| `SCRAPER_NORMALIZE_QUEUE` | `normalize_queue` | Redis list of incoming raw results (shared with A-05). |
| `NORMALIZATION_QUEUE_POLL_TIMEOUT` | `5s` | Blocking dequeue bound (shutdown responsiveness). |
| `NORMALIZATION_AFFILIATE_CONFIG` | _(empty)_ | Path to the affiliate JSON config (empty = pass-through). |
| `DATABASE_URL` | _(assembled)_ | Postgres DSN; if unset, built from `DB_*`. |
| `DB_HOST`/`DB_PORT`/`DB_NAME`/`DB_USER`/`DB_PASSWORD` | `localhost`/`5432`/`mergemarket`/`postgres`/`` | DB connection. |
| `REDIS_HOST`/`REDIS_PORT`/`REDIS_PASSWORD`/`REDIS_DB` | `localhost`/`6379`/``/`0` | Redis connection. |

## Run

```bash
go test ./...
DB_TEST_DSN="postgres://postgres:pass@localhost:5432/mergemarket?sslmode=disable" \
  go test ./internal/store/...        # opt-in live DB test
go run ./cmd/normalization
```

PostgreSQL is a hard dependency (it is the sink), so the service fails fast if it
cannot connect. Redis being momentarily down at startup is tolerated — workers
retry on the next poll (NFR-2).

End-to-end smoke test (with the scraper-service running):

```bash
redis-cli RPUSH scrape_queue '{"store_id":"dummyjson","query":"phone","location":"CM"}'
# scraper publishes a RawResult to normalize_queue; this service drains it:
curl localhost:8084/stats
```
