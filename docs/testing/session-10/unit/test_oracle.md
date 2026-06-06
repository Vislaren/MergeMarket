## Oracle - Savings Dashboard

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| `GET /api/v1/savings` | 200 with `total_saved`, `currency`, `transactions` | `SavingsSummary` decodes all fields | API_CONTRACTS.md - Savings Dashboard |
| `GET /api/v1/savings` | Transport failure | `ApiException(network)` | Agent B repository error conventions |
| `GET /api/v1/savings` | Non-200 with `{error,message}` | `ApiException(server)` using `message` | API_CONTRACTS.md - Error Shape |
| `total_saved = 33500` | Level size 50000 | Level 1, progress 0.67, 16500 to next level | B-07 gamification rule |
| `total_saved >= 500000` | Top level reached | Level 10, progress 1, no remaining amount | B-07 gamification rule |
| Empty savings body | total 0 and no transactions | Empty-state prompt and Wishlist CTA | UI_MIGRATION_PROMPT.md - four states |
| Screen load success | three transactions | `MMSavingsCard` and recent savings list render | USER_FLOWS.md - Flow 7 |
| Share button tap | savings loaded | User receives share feedback | COMPONENT_LIBRARY.md - MMSavingsCard share button |
