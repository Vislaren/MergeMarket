## Oracle — B-11 End-to-End (Real Backend)

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| `POST /auth/login` valid | real auth (A-08) up | 200 `{token,refresh_token,expires_at}` | API_CONTRACTS · Auth |
| `GET /wishlist` no token | Kong JWT plugin | 401 Unauthorized | A-09 Kong config |
| `GET /wishlist` with Bearer | wishlist service up | 200 `{items:[...]}` | API_CONTRACTS · Wishlist |
| `GET /search?q=phone` | cache hit | `cached: true`, fast response | API_CONTRACTS · Search |
| `GET /search?q=phone` | cache miss, scrape | results within NFR-1 budget | NFR-1 |
| expired access token | refresh valid | transparent refresh + replay, request succeeds | refresh-on-401 (B-11) |
| expired access token | refresh expired | 401 surfaces, app routes to login | API_CONTRACTS · refresh 401 |
| price ≤ threshold (down-cross) | followed product | one alert event → notification deep-links to product | A-07 history · USER_FLOWS Flow 6 |
| `GET /products/{id}/detail` w/ Bearer | BFF + upstream up | aggregate 200; JWT seen by upstream | A-09 · B-09 · B-11 |

**Availability note (source of truth: `CONTRACT_AUDIT.md`):** only the auth and
price-history rows have a real backend today; search, wishlist, alerts, savings
and truth-score rows are blocked until Agent A builds those services.
