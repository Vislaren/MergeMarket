# Scraper Service (Go) — A-05

Config-driven worker-queue scraper engine. Each store is defined by a
`StoreConfig` JSON/YAML file in `configs/`. Workers pull jobs off a Redis queue,
load the relevant store config, execute the scrape (optionally through a
validated proxy), and publish the raw results to the normalization queue for the
normalization service (A-06). A per-store **Circuit Breaker** trips after
consecutive `403`/`429` responses so a store that is actively blocking us is
left alone until it cools down.

- **Port:** `8083` (`SCRAPER_PORT`)
- **Endpoints:** `GET /health` (API contract shape), `GET /stats` (worker-pool counters)
- **Module:** `github.com/Vislaren/MergeMarket/services/scraper-service`

## Architecture

```
Redis list (scrape_queue)
        │  BLPOP
        ▼
   worker pool ──► circuit breaker (per store_id)
        │              │ open? skip
        ▼              ▼
   StoreConfig ──► scrape engine ──► HTTP (via proxy_pool)
        │                              403/429 ─► breaker.Failure()
        ▼
   RawResult ──► RPUSH normalize_queue  (consumed by A-06)
```

| Package | Responsibility |
|---|---|
| `internal/config` | Env-driven service configuration (`Load`). |
| `internal/storeconfig` | `StoreConfig` schema + JSON/YAML directory loader (`LoadDir`) → `Registry`. |
| `internal/circuitbreaker` | Per-store breaker (closed → open → half-open) on consecutive 403/429. |
| `internal/queue` | `Job`/`RawResult` types, `Queue`/`Sink` interfaces, in-memory + Redis impls. |
| `internal/proxypool` | Reads a random proxy from the `proxy_pool` Redis Set (written by A-04). |
| `internal/scraper` | Renders the store search URL, performs the request, extracts products from JSON. |
| `internal/worker` | Worker pool orchestration: dequeue → breaker → scrape → publish, with stats. |
| `internal/server` | `GET /health` + `GET /stats`, graceful shutdown. |

## Store configs

One file per store in `SCRAPER_CONFIG_DIR` (default `configs/`), `*.json`,
`*.yaml`, or `*.yml`. `{query}` and `{location}` in `search.url_template` are
URL-escaped at request time. Field paths are dotted paths into the JSON response
(numeric segments index arrays, e.g. `data.items.0.price`). See
[`configs/example-jumia.json`](configs/example-jumia.json) and
[`configs/example-dummyjson.yaml`](configs/example-dummyjson.yaml).

Currently the engine implements `mode: json_api` (preferred per ARCHITECTURE
§2). `mode: html` (CSS selectors) is declared in the schema but extraction is
not yet implemented (`scraper.ErrUnsupportedMode`).

## Queues (Redis lists)

| Key | Direction | Notes |
|---|---|---|
| `scrape_queue` (`SCRAPER_JOB_QUEUE`) | in | `Job{ job_id, store_id, query, location }`, consumed via `BLPOP`. |
| `normalize_queue` (`SCRAPER_NORMALIZE_QUEUE`) | out | `RawResult` JSON, pushed via `RPUSH`. |
| `proxy_pool` (`PROXY_POOL_KEY`) | read | working `ip:port` Set from the proxy-validator. |
| `circuit:{store_id}` | — | reserved per DATABASE_SCHEMA §3; breaker state is currently in-memory. |

Enqueue a job for local testing:

```bash
redis-cli RPUSH scrape_queue '{"store_id":"dummyjson","query":"phone","location":"CM"}'
redis-cli LRANGE normalize_queue 0 -1
```

## Configuration (environment)

All values come from the environment (see the repo-root `.env.example`):

| Var | Default | Meaning |
|---|---|---|
| `SCRAPER_PORT` | `8083` | HTTP port for `/health` + `/stats`. |
| `SCRAPER_CONFIG_DIR` | `configs` | Directory of StoreConfig files. |
| `SCRAPER_WORKERS` | `10` | Concurrent worker goroutines. |
| `SCRAPER_JOB_QUEUE` | `scrape_queue` | Redis list of incoming jobs. |
| `SCRAPER_NORMALIZE_QUEUE` | `normalize_queue` | Redis list of outgoing raw results. |
| `SCRAPER_QUEUE_POLL_TIMEOUT` | `5s` | Blocking dequeue bound (shutdown responsiveness). |
| `SCRAPER_SCRAPE_TIMEOUT` | `15s` | Per-store request timeout. |
| `SCRAPER_CIRCUIT_THRESHOLD` | `5` | Consecutive 403/429 before the breaker opens. |
| `SCRAPER_CIRCUIT_COOLDOWN` | `60s` | Open duration before a half-open trial. |
| `SCRAPER_USE_PROXY` | `true` | Route requests through `proxy_pool` (else direct). |
| `PROXY_POOL_KEY` | `proxy_pool` | Shared with the proxy-validator. |
| `REDIS_HOST` / `REDIS_PORT` / `REDIS_PASSWORD` / `REDIS_DB` | `localhost`/`6379`/``/`0` | Redis connection. |

## Run

```bash
go test ./...
go run ./cmd/scraper-service        # reads ./configs by default
```

The service tolerates Redis being down at startup (it warns and the workers
retry on the next poll), matching the resilience standard (NFR-2).
