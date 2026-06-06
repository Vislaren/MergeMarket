# User-data Service (A-16 · A-17 · A-18)

JWT-protected, per-user CRUD behind Kong→BFF. Consolidates three
`API_CONTRACTS.md` resource groups that share the same pattern (authenticated
user, Postgres CRUD scoped to that user) into one deployable:

- **A-16 Wishlist** — `GET/POST/DELETE /api/v1/wishlist`
- **A-17 Alerts** — `GET/POST/DELETE /api/v1/alerts`
- **A-18 Savings** — `GET /api/v1/savings`

Plus `GET /health`.

## Why one service

Kong (A-09) already routes all three paths to a single upstream (the BFF), and
all three are "act on the authenticated user's rows." Splitting them into three
near-identical Go services would triple the boilerplate (config, JWT, pgx pool,
Dockerfile) for no isolation benefit. They live here as one service; the contract
paths and shapes are unchanged.

## Authentication

Every `/api/v1/*` route requires `Authorization: Bearer <jwt>`. Kong validates
the token at the edge and the BFF forwards the header; this service **re-verifies**
the HS256 signature with the shared `JWT_SECRET` (same key the auth service, A-08,
signs with) and reads `user_id` from the claims. Failures return `401`
(`token_expired` for an expired-but-valid token, `unauthorized` otherwise). No
request can act on another user's data — every query is scoped by `user_id`.

## Endpoints & status codes

| Method | Path | Success | Errors |
|--------|------|---------|--------|
| GET | `/api/v1/wishlist` | 200 `{items:[…]}` | 401 |
| POST | `/api/v1/wishlist` | 201 `{wishlist_id, added_at}` | 400, 401, 409 `already_in_wishlist` |
| DELETE | `/api/v1/wishlist/{wishlist_id}` | 204 | 401, 404 |
| GET | `/api/v1/alerts` | 200 `{alerts:[…]}` | 401 |
| POST | `/api/v1/alerts` | 201 `{alert_id, created_at}` | 400, 401 |
| DELETE | `/api/v1/alerts/{alert_id}` | 204 | 401, 404 |
| GET | `/api/v1/savings` | 200 `{total_saved, currency, transactions:[…]}` | 401 |

Wishlist items include a per-store price comparison (`stores[]`) built from every
store selling a product with the same title, cheapest first.

## Data

Owns `wishlist_items`, `price_alerts`, and `purchases` (DATABASE_SCHEMA.md). The
`purchases` table (backing savings) is new in this session — added to the
canonical schema and `infra/db/init/01-schema.sql`, and created idempotently at
startup via `EnsureSchema` for databases that predate it.

## Configuration

Env only (`.env.example`): `USERDATA_PORT` (8090), `DB_*`/`DATABASE_URL`,
`JWT_SECRET` (**must match the auth service**), `USERDATA_JWT_ISSUER`
(`mergemarket-auth`).

## Known limitations

- No write path for `purchases` yet — the savings dashboard reads what the
  checkout/affiliate-conversion flow will eventually write. Seed manually for now.
- No `/metrics` endpoint yet — consistent with the other A-04..A-08 services.
