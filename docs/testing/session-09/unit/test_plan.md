## Test Plan — Unit — Session 09 — B-06 Flutter Alerts Screen

**Scope:** The Alerts feature delivered by B-06:
- Data models: `Alert` / `AlertList` `fromJson` and safe fallbacks.
- `AlertsRepository.list` / `.create` / `.remove` — verbs, URLs, request body,
  status-code → `ApiException` mapping (200 / 201 / 400 / 204 / 404 / 5xx),
  transport-error handling.
- Widgets: `MMAlertCard` (title, threshold, active/inactive chip,
  swipe-to-delete); `MMSetAlertSheet` body (product + current price + bounds +
  default suggestion); the `showSetAlertSheet` helper (confirm → POST create +
  returns the price; cancel → null + no POST).
- Alerts screen: success list + summary, empty prompt, error with retry, driven
  through the real `alertsProvider` chain.

**Out of scope:** Live backend (mocked via `MockClient`); push notification
delivery (B-10); navigation targets (asserted at integration).

**Approach:** Pure unit tests for models + repository. Widget tests for the
card, the sheet body, the full `showSetAlertSheet` flow (real provider chain
with `httpClientProvider` overridden), and the screen states.

**Entry criteria:** `flutter pub get` succeeds; `flutter analyze` clean;
B-01/B-02 and B-05 (wishlist providers, for the bell wiring) complete.

**Exit criteria:** All unit cases (TC-09-U-001 … 018) pass; analyzer clean.

**Tools:** `flutter_test`, `package:http/testing.dart` (`MockClient`),
`flutter_riverpod` `ProviderScope` overrides.

**Assumptions:** Fixtures mirror the B-02 `Alerts` body; the modal sheet needs a
tall test surface so the bottom actions are visible — the flow tests set
`tester.view.physicalSize` accordingly.

**Risk:** The sheet's product image shimmer animates forever, so the flow tests
use bounded `pump(Duration)` for the modal transition rather than
`pumpAndSettle` on loading content.
