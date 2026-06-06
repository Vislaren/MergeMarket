## Oracle — Session 11 (Auth, BFF, Push)

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|----------------|----------------|
| `POST /auth/register` valid | new email | 201 + token bundle | API_CONTRACTS Auth |
| `POST /auth/register` `taken@mergemarket.app` | duplicate | 409 `email_exists` → `ApiException(conflict)` | API_CONTRACTS + mock sentinel |
| `POST /auth/login` `wrongpassword` | bad creds | 401 `invalid_credentials` → `ApiException(unauthorized)` | API_CONTRACTS + mock sentinel |
| `POST /auth/refresh` `expired` | stale token | 401 `token_expired` → `ApiException(unauthorized)` | API_CONTRACTS + mock sentinel |
| persisted session, `expires_at` future | app launch | restored as authenticated | B-08 design (secure storage) |
| persisted session, `expires_at` past | app launch | treated as signed out | `AuthSession.isExpired` |
| logout | tap account → logout | secure storage cleared, signed out | B-08 design |
| `GET /bff/products/{id}/detail` | history+search+truth OK | one merged payload, offers sorted by total cost, cheapest = best_offer, its deal_score | API_CONTRACTS + B-04 aggregation |
| `GET /bff/products/missing/detail` | upstream history 404 | 404 `{error:not_found}` | API_CONTRACTS Products |
| `GET /bff/api/v1/alerts` | no BFF handler | reverse-proxied to upstream unchanged | B-09 spec ("forwarding") |
| FCM data `type=price_drop, product_id=X` | foreground | in-app Alerts banner with body + View | USER_FLOWS Flow 6 |
| FCM data `type=price_drop` | notification tap | deep link to `/product/X` | USER_FLOWS Flow 6 |
| FCM data unknown type / no product id | any | dropped, never routed | B-10 design (`isRoutable`) |
