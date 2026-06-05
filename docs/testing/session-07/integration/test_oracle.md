## Oracle — Product Detail Integration (B-04)

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| Tap a result card | mock server up | Product Detail route loads from `/history` + `/search` + `/truth-score` | USER_FLOWS Flow 2 |
| Detail loaded | offers returned | best-price headline = cheapest total; all four sections render | COMPONENT_LIBRARY |
| Add to Wishlist (prod-001) | mock server up | POST `/api/v1/wishlist` → 409 → SnackBar with the message | API_CONTRACTS §Wishlist |
| Add to Wishlist (other id) | mock server up | POST → 201 → "Added to your wishlist." SnackBar | API_CONTRACTS §Wishlist |
