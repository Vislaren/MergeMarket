# Session 08 — Test Artefacts

**Task under test:** B-05 — Flutter Wishlist Screen (Agent B, client-developer role)
**Tester:** Agent B (Quality Engineer role)
**Date:** 2026-06-05

---

## What was tested

B-05 built the Wishlist flow (USER_FLOWS Flow 4) on the B-01 foundation,
pointed at the B-02 mock server, and wired Add-to-Wishlist into Product Detail:

- **Data layer** — `Wishlist` / `WishlistItem` / `WishlistStore` models decoding
  `GET /api/v1/wishlist`; `WishlistRepository` for list (200), add (201/409/400),
  and remove (204/404), mapping status codes + transport failures to a typed
  `ApiException`.
- **Business logic** — `wishlistProvider` (the list) and `WishlistActions`
  (add/remove → invalidate the list).
- **Widget** — `MMWishlistBoard` (2-column grid, image/title/store-count/best
  price, bell to set an alert, swipe-to-remove).
- **Screen** — Wishlist with the four states (loading skeleton grid, success
  board, empty prompt + "Start Searching", error with retry).
- **Cross-task wiring** — the Product Detail favourite button now calls
  `wishlistActions.add` (B-04 had a stub).

Out of scope: the Set-Alert sheet behind the bell (delivered in B-06; a SnackBar
stubbed it during B-05 and is replaced in B-06); real backend (B-11). The B-02
mock is stateless, so add/remove don't change subsequent GETs.

---

## Layout

```
docs/testing/session-08/
├── README.md                 (this file)
├── unit/{test_plan,test_cases,test_oracle}.md
└── integration/{test_plan,test_cases,test_oracle}.md
```

**Runnable test code lives in the Flutter package:**

| Suite | File |
|-------|------|
| Unit — wishlist models | `apps/mobile/test/unit/wishlist_models_test.dart` |
| Unit — repository (MockClient) | `apps/mobile/test/unit/wishlist_repository_test.dart` |
| Unit — MMWishlistBoard | `apps/mobile/test/unit/mm_wishlist_board_test.dart` |
| Unit — wishlist screen states | `apps/mobile/test/unit/wishlist_screen_test.dart` |
| Shared fixtures | `apps/mobile/test/mocks/wishlist_mock_data.dart` |
| Integration — wishlist flow | `apps/mobile/integration_test/wishlist_flow_test.dart` |

Run unit tests: `cd apps/mobile && flutter test test/unit`
Run integration: `cd apps/mobile && flutter test integration_test/wishlist_flow_test.dart --dart-define=API_BASE_URL=http://10.0.2.2:8081`

---

## Results summary

| Suite | Cases | Status |
|-------|-------|--------|
| Unit | TC-08-U-001 … 017 | **17/17 PASS** (executed; `flutter test`) |
| Integration | TC-08-I-001 … 002 | **PENDING** — no device/emulator this session |

`flutter analyze`: **No issues found.**

**Quality gate:** PENDING PIPELINE RUN — no SonarQube scan has run against this
branch. Analyzer clean; all executable cases pass.

**Known gaps:** Integration cases need a device/emulator + running mock server.
Mutations are not persisted by the stateless mock.
