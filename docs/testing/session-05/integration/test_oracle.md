## Oracle — Mock Server (B-02) — Integration

Source of truth: `project_docs/api/API_CONTRACTS.md`. Every expected status code
and field below is taken directly from the contract document. The mock server is
stateless; non-200 paths are reached via deterministic sentinel inputs.

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| `GET /health` | always | `200` `{status:"ok", service:"mock-server", version}` | API_CONTRACTS Health |
| `GET /api/v1/search?q=phone&location=CM` | valid query | `200`, results[], each `total_cost = price + shipping`, `0≤deal_score≤100`, `cached`, `latency_ms` | API_CONTRACTS Search |
| `GET /api/v1/search` | `q` missing | `400 missing_query` | API_CONTRACTS Search 400 |
| `GET /api/v1/search?q=timeout` | timeout sentinel | `504 timeout` | API_CONTRACTS Search 504 |
| `GET /api/v1/products/prod-001/history` | known product | `200`, 6 history points, `average_6m`, `lowest_30d` | API_CONTRACTS History |
| `GET /api/v1/products/unknown/history` | unknown sentinel | `404 not_found` | API_CONTRACTS History 404 |
| `POST /api/v1/auth/register` (new email) | valid | `201` `{token, refresh_token, expires_at}` | API_CONTRACTS Register 201 |
| `POST /api/v1/auth/login` (`wrongpassword`) | bad creds | `401 invalid_credentials` | API_CONTRACTS Login 401 |
| `POST /api/v1/auth/refresh` (`expired`) | bad token | `401 token_expired` | API_CONTRACTS Refresh 401 |
| `GET /api/v1/wishlist` | always | `200` `{items:[...]}` with multi-store tracking | API_CONTRACTS Wishlist |
| `POST /api/v1/wishlist` (`prod-001`) | duplicate sentinel | `409 already_in_wishlist` | API_CONTRACTS Wishlist 409 |
| `DELETE /api/v1/wishlist/wl-001` | exists | `204` no body | API_CONTRACTS Wishlist DELETE |
| `POST /api/v1/alerts` (valid) | valid | `201` `{alert_id, created_at}` | API_CONTRACTS Alerts 201 |
| `DELETE /api/v1/alerts/unknown` | unknown sentinel | `404 not_found` | API_CONTRACTS Alerts DELETE 404 |
| `GET /api/v1/savings` | always | `200`, `total_saved == Σ transactions[].saved` | API_CONTRACTS Savings |
| `OPTIONS /api/v1/*` | CORS preflight | `204` + `Access-Control-Allow-Origin: *` | B-02 dev-client requirement |
