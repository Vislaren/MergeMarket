# Test Cases — Unit — Session 08 — B-05 Wishlist

_(U = Unit. All executed via `flutter test`; all PASS.)_

| ID | Task | Preconditions | Input | Expected | Actual |
|----|------|---------------|-------|----------|--------|
| TC-08-U-001 | B-05 | — | `kWishlistJson` | decodes 2 items; first has 2 stores | [PASS] |
| TC-08-U-002 | B-05 | — | first item | `storeCount`=2, `bestTotalCost`=247900 | [PASS] |
| TC-08-U-003 | B-05 | — | `kWishlistEmptyJson` | empty items list | [PASS] |
| TC-08-U-004 | B-05 | — | item w/ no stores | empty stores, count 0, bestTotalCost null | [PASS] |
| TC-08-U-005 | B-05 | MockClient 200 | list() | decodes 2 items | [PASS] |
| TC-08-U-006 | B-05 | MockClient throws | list() | `ApiException(network)` | [PASS] |
| TC-08-U-007 | B-05 | MockClient 201 | add('prod-014') | POST `/api/v1/wishlist`, body `{product_id}` | [PASS] |
| TC-08-U-008 | B-05 | MockClient 409 | add('prod-001') | `ApiException(badRequest)` msg "already" | [PASS] |
| TC-08-U-009 | B-05 | MockClient 204 | remove('wl-002') | DELETE `/api/v1/wishlist/wl-002` | [PASS] |
| TC-08-U-010 | B-05 | MockClient 404 | remove('unknown') | `ApiException(notFound)` | [PASS] |
| TC-08-U-011 | B-05 | — | board w/ 2 items | tile per item + "2 stores tracking"/"1 store tracking" | [PASS] |
| TC-08-U-012 | B-05 | — | tap bell (item 1) | `onSetAlert('prod-001')` | [PASS] |
| TC-08-U-013 | B-05 | — | tap tile (item 2) | `onTap('prod-014')` | [PASS] |
| TC-08-U-014 | B-05 | — | swipe item 1 left | `onRemove('wl-001')` | [PASS] |
| TC-08-U-015 | B-05 | MockClient 200 | open Wishlist | `MMWishlistBoard` with items | [PASS] |
| TC-08-U-016 | B-05 | MockClient 200 empty | open Wishlist | "Your wishlist is empty" + "Start Searching" | [PASS] |
| TC-08-U-017 | B-05 | MockClient throws | open Wishlist | `MMErrorState` + "Try Again" | [PASS] |
