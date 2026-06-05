## Oracle — Alerts (B-06)

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| `GET /api/v1/alerts` 200 | has alerts | `{alerts:[{alert_id,product_id,title,threshold_price,currency,is_active,created_at}]}` decodes | API_CONTRACTS §Alerts |
| `POST /api/v1/alerts` | valid body | 201 `{alert_id,created_at}` → success | API_CONTRACTS §Alerts |
| `POST /api/v1/alerts` | missing/invalid fields | 400 `invalid_input` → `ApiException(badRequest)` | API_CONTRACTS §Alerts |
| `DELETE /alerts/{id}` | valid id | 204 no body → success | API_CONTRACTS §Alerts |
| `DELETE /alerts/{id}` | id "unknown" | 404 `not_found` → `ApiException(notFound)` | API_CONTRACTS §Alerts |
| Alert card | `is_active` true | "Active" chip (green) | COMPONENT_LIBRARY §MMAlertCard |
| Alert card | `is_active` false | "Inactive" chip | COMPONENT_LIBRARY §MMAlertCard |
| Set-Alert sheet | average present | default threshold ≈ average, clamped to [min,max] | UI sample (set_alert) |
| Set Alert confirmed | sheet open | POST create + return threshold + navigate to Alerts | USER_FLOWS Flow 5 |
| Set Alert cancelled | sheet open | no POST; nothing created | USER_FLOWS Flow 5 |
| Alerts load | empty | empty-state prompt | UI_MIGRATION §Step 5 |
| Alerts load | request fails | `MMErrorState` with retry | UI_MIGRATION §Step 5 |
