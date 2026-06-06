## Test Plan — Unit — Session 06 — B-03 Flutter Search Screen

**Scope:** The Search feature delivered by B-03:
- Data models: `Product.fromJson`, `SearchResponse.fromJson`, the
  `total_cost = price + shipping` invariant, and `storeCount`.
- `SearchRepository.search` — URL construction (`/api/v1/search?q=&location=`),
  status-code → `ApiException` mapping (200/400/504/5xx), and transport-error
  handling.
- Ordering: the pure `sortResults` function for the three sort modes.
- Widgets: `MMSearchBar` (submit + loading), `MMProductCard` (store/title/total
  rendering, deal badge, free-shipping label).
- Results screen: the four states (loading skeletons, success list sorted
  cheapest-first, empty, error with retry) and the no-query prompt.

**Out of scope:** Live backend behaviour (mocked via `package:http`'s
`MockClient`); Product Detail / Wishlist / Alerts / Savings / Auth screens;
the trending-search shortcut content (static, not API-driven); pixel-level
visual fidelity; the wishlist heart action (lands in B-05).

**Approach:** Pure unit tests (`test(...)`) cover models, repository, and sort
logic in isolation. Widget tests (`testWidgets(...)`) pump individual widgets
and the Results screen. The Results screen is driven through its **real
provider chain** with `httpClientProvider` overridden by a `MockClient`, so it
is tested exactly as it runs against the mock server — no provider stubbing.

**Entry criteria:** `flutter pub get` succeeds; `flutter analyze` reports no
issues; B-01 (foundation) and B-02 (mock contracts) complete.

**Exit criteria:** All unit cases (TC-06-U-001 … 025) pass; `flutter analyze`
clean.

**Tools:** `flutter_test`, `package:http/testing.dart` (`MockClient`),
`flutter_riverpod` `ProviderScope` overrides.

**Assumptions:**
- `MockClient` returns the canned fixtures in
  `test/mocks/search_mock_data.dart`, which mirror the B-02 mock server's
  `Search` fixture exactly (including `total_cost = price + shipping`).
- The `MMSkeletonLoader` shimmer and `CachedNetworkImage` placeholder animate
  indefinitely, so widget tests that render cards use bounded `pump(Duration)`
  calls rather than `pumpAndSettle()` (which would never settle).

**Risk:**
- `CachedNetworkImage` makes no real request under the test binding (host
  returns 400) and falls back to its error widget; this is harmless for the
  assertions (which check card count/text, not the image bytes).
- flutter_riverpod 3.x moved `StateProvider` to a legacy export; the sort state
  uses the modern `NotifierProvider` instead. Covered by analyzer + execution.
