## Test Plan — Unit — Session 08 — B-05 Flutter Wishlist Screen

**Scope:** The Wishlist feature delivered by B-05:
- Data models: `Wishlist` / `WishlistItem` / `WishlistStore` `fromJson`, the
  derived `storeCount` and `bestTotalCost`, and safe fallbacks.
- `WishlistRepository.list` / `.add` / `.remove` — verbs, URLs, request body,
  status-code → `ApiException` mapping (200 / 201 / 409 / 204 / 404 / 5xx),
  transport-error handling.
- Widget: `MMWishlistBoard` (renders a tile per item with store count; bell →
  `onSetAlert(productId)`; tap → `onTap(productId)`; swipe → `onRemove(wishlistId)`).
- Wishlist screen: success board, empty prompt, error with retry, driven through
  the real `wishlistProvider` chain.

**Out of scope:** Live backend (mocked via `MockClient`); the Set-Alert sheet
(B-06); navigation targets (asserted at integration); pixel fidelity.

**Approach:** Pure unit tests for models + repository. Widget tests for the
board (callbacks verified directly) and the screen (real provider chain with
`httpClientProvider` overridden by a `MockClient`).

**Entry criteria:** `flutter pub get` succeeds; `flutter analyze` clean; B-01/B-02
complete.

**Exit criteria:** All unit cases (TC-08-U-001 … 017) pass; analyzer clean.

**Tools:** `flutter_test`, `package:http/testing.dart` (`MockClient`),
`flutter_riverpod` `ProviderScope` overrides.

**Assumptions:** Fixtures mirror the B-02 `Wishlist` body; the stateless mock
means add/remove don't change later GETs (tests assert on actions, not on
post-mutation list contents).

**Risk:** The grid tiles host a `CachedNetworkImage`/`MMSkeletonLoader` shimmer
that animates forever, so widget tests use bounded `pump(Duration)` for the
swipe-dismiss animation rather than `pumpAndSettle` on loading content.
