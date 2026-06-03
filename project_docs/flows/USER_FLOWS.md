# MergeMarket — User Flows

> Agent B reads this before building any screen.
> Each flow maps directly to screens in the Flutter app.
> Follow these flows exactly — do not add steps that aren't here.

---

## Flow 1 — Onboarding

```
App launch (first time)
        ↓
Splash screen (logo + loading)
        ↓
Welcome screen
    [Register]          [Log In]
        ↓                   ↓
Register screen         Login screen
(email + password)      (email + password)
        ↓                   ↓
    Validation          Validation
        ↓                   ↓
    Success             Success
        ↓                   ↓
        └────────┬──────────┘
                 ↓
        Home / Search screen
```

**Screens involved:** Splash, Welcome, Register, Login, Home

---

## Flow 2 — Real-Time Search

```
Home screen
        ↓
User types product name in search bar
        ↓
Loading state (skeleton cards, "Searching 50+ stores...")
        ↓
Results screen
    — Sorted by Total Cost (price + shipping)
    — Each card: store name, price, shipping, total, deal score badge
        ↓
User taps a result
        ↓
Product Detail screen
    — Price history chart (6 months)
    — Store comparison table (all results for this product)
    — AI Deal Meter rating
    — Product Truth Score (review sentiment)
    — Return policy summary
    — "Add to Wishlist" button
    — "Go to Store" button (affiliate link)
```

**Screens involved:** Home, Results, Product Detail

---

## Flow 3 — Share to Scrape

```
User is on Amazon / eBay app
        ↓
User taps Share on a product
        ↓
Selects MergeMarket from share sheet
        ↓
MergeMarket opens with that product URL pre-loaded
        ↓
App scrapes alternatives automatically
        ↓
Results screen (same as Flow 2)
```

**Screens involved:** Results (same component, different entry point)

---

## Flow 4 — Wishlist Management

```
Product Detail screen
        ↓
User taps "Add to Wishlist"
        ↓
Product appears in Wishlist screen
    — Visual board layout
    — Each item shows: image, title, best current price, store count
        ↓
User taps a wishlist item
        ↓
Product Detail screen (same as Flow 2)
        ↓
User can remove from wishlist via swipe or detail screen button
```

**Screens involved:** Product Detail, Wishlist

---

## Flow 5 — Price Alert Setup

```
Wishlist screen
        ↓
User taps bell icon on a wishlist item
        ↓
Set Alert bottom sheet / screen
    — Current price shown
    — Slider or input: "Alert me when price drops below ___"
    — Confirm button
        ↓
Alert saved
        ↓
Alerts screen shows the new alert as active
```

**Screens involved:** Wishlist, Set Alert, Alerts

---

## Flow 6 — Price Drop Notification

```
Background: History service detects price drop below threshold
        ↓
Push notification sent to device
    "iPhone 15 dropped to $799 on Amazon — below your $850 alert"
        ↓
User taps notification
        ↓
App opens directly to Product Detail screen for that product
        ↓
"Go to Store" button → affiliate link → store checkout
```

**Screens involved:** Product Detail (deep link from notification)

---

## Flow 7 — Savings Dashboard

```
Bottom nav → Savings tab
        ↓
Savings Dashboard screen
    — Total saved (large, prominent number)
    — Progress bar / level gamification
    — List of past savings events:
        "Saved $51 on iPhone 15 — Mar 2026"
    — Share savings button
```

**Screens involved:** Savings Dashboard

---

## Screen Inventory

| Screen | Flow(s) | Key Actions |
|---|---|---|
| Splash | 1 | Auto-navigate after load |
| Welcome | 1 | Register or Login |
| Register | 1 | Form submit → Home |
| Login | 1 | Form submit → Home |
| Home / Search | 2 | Type query → Results |
| Results | 2, 3 | Tap card → Product Detail |
| Product Detail | 2, 3, 4, 6 | Add to Wishlist, Go to Store |
| Wishlist | 4, 5 | Tap item, Set alert, Remove |
| Set Alert | 5 | Set threshold, Confirm |
| Alerts | 5 | View active alerts, Delete |
| Savings Dashboard | 7 | View savings history |
