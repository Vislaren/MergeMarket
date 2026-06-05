import 'package:flutter/material.dart';

/// MergeMarket typography tokens.
///
/// The type scale from `project_docs/ui/COMPONENT_LIBRARY.md`. All text uses
/// the Inter font family, applied globally via [buildTextTheme] / Google Fonts
/// in `theme.dart`. Screen code references these named styles and may
/// `.copyWith(...)` colour only — never redefine font sizes inline.
const TextStyle headingLarge =
    TextStyle(fontSize: 26, fontWeight: FontWeight.w700);
const TextStyle headingMedium =
    TextStyle(fontSize: 18, fontWeight: FontWeight.w700);
const TextStyle headingSmall =
    TextStyle(fontSize: 15, fontWeight: FontWeight.w600);
const TextStyle bodyRegular =
    TextStyle(fontSize: 14, fontWeight: FontWeight.w400);
const TextStyle bodySmall =
    TextStyle(fontSize: 12, fontWeight: FontWeight.w400);
const TextStyle labelBold = TextStyle(
  fontSize: 11,
  fontWeight: FontWeight.w700,
  letterSpacing: 0.5,
);
