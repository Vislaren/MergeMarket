## Test Plan - Unit - Session 10 - B-07 Flutter Savings Dashboard Screen

**Scope:** The Savings Dashboard feature delivered by B-07:
- Data models: `SavingsSummary` / `SavingsTransaction` `fromJson`, safe
  fallbacks, and derived level/progress values.
- `SavingsRepository.getSavings` - verb, URL, headers, status-code to
  `ApiException` mapping, and transport-error handling.
- Widget: `MMSavingsCard` total saved amount, level badge, next-level progress
  hint, top-level copy, and share callback.
- Savings screen: success dashboard, empty prompt, error with retry, driven
  through the real `savingsProvider` chain.

**Out of scope:** Live backend (mocked via `MockClient`); native share sheets;
real purchase history persistence.

**Approach:** Pure unit tests for models and repository. Widget tests for the
card and screen states with `httpClientProvider` overridden by a `MockClient`.

**Entry criteria:** `flutter pub get` succeeds; B-01/B-02 are complete; the mock
server contract includes `GET /api/v1/savings`.

**Exit criteria:** All unit cases (TC-10-U-001 ... 014) pass; analyzer clean.

**Tools:** `flutter_test`, `package:http/testing.dart` (`MockClient`),
`flutter_riverpod` `ProviderScope` overrides.

**Assumptions:** Fixtures mirror the B-02 `Savings` body. The savings API does
not provide product images, store names, original prices, or monthly comparison
percentages, so those visual-only sample details are not asserted.

**Risk:** Integration with Product Detail from a savings event depends on
existing product-detail aggregation over history/search/truth-score and is
covered at integration level.
