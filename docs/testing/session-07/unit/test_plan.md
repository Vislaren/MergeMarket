## Test Plan — Unit — Session 07 — B-04 Flutter Product Detail Screen

**Scope:** The Product Detail feature delivered by B-04:
- Data models: `PriceHistory` / `PricePoint` / `TruthScore` `fromJson`, the
  derived helpers (`latestPrice`, `currency`, `recordedAtDate`), and safe
  fallbacks for missing/mistyped fields.
- `ProductRepository.history` / `.truthScore` — URL construction, status-code →
  `ApiException` mapping (200/404/5xx), transport-error handling.
- Widgets: `MMDealMeter` (score band + comparison text), `MMStoreComparisonTable`
  (total-cost sort, best-deal highlight, per-row Go-to-Store callback),
  `MMTruthScore` (score/sentiment/risk render + expandable summary).
- Product Detail screen: loading skeletons, success (all sections present, best
  price = cheapest total), and 404 error with retry, driven through the real
  `productDetailProvider` chain.

**Out of scope:** Live backend behaviour (mocked via `package:http` `MockClient`);
affiliate-link opening / share (SnackBar stubs this session); `MMPriceChart`
pixel rendering (fl_chart internals — verified only that it mounts with data);
other screens.

**Approach:** Pure unit tests cover models and the repository. Widget tests pump
individual widgets and the Product Detail screen. The screen is driven through
its real provider chain with `httpClientProvider` overridden by a routing
`MockClient` that returns the right fixture per path (history / search /
truth-score) — exactly as it runs against the mock server.

**Entry criteria:** `flutter pub get` succeeds; `flutter analyze` clean;
B-01 (foundation), B-02 (mock contracts), B-03 (search data layer) complete.

**Exit criteria:** All unit cases (TC-07-U-001 … 022) pass; `flutter analyze`
clean.

**Tools:** `flutter_test`, `package:http/testing.dart` (`MockClient`),
`flutter_riverpod` `ProviderScope` overrides.

**Assumptions:**
- Fixtures in `test/mocks/product_detail_mock_data.dart` mirror the B-02 mock
  server's `History` / `TruthScore` / `Search` bodies.
- The provider fetches history first (its `title` keys the store-comparison
  search), then search + truth-score; the routing MockClient honours that order.

**Risk:**
- `CachedNetworkImage` / `MMSkeletonLoader` shimmer animates indefinitely, so
  widget tests use bounded `pump(Duration)` (or `pumpAndSettle` only on static
  error states), never `pumpAndSettle` on loading content.
- fl_chart is exercised only for "renders with data"; its drawing is not
  asserted pixel-by-pixel.
