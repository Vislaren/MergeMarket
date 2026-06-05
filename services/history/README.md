# History Service (Go) — A-07

Records price snapshots into the TimescaleDB `price_history` hypertable, runs a
scheduled **heartbeat** that re-checks followed (alerted) products and fires a
**price-drop alert** when a price crosses *down* through a threshold, and serves
the product price-history API.

- **Port:** `8085` (`HISTORY_PORT`)
- **Endpoints:**
  - `GET /api/v1/products/{product_id}/history` (API contract)
  - `GET /health` (API contract shape)
  - `GET /stats` (runner counters)
- **Module:** `github.com/Vislaren/MergeMarket/services/history`

## Jobs

```
                ┌─ Snapshot loop (HISTORY_SNAPSHOT_INTERVAL, default 24h) ─┐
                │   INSERT INTO price_history                              │
                │   SELECT id,last_price,last_shipping,currency,NOW()      │
                │   FROM products WHERE last_price IS NOT NULL             │
                └──────────────────────────────────────────────────────────┘

                ┌─ Heartbeat loop (HISTORY_HEARTBEAT_INTERVAL, default 1h) ┐
followed   ───► │ for each product with an active alert:                   │
products        │   price  = PriceSource.CurrentPrice(product)            │
(price_alerts)  │   prev   = LatestPrice(product)        (before insert)  │
                │   InsertSnapshot(product, price)                        │
                │   for each alert:                                       │
                │     if price ≤ threshold AND prev > threshold (or none) │
                │         → publish Event to Redis alert queue            │
                └──────────────────────────────────────────────────────────┘
```

| Package | Responsibility |
|---|---|
| `internal/config` | Env-driven configuration (`Load`). |
| `internal/store` | pgx `Repository`: `SnapshotAll`, `FollowedProducts`, `LatestPrice`, `InsertSnapshot`, `History`. |
| `internal/pricesource` | `Source` interface; `DBSource` (default) and best-effort `HTTPSource`. |
| `internal/alert` | `Event` type + Redis `Publisher`. |
| `internal/runner` | Snapshot + heartbeat loops, downward-crossing alert logic, stats. |
| `internal/server` | History API + `/health` + `/stats`, graceful shutdown. |

## Alert crossing logic

An alert fires only on the **downward crossing**: the current price is at or
below the threshold *and* the previous recorded price was above it (or there is
no prior observation). This fires once on the drop instead of every heartbeat
while the price stays low. Fired events are pushed to the Redis list
`price_alert_events` (`HISTORY_ALERT_QUEUE`) for the Notification Worker
(ARCHITECTURE §1) to deliver. The service does not deactivate the alert — alert
lifecycle is owned by the auth/notification side.

## Heartbeat price source (`HISTORY_HEARTBEAT_MODE`)

- `db` (default) — reads the freshest price the scraper/normalization pipeline
  already persisted (`products.last_price`). Reliable, no extra network load.
- `http` — re-fetches each followed product URL and best-effort extracts an
  embedded price (JSON-LD `"price"`, `product:price:amount`, `itemprop="price"`
  meta tags). Honours "scrape followed product URLs"; inherently fragile, so it
  is opt-in. A page with no recognizable price is skipped (NFR-2).

## Price history API

`GET /api/v1/products/{product_id}/history` returns the API_CONTRACTS.md shape:

```json
{
  "product_id": "…",
  "title": "…",
  "history": [{ "price": 10.0, "currency": "USD", "recorded_at": "…" }],
  "average_6m": 12.50,
  "lowest_30d": 9.00
}
```

`history` covers the last 6 months; `average_6m` is the 6-month mean and
`lowest_30d` the 30-day minimum (computed in SQL). Unknown product → `404
not_found`.

## Configuration (environment)

| Var | Default | Meaning |
|---|---|---|
| `HISTORY_PORT` | `8085` | HTTP port for the API + `/health` + `/stats`. |
| `HISTORY_SNAPSHOT_INTERVAL` | `24h` | How often the full price snapshot runs. |
| `HISTORY_SNAPSHOT_ON_START` | `false` | Run one snapshot immediately at startup. |
| `HISTORY_HEARTBEAT_INTERVAL` | `1h` | How often followed products are re-checked. |
| `HISTORY_HEARTBEAT_ON_START` | `true` | Run one heartbeat immediately at startup. |
| `HISTORY_HEARTBEAT_MODE` | `db` | Price source: `db` or `http`. |
| `HISTORY_HEARTBEAT_TIMEOUT` | `10s` | Per-product fetch timeout (http mode). |
| `HISTORY_ALERT_QUEUE` | `price_alert_events` | Redis list fired alerts are pushed to. |
| `DATABASE_URL` | _(assembled)_ | Postgres/Timescale DSN; if unset, built from `DB_*`. |
| `DB_HOST`/`DB_PORT`/`DB_NAME`/`DB_USER`/`DB_PASSWORD` | `localhost`/`5432`/`mergemarket`/`postgres`/`` | DB connection. |
| `REDIS_HOST`/`REDIS_PORT`/`REDIS_PASSWORD`/`REDIS_DB` | `localhost`/`6379`/``/`0` | Redis connection. |

## Run

```bash
go test ./...
DB_TEST_DSN="postgres://postgres:pass@localhost:5432/mergemarket?sslmode=disable" \
  go test ./internal/store/...        # opt-in live DB test
go run ./cmd/history
```

PostgreSQL/TimescaleDB is a hard dependency (source of followed products and sink
for snapshots), so the service fails fast if it cannot connect. Redis being
momentarily down at startup is tolerated — alert publishing retries per cycle.
