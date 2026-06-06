## Oracle — B-01 Flutter Project Setup (Integration)

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|----------------|----------------|
| Tap "Wishlist" tab | app on Home | Wishlist screen shown (AppBar "Wishlist") | COMPONENT_LIBRARY §Navigation (bottom nav → /wishlist) |
| Tap "Alerts" tab | app on Wishlist | Alerts screen shown (AppBar "Alerts") | COMPONENT_LIBRARY §Navigation (bottom nav → /alerts) |
| `go('/product/abc123')` | from any screen | Product Detail shown; id `abc123` rendered | COMPONENT_LIBRARY §Navigation (`/product/:id`) + USER_FLOWS Flow 6 |
| Standalone routes | navigated to | Results / Product Detail / Login / Register render **without** bottom nav | COMPONENT_LIBRARY §Navigation ("Hidden on: Login, Register, Results, Product Detail") |
