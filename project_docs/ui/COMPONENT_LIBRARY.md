# MergeMarket — Component Library

> Agent B uses this file to build consistent UI across all screens.
> Every component here maps to a file in `apps/mobile/lib/widgets/`.
> Do not create a new widget if one already exists here — reuse it.

---

## Design Tokens

### Colours
```dart
// lib/theme/colours.dart
const Color primaryNavy    = Color(0xFF1A2B4A);
const Color accentRed      = Color(0xFFC0392B);
const Color backgroundLight = Color(0xFFF4F6F9);
const Color surfaceWhite   = Color(0xFFFFFFFF);
const Color borderGrey     = Color(0xFFD5DCE8);
const Color textPrimary    = Color(0xFF1A1A1A);
const Color textSecondary  = Color(0xFF555555);
const Color successGreen   = Color(0xFF27AE60);
const Color warningAmber   = Color(0xFFF39C12);
const Color dealGold       = Color(0xFFFFD700);
```

### Typography
```dart
// lib/theme/typography.dart
// All use system font (Inter via Google Fonts)
const headingLarge  = TextStyle(fontSize: 26, fontWeight: FontWeight.w700);
const headingMedium = TextStyle(fontSize: 18, fontWeight: FontWeight.w700);
const headingSmall  = TextStyle(fontSize: 15, fontWeight: FontWeight.w600);
const bodyRegular   = TextStyle(fontSize: 14, fontWeight: FontWeight.w400);
const bodySmall     = TextStyle(fontSize: 12, fontWeight: FontWeight.w400);
const labelBold     = TextStyle(fontSize: 11, fontWeight: FontWeight.w700,
                                letterSpacing: 0.5);
```

### Spacing
```dart
// lib/theme/spacing.dart
const double xs  = 4.0;
const double sm  = 8.0;
const double md  = 16.0;
const double lg  = 24.0;
const double xl  = 32.0;
const double xxl = 48.0;
```

---

## Components

---

### MMSearchBar
**File:** `lib/widgets/mm_search_bar.dart`
**Used on:** Home / Search screen

```dart
MMSearchBar({
  required TextEditingController controller,
  required VoidCallback onSearch,
  String hint = 'Search 50+ stores...',
  bool isLoading = false,
})
```

States: idle, focused, loading
Visual: Full-width rounded input, search icon right, navy border on focus

---

### MMProductCard
**File:** `lib/widgets/mm_product_card.dart`
**Used on:** Results screen, Wishlist screen

```dart
MMProductCard({
  required String productId,
  required String title,
  required String imageUrl,
  required double price,
  required double shipping,
  required double totalCost,
  required String store,
  required String currency,
  required int dealScore,         // 0–100
  required VoidCallback onTap,
  bool showWishlistButton = true,
})
```

States: default, loading skeleton, error (image fallback)
Visual: Card with image left, details right, deal score badge top-right,
total cost bold in accentRed, store name in textSecondary

---

### MMDealMeter
**File:** `lib/widgets/mm_deal_meter.dart`
**Used on:** Product Detail screen

```dart
MMDealMeter({
  required int score,    // 0–100
  required double currentPrice,
  required double averagePrice,
  required String currency,
})
```

Score ranges:
- 0–30: Poor deal (red)
- 31–60: Average (amber)
- 61–80: Good deal (green)
- 81–100: Exceptional (gold + animation)

Visual: Horizontal gauge bar with needle, score label, comparison text

---

### MMPriceChart
**File:** `lib/widgets/mm_price_chart.dart`
**Used on:** Product Detail screen

```dart
MMPriceChart({
  required List<PricePoint> history,   // { price, recordedAt }
  required String currency,
  int monthsBack = 6,
})
```

Uses `fl_chart` package. Navy line, light background, touch tooltip showing
exact price and date. Current price highlighted with a dot in accentRed.

---

### MMStoreComparisonTable
**File:** `lib/widgets/mm_store_comparison_table.dart`
**Used on:** Product Detail screen

```dart
MMStoreComparisonTable({
  required List<StoreResult> stores,  // { store, price, shipping, totalCost }
  required String currency,
  required VoidCallback Function(StoreResult) onGoToStore,
})
```

Sorted by total cost ascending. Best deal row highlighted with
a subtle green background. Each row has a "Go to Store" chip.

---

### MMTruthScore
**File:** `lib/widgets/mm_truth_score.dart`
**Used on:** Product Detail screen

```dart
MMTruthScore({
  required int score,                  // 0–100
  required String sentiment,           // positive | mixed | negative
  required String fakeReviewRisk,      // low | medium | high
  required String summary,
})
```

Visual: Circular score badge, sentiment label, risk chip, expandable summary

---

### MMWishlistBoard
**File:** `lib/widgets/mm_wishlist_board.dart`
**Used on:** Wishlist screen

```dart
MMWishlistBoard({
  required List<WishlistItem> items,
  required void Function(String productId) onTap,
  required void Function(String wishlistId) onRemove,
  required void Function(String productId) onSetAlert,
})
```

Visual: Grid layout (2 columns), image card, product name, best price,
bell icon for alert, swipe-to-remove gesture

---

### MMAlertCard
**File:** `lib/widgets/mm_alert_card.dart`
**Used on:** Alerts screen

```dart
MMAlertCard({
  required String alertId,
  required String productTitle,
  required double thresholdPrice,
  required String currency,
  required bool isActive,
  required VoidCallback onDelete,
})
```

Visual: List tile, threshold price in bold, active/inactive status chip,
delete swipe action

---

### MMSavingsCard
**File:** `lib/widgets/mm_savings_card.dart`
**Used on:** Savings Dashboard

```dart
MMSavingsCard({
  required double totalSaved,
  required String currency,
  required int savingsLevel,         // gamification level 1–10
  required double progressToNextLevel, // 0.0–1.0
})
```

Visual: Large savings number in successGreen, level badge, progress bar,
share button

---

### MMSkeletonLoader
**File:** `lib/widgets/mm_skeleton_loader.dart`
**Used on:** All screens during loading states

```dart
MMSkeletonLoader({
  required double width,
  required double height,
  double borderRadius = 8.0,
})
```

Animated shimmer effect using backgroundLight and borderGrey.

---

### MMErrorState
**File:** `lib/widgets/mm_error_state.dart`
**Used on:** All screens on error

```dart
MMErrorState({
  required String message,
  required VoidCallback onRetry,
  IconData icon = Icons.wifi_off_rounded,
})
```

Visual: Centred icon, message, "Try Again" button in primaryNavy

---

## Navigation Structure (go_router)

```
/                   → Home / Search
/results            → Results screen
/product/:id        → Product Detail
/wishlist           → Wishlist
/alerts             → Alerts
/savings            → Savings Dashboard
/login              → Login
/register           → Register
```

Bottom navigation bar visible on: Home, Wishlist, Alerts, Savings
Hidden on: Login, Register, Results, Product Detail
