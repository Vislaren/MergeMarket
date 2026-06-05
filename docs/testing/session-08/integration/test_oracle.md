## Oracle — Wishlist Integration (B-05)

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| Open Wishlist tab | mock server up | board loads from `GET /api/v1/wishlist` | USER_FLOWS Flow 4 |
| Board rendered | items present | each tile shows title, "{n} store(s) tracking", best price | COMPONENT_LIBRARY §MMWishlistBoard |
| Tap a bell | mock server up | Set-Alert sheet ("Set Price Alert") opens | USER_FLOWS Flow 5 |
| Swipe a tile | mock server up | DELETE `/api/v1/wishlist/{id}` → item removed + SnackBar | API_CONTRACTS §Wishlist |
