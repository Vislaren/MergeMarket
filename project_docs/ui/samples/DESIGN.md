---
name: MergeMarket Logic
colors:
  surface: '#f7f9fc'
  surface-dim: '#d8dadd'
  surface-bright: '#f7f9fc'
  surface-container-lowest: '#ffffff'
  surface-container-low: '#f2f4f7'
  surface-container: '#eceef1'
  surface-container-high: '#e6e8eb'
  surface-container-highest: '#e0e3e6'
  on-surface: '#191c1e'
  on-surface-variant: '#44474e'
  inverse-surface: '#2d3133'
  inverse-on-surface: '#eff1f4'
  outline: '#75777e'
  outline-variant: '#c5c6cf'
  surface-tint: '#4e5e80'
  primary: '#031634'
  on-primary: '#ffffff'
  primary-container: '#1a2b4a'
  on-primary-container: '#8293b7'
  inverse-primary: '#b6c6ee'
  secondary: '#b02d21'
  on-secondary: '#ffffff'
  secondary-container: '#fc6451'
  on-secondary-container: '#650001'
  tertiary: '#705d00'
  on-tertiary: '#ffffff'
  tertiary-container: '#c9a900'
  on-tertiary-container: '#4c3f00'
  error: '#ba1a1a'
  on-error: '#ffffff'
  error-container: '#ffdad6'
  on-error-container: '#93000a'
  primary-fixed: '#d8e2ff'
  primary-fixed-dim: '#b6c6ee'
  on-primary-fixed: '#081b39'
  on-primary-fixed-variant: '#364767'
  secondary-fixed: '#ffdad5'
  secondary-fixed-dim: '#ffb4a9'
  on-secondary-fixed: '#410000'
  on-secondary-fixed-variant: '#8e130c'
  tertiary-fixed: '#ffe16d'
  tertiary-fixed-dim: '#e9c400'
  on-tertiary-fixed: '#221b00'
  on-tertiary-fixed-variant: '#544600'
  background: '#f7f9fc'
  on-background: '#191c1e'
  surface-variant: '#e0e3e6'
typography:
  headline-lg:
    fontFamily: Inter
    fontSize: 26px
    fontWeight: '700'
    lineHeight: 32px
  headline-lg-mobile:
    fontFamily: Inter
    fontSize: 22px
    fontWeight: '700'
    lineHeight: 28px
  headline-md:
    fontFamily: Inter
    fontSize: 18px
    fontWeight: '700'
    lineHeight: 24px
  headline-sm:
    fontFamily: Inter
    fontSize: 15px
    fontWeight: '600'
    lineHeight: 20px
  body:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: '400'
    lineHeight: 20px
  body-sm:
    fontFamily: Inter
    fontSize: 12px
    fontWeight: '400'
    lineHeight: 16px
  label:
    fontFamily: Inter
    fontSize: 11px
    fontWeight: '700'
    lineHeight: 14px
    letterSpacing: 0.02em
rounded:
  sm: 0.25rem
  DEFAULT: 0.5rem
  md: 0.75rem
  lg: 1rem
  xl: 1.5rem
  full: 9999px
spacing:
  xs: 4px
  sm: 8px
  md: 16px
  lg: 24px
  xl: 32px
  xxl: 48px
---

## Brand & Style

The design system is engineered for a data-forward e-commerce environment where clarity and speed are paramount. The brand personality is authoritative yet accessible, positioning itself as a reliable tool for financial decision-making. 

The style utilizes a **Modern Corporate** aesthetic with a lean toward **Minimalism**. It prioritizes information density without sacrificing legibility. High-contrast color applications highlight price fluctuations and deal quality, while the overall UI remains grounded in a stable, neutral foundation. All elements are flat, eschewing gradients in favor of solid fills and crisp geometric alignment to maintain a technical, "real-time" feel.

## Colors

The color palette is functionally driven to categorize information at a glance. 

- **Primary (Navy):** Used for structural integrity—navigation bars, primary action buttons, and header backgrounds.
- **Accent (Red):** Reserved for urgent data points, including current prices, "Price Drop" badges, and primary Call-to-Action highlights.
- **Deal Gold:** A specialized tertiary color used exclusively for top-tier, "exceptional" value rankings.
- **Neutrals:** The background uses a cool light grey to reduce eye strain, while white surfaces provide a clean canvas for product data.
- **Semantic:** Green and Amber are strictly applied to "Value Indicators" (e.g., "Good Deal" vs. "Average Deal").

## Typography

This design system utilizes **Inter** exclusively to ensure a clean, systematic appearance across all platforms. 

The type hierarchy is optimized for scanability. Headings use a bold weight to anchor the user's eye, while body text maintains a comfortable line height for reading technical specifications. Labels are reduced in size but given extra weight and slight letter spacing to differentiate them from metadata. For mobile views, the large headline scales down to 22px to prevent awkward text wrapping in product titles.

## Layout & Spacing

The design system follows a rigid 8pt grid system (with a 4px half-step for tight components). 

- **Layout Model:** A fluid grid for mobile and tablet, transitioning to a fixed-width centered grid (12 columns) for desktop.
- **Breakpoints:** 
  - Mobile: < 600px (4 columns)
  - Tablet: 600px - 1024px (8 columns)
  - Desktop: > 1024px (12 columns)
- **Rhythm:** Use `16px (md)` for standard padding within cards and `24px (lg)` for vertical separation between sections. The `48px (xxl)` unit is reserved for top-level page margins or large hero-section padding.

## Elevation & Depth

This design system uses a **Low-Contrast Tonal Layering** strategy to create depth without visual clutter. 

- **Surface Levels:** The base background is the lowest level (`#F4F6F9`). Cards and interactive containers sit on the primary surface level (`#FFFFFF`).
- **Shadows:** Shadows are subtle and functional. Use a 10% opacity navy-tinted shadow (`#1A2B4A` at 10%) with an 8px blur and 2px Y-offset for default cards.
- **Borders:** All cards and input fields must use a 1px solid border (`#D5DCE8`) to maintain structural integrity, especially when multiple cards are stacked. Shadows should never exist without a border in this system.

## Shapes

The shape language is varied to denote the "hardness" or "permanence" of a UI element:

- **Cards/Containers:** 12px (rounded-lg equivalent) for a modern, substantial feel.
- **Form Inputs:** 8px (standard roundedness) to balance clickability with professional structure.
- **Pills/Badges/Chips:** 24px (pill-shaped) to create a distinct visual contrast against rectangular product cards, making them instantly recognizable as status indicators or categories.

## Components

- **Buttons:** Primary buttons use the Navy background with White text. Accent buttons (for "Buy Now") use Red. All buttons utilize the 8px corner radius.
- **Price Badges:** Real-time price updates are housed in high-contrast Red or Gold badges using the 24px pill shape. 
- **Cards:** Product cards must have 16px internal padding, a 12px corner radius, and a 1px border. The price should be positioned at the bottom right in Red `headline-md`.
- **Inputs:** Search bars and filters use a White surface, 1px border, and 8px corner radius. Placeholder text uses `text-secondary`.
- **Price Tracker Lists:** List items should be separated by 1px dividers (`#D5DCE8`) with 12px vertical padding. Use `body-sm` for timestamps and merchant names.
- **Deal Indicator:** A "Savings" label should always use the Green success color and `label` typography.