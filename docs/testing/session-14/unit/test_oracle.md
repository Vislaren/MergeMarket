## Oracle — Search service (A-14)

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| `GET /api/v1/search?q=phone&location=CM` | valid | 200 `{query, results[], cached, latency_ms}` | API_CONTRACTS.md → Search |
| result item | any | has `product_id,title,price,currency,shipping,total_cost,image_url,store,affiliate_url,deal_score,scraped_at` | API_CONTRACTS.md |
| `q` or `location` missing | invalid | 400 `{error:"missing_query"}` | API_CONTRACTS.md (400) |
| same `q`+`location` twice | cache fresh | 2nd response `cached:true` | DATABASE_SCHEMA §3 `search:{query_hash}` + ARCH §10 |
| cached entry older than stale threshold | stale hit | served immediately + background refresh | ARCHITECTURE §10 (SWR) |
| result set | scoring | cheapest total → `deal_score` 100, dearest → 0 | A-14 Deal Meter (heuristic placeholder) |
| `SEARCH_PORT` unset | startup | listens on 8087 | PORTS_README / .env.example |
| Postgres down | startup | service fails fast (hard dependency) | ARCHITECTURE §2 |

## Oracle — User-data service (A-16..A-18)

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| any `/api/v1/*` without Bearer | unauth | 401 `{error:"unauthorized"}` | API_CONTRACTS.md (Authorization) |
| expired-but-valid JWT | unauth | 401 `{error:"token_expired"}` | Auth contract |
| `POST /api/v1/wishlist {product_id}` | new | 201 `{wishlist_id, added_at}` | API_CONTRACTS.md → Wishlist |
| `POST /api/v1/wishlist` duplicate | conflict | 409 `{error:"already_in_wishlist"}` | API_CONTRACTS.md (409) |
| `DELETE /api/v1/wishlist/{id}` other user's row | not owned | 404 `{error:"not_found"}` | per-user scoping (DATABASE_SCHEMA wishlist_items.user_id) |
| `POST /api/v1/alerts {threshold_price<=0}` | invalid | 400 `{error:"invalid_input"}` | API_CONTRACTS.md (400) |
| `GET /api/v1/savings` | valid | 200 `{total_saved, currency, transactions[]}` | API_CONTRACTS.md → Savings |
| `savings` storage | schema | reads the `purchases` table | DATABASE_SCHEMA.md (new this session) |
| JWT signed with wrong secret | tampered | rejected (401) | shares auth `JWT_SECRET` (A-08) |
| `JWT_SECRET` unset | startup | service refuses to start | NFR-4 / A-08 contract |
