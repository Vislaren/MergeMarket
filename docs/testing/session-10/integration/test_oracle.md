## Oracle - Savings Dashboard Integration

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| App launch | Bottom nav visible | Savings destination is available | COMPONENT_LIBRARY.md - Navigation |
| Tap Savings tab | Mock server running | `GET /api/v1/savings` drives dashboard content | API_CONTRACTS.md - Savings Dashboard |
| Savings payload with transactions | Success response | Total saved, level card, and recent savings list render | USER_FLOWS.md - Flow 7 |
| Tap transaction | Product id present | App opens `/product/{product_id}` | COMPONENT_LIBRARY.md - Navigation |
| No device/emulator | Local dev environment | Integration cases remain PENDING, not failed | Agent B testing protocol |
