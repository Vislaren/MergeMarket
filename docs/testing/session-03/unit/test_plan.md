## Test Plan — Unit — Session 03 — B-01 Flutter Project Setup

**Scope:** The foundation delivered by B-01:
- Design tokens — colours, typography, spacing (`lib/theme/`).
- Global Material 3 theme `buildAppTheme()` (navy/red scheme, Inter font).
- `go_router` route table and the `routerProvider` (Riverpod).
- App boot inside `ProviderScope` and the bottom-navigation shell.

**Out of scope:** Feature behaviour of any screen (search, product detail,
wishlist, alerts, savings, auth) — those are placeholders in B-01 and are
tested under their own tasks (B-03–B-08). Network/API calls (no service layer
exists yet). Pixel-level visual fidelity to the Stitch samples.

**Approach:** Pure unit tests (`test(...)`) assert token values and theme
configuration in isolation. Widget tests (`testWidgets(...)`) pump the real
`MergeMarketApp` inside a `ProviderScope` to verify the router boots and the
shell renders — these isolate the app from any backend (no mocks needed
because B-01 has no data layer yet).

**Entry criteria:** `flutter pub get` succeeds; `flutter analyze` reports no
issues.

**Exit criteria:** All unit cases (TC-03-U-001 … 008) pass; `flutter analyze`
clean.

**Tools:** `flutter_test` (+ `mockito` available for later tasks).

**Assumptions:**
- Inter is bundled as an asset, so tests need no network and no live font
  fetch. In the headless test renderer the glyphs fall back silently; this does
  not affect the assertions (which check configuration, not rendering).
- `routerProvider` is read from a fresh `ProviderContainer` for pure-unit
  inspection.

**Risk:**
- go_router 17.x / flutter_riverpod 3.x are recent majors; API drift from
  older examples is possible. Mitigated by `flutter analyze` + executing the
  suite.
- Placeholder screens will be replaced; tests assert on stable text
  ("MergeMarket", tab labels) that should survive into the real screens.
