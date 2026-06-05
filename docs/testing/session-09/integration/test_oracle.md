## Oracle — Alerts Integration (B-06)

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| Open Alerts tab | mock server up | list loads from `GET /api/v1/alerts` | USER_FLOWS Flow 5 |
| List rendered | alerts present | each card shows title, threshold, Active/Inactive | COMPONENT_LIBRARY §MMAlertCard |
| Wishlist bell → Set Alert | mock server up | POST `/api/v1/alerts` → navigate to Alerts tab | USER_FLOWS Flow 5 |
| Swipe an alert card | mock server up | DELETE `/api/v1/alerts/{id}` → removed + SnackBar | API_CONTRACTS §Alerts |
