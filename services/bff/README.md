# BFF — Backend-for-Frontend (Go)

Shapes and aggregates data specifically for the Flutter client. **No business
logic** — it only forwards requests and merges already-computed data.

- **Port:** `8082` (env `BFF_PORT`, falls back to `PORT`)
- **Upstream:** `BFF_UPSTREAM_URL` (default `http://localhost:8089`, the B-02 mock
  server; in production this is the Kong gateway / real services)
- **Task:** B-09
- **Dependencies:** none — Go standard library only (always builds, tests run
  offline against an `httptest` upstream)

## Endpoints

| Method | Path | Behaviour |
|--------|------|-----------|
| `GET` | `/health` | `{status, service, version}` |
| `GET` | `/metrics` | Prometheus text: `bff_requests_total` counter |
| `GET` | `/api/v1/products/{product_id}/detail` | **Aggregated view** — history + truth-score + offers in one payload |
| `*`   | everything else under `/` | **Reverse-proxied** unchanged to the upstream API |

### Aggregated product detail

The one piece of shaping the BFF does. It replaces the three client round-trips
the Flutter Product Detail screen makes today (B-04) with a single call:

1. fetch the product's **history** (its title keys the offer search); a 404 here
   is the product's 404;
2. concurrently fetch the **offer search** (by title) and the **truth score** —
   both best-effort, so a failure in either degrades to an empty section rather
   than failing the whole view;
3. return history + truth-score + offers (sorted by total cost ascending), plus
   the cheapest `best_offer`, its `deal_score`, and the distinct `store_count`.

```json
GET /api/v1/products/prod-001/detail → 200
{
  "product_id": "prod-001",
  "title": "iPhone 15",
  "history": { "...": "..." },
  "truth_score": { "...": "..." },
  "offers": [ { "store": "StoreA", "total_cost": 702000, "...": "..." } ],
  "best_offer": { "store": "StoreA", "...": "..." },
  "deal_score": 88,
  "store_count": 2
}
```

## Run

```bash
# defaults: listen :8082, forward to the mock server on :8089
go run ./cmd/bff

# point at Kong / a custom port
BFF_PORT=9000 BFF_UPSTREAM_URL=http://kong:8088 go run ./cmd/bff
```

## Test

```bash
go test ./... -cover    # config 100%, server 91% (also exercises upstream)
```

## Notes

- `/metrics` is a hand-rolled single counter to honour "every Go service exposes
  `/metrics`" without adding a dependency; richer metrics arrive when the service
  is wired into Prometheus (A-12).
- Wiring the Flutter app to consume `/detail` (replacing its client-side
  aggregation) is **B-11**.
