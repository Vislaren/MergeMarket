# MergeMarket — UI Migration Prompt (Stitch → Flutter)

> Agent B reads this file before building any screen.
> Follow this process every time a Stitch design is being converted to Flutter.
> The UI sample images are in `ui/samples/`. Reference them while building.

---

## What is Stitch (Firebase)?

Stitch is Firebase's AI-powered UI prototyping tool. It produces visual
screen designs with named components, colour variables, and data bindings.
Stitch exports do not generate production Flutter code — Agent B translates
them manually using the rules in this file.

---

## Step-by-Step Migration Process

### Step 1 — Identify the Screen

Before writing any code, open the corresponding image in `ui/samples/`
and identify:

- The screen name (maps to a screen in `USER_FLOWS.md`)
- Every visible UI element on the screen
- Loading, empty, and error states if shown

### Step 2 — Map Stitch Components to Flutter Widgets

Use this translation table for every element you see in the design:

| Stitch Element | Flutter / MergeMarket Widget |
|---|---|
| Container / Frame | `Container` or `Card` |
| Text (heading) | `Text` with `headingMedium` or `headingLarge` style |
| Text (body) | `Text` with `bodyRegular` style |
| Input field | `TextField` wrapped in `MMSearchBar` if search |
| Button (primary) | `ElevatedButton` with `primaryNavy` background |
| Button (secondary) | `OutlinedButton` with `primaryNavy` border |
| Button (text only) | `TextButton` |
| List / Repeating row | `ListView.builder` |
| Grid | `GridView.builder` |
| Image placeholder | `CachedNetworkImage` with `MMSkeletonLoader` fallback |
| Badge / Chip | `Chip` or custom `Container` with `BorderRadius.circular(12)` |
| Bottom sheet | `showModalBottomSheet` |
| Navigation bar | `NavigationBar` (Material 3) |
| Progress bar | `LinearProgressIndicator` |
| Chart / Graph | `MMPriceChart` (uses fl_chart) |
| Gauge / Meter | `MMDealMeter` |
| Loading state | `MMSkeletonLoader` |
| Error state | `MMErrorState` |
| Toggle / Switch | `Switch.adaptive` |
| Swipe action | `Dismissible` |

### Step 3 — Apply Design Tokens

Never use hardcoded colour hex values or font sizes in screen code.
Always use the tokens from `COMPONENT_LIBRARY.md`:

```dart
// ✅ Correct
Text('Total Cost', style: headingSmall.copyWith(color: primaryNavy))
Container(color: backgroundLight)

// ❌ Wrong
Text('Total Cost', style: TextStyle(fontSize: 15, color: Color(0xFF1A2B4A)))
Container(color: Color(0xFFF4F6F9))
```

### Step 4 — Wire Data with Riverpod

Every screen gets a Provider. Map each Stitch data binding to a Riverpod
state:

```dart
// Stitch data binding: "searchResults" list
// Flutter equivalent:
final searchResultsProvider = FutureProvider.family<List<Product>, String>(
  (ref, query) async {
    final repo = ref.read(searchRepositoryProvider);
    return repo.search(query);
  }
);

// In the screen widget:
final results = ref.watch(searchResultsProvider(query));
return results.when(
  data: (products) => ProductList(products: products),
  loading: () => MMSkeletonLoader(width: double.infinity, height: 80),
  error: (e, _) => MMErrorState(message: e.toString(), onRetry: () => ref.invalidate(searchResultsProvider)),
);
```

### Step 5 — Handle All States

Every screen must handle four states. Map them from the Stitch design:

| State | What to Show |
|---|---|
| Loading | `MMSkeletonLoader` cards in the same layout as real content |
| Success | The actual content |
| Empty | Centred illustration + helpful message ("Search for a product above") |
| Error | `MMErrorState` with retry button |

### Step 6 — Navigation

Use `go_router` for all navigation. Map Stitch navigation arrows to routes
defined in `COMPONENT_LIBRARY.md §Navigation`:

```dart
// Stitch: arrow from Results card → Product Detail
// Flutter:
onTap: () => context.go('/product/${product.id}'),

// Stitch: back arrow
// Flutter: go_router handles this automatically via AppBar back button
```

### Step 7 — Verify Against the Sample Image

After building the screen, compare it side-by-side with the image in
`ui/samples/`. Check:

- [ ] Layout matches (column/row structure, spacing)
- [ ] Colours match the design tokens
- [ ] Typography sizes and weights match
- [ ] All states are implemented (not just the happy path)
- [ ] All interactive elements are wired up

---

## Common Stitch Patterns and Their Flutter Equivalents

### Stitch: Repeated card list with image left, text right

```dart
ListView.builder(
  itemCount: items.length,
  itemBuilder: (context, index) => MMProductCard(
    productId: items[index].id,
    title: items[index].title,
    imageUrl: items[index].imageUrl,
    price: items[index].price,
    shipping: items[index].shipping,
    totalCost: items[index].totalCost,
    store: items[index].store,
    currency: items[index].currency,
    dealScore: items[index].dealScore,
    onTap: () => context.go('/product/${items[index].id}'),
  ),
)
```

### Stitch: Header with title + subtitle + action button

```dart
Row(
  mainAxisAlignment: MainAxisAlignment.spaceBetween,
  children: [
    Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Title', style: headingMedium),
        Text('Subtitle', style: bodySmall.copyWith(color: textSecondary)),
      ],
    ),
    TextButton(onPressed: onAction, child: const Text('Action')),
  ],
)
```

### Stitch: Floating bottom bar with primary action

```dart
Scaffold(
  bottomNavigationBar: Padding(
    padding: const EdgeInsets.all(md),
    child: ElevatedButton(
      style: ElevatedButton.styleFrom(backgroundColor: primaryNavy),
      onPressed: onPrimaryAction,
      child: const Text('Primary Action'),
    ),
  ),
)
```

### Stitch: Badge / score pill

```dart
Container(
  padding: const EdgeInsets.symmetric(horizontal: sm, vertical: xs),
  decoration: BoxDecoration(
    color: accentRed,
    borderRadius: BorderRadius.circular(12),
  ),
  child: Text('85', style: labelBold.copyWith(color: surfaceWhite)),
)
```

---

## What NOT to Do

- Do not use `StatefulWidget` when Riverpod can manage the state
- Do not hardcode strings — use a constants file
- Do not skip the loading or error state — both are required
- Do not create a new widget if one exists in `COMPONENT_LIBRARY.md`
- Do not deviate from the design tokens for colours or typography
- Do not add screens or flows that are not in `USER_FLOWS.md`
