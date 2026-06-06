## Oracle — Wishlist (B-05)

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| `GET /api/v1/wishlist` 200 | has items | `{items:[{wishlist_id,product_id,title,image_url,stores[],added_at}]}` decodes | API_CONTRACTS §Wishlist |
| `POST /api/v1/wishlist` | new product | 201 `{wishlist_id,added_at}` → success | API_CONTRACTS §Wishlist |
| `POST /api/v1/wishlist` | product_id "prod-001" | 409 `already_in_wishlist` → `ApiException(badRequest)` | API_CONTRACTS §Wishlist |
| `DELETE /wishlist/{id}` | valid id | 204 no body → success | API_CONTRACTS §Wishlist |
| `DELETE /wishlist/{id}` | id "unknown" | 404 `not_found` → `ApiException(notFound)` | API_CONTRACTS §Wishlist |
| item `bestTotalCost` | stores present | min of `total_cost` across stores | USER_FLOWS Flow 4 |
| Wishlist load | empty items | empty-state panel + "Start Searching" | UI sample (wishlist) |
| Wishlist load | request fails | `MMErrorState` with retry | UI_MIGRATION §Step 5 |
| Swipe a tile | end-to-start | item removed (`onRemove(wishlist_id)`) | COMPONENT_LIBRARY §MMWishlistBoard |
