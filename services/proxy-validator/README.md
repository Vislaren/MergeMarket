# Proxy Validator Service (Go)

Continuously scrapes public proxy lists, tests each proxy against a real
endpoint, and writes the working IPs to Redis (`proxy_pool` Set, TTL 5min).
Implements the politeness protocol (adaptive random delays). The scraper
service (A-05) consumes the pool from Redis.

- **Port:** 8086 (`PROXY_VALIDATOR_PORT`)
- **Endpoints:** `GET /health`, `GET /stats`
- **Task:** A-04

## How it works

```
public proxy lists ──fetch──▶ parse/dedup ──validate (bounded, polite)──▶ working set
                                                                              │
                                                            atomic RENAME swap ▼
                                                                Redis  proxy_pool (TTL 5m)
```

1. **fetch** — downloads every source in `PROXY_SOURCES` (plaintext `ip:port`
   per line). A single failing source never aborts the others (NFR-2).
2. **validate** — routes a real request through each proxy to `PROXY_TEST_URL`
   with bounded concurrency; only proxies returning `< 400` are kept.
3. **politeness** — an adaptive random delay paces dispatch; it backs off after
   failures and relaxes after successes.
4. **write** — the working set is staged under a temp key and atomically
   `RENAME`d over `proxy_pool`, so readers never see a half-written pool.

The cycle repeats every `PROXY_REFRESH_INTERVAL` (must be `< PROXY_POOL_TTL` so
the pool never fully expires).

## Layout

```
cmd/proxy-validator/   # main: wiring, signals, graceful shutdown
internal/config/       # env-driven configuration + validation
internal/proxy/        # Addr value type + proxy-list parsing
internal/fetcher/      # downloads & parses public proxy lists
internal/validator/    # tests a single proxy against the test URL
internal/politeness/   # adaptive random-delay limiter
internal/store/        # Redis proxy_pool persistence (atomic swap)
internal/runner/       # orchestrates the scrape→validate→write cycle
internal/server/       # /health + /stats HTTP server
```

## Run locally

```bash
# from services/proxy-validator
cp ../../.env.example ../../.env   # then export the vars, or use docker-compose
go run ./cmd/proxy-validator
curl localhost:8086/health
```

## Test

```bash
go test ./...
# Live Redis store tests (skipped by default) — opt in:
REDIS_TEST_ADDR=localhost:6379 go test ./internal/store/...
```

## Configuration

All config comes from the environment (see `.env.example`, "Proxy Validator
service" block): `PROXY_SOURCES`, `PROXY_POOL_KEY`, `PROXY_POOL_TTL`,
`PROXY_TEST_URL`, `PROXY_TEST_TIMEOUT`, `PROXY_VALIDATOR_CONCURRENCY`,
`PROXY_REFRESH_INTERVAL`, `PROXY_POLITENESS_MIN`, `PROXY_POLITENESS_MAX`,
plus `REDIS_HOST`/`REDIS_PORT`/`REDIS_PASSWORD`/`REDIS_DB` and
`PROXY_VALIDATOR_PORT`.
