import 'package:flutter/material.dart';

import 'colours.dart';
import 'typography.dart';

/// The font family bundled in `assets/fonts/` and declared in `pubspec.yaml`.
///
/// Inter is shipped as an asset (a single variable TTF) rather than fetched at
/// runtime via `google_fonts`, so the app — and its tests — never depend on
/// network access. This matches the project's $0 / offline-resilient ethos.
const String interFontFamily = 'Inter';

/// Builds the global Material 3 [ThemeData] for MergeMarket.
///
/// Wires the design tokens from `COMPONENT_LIBRARY.md` into a single theme:
/// the bundled Inter font family is applied across the whole text theme, the
/// navy/red colour scheme is seeded, and surfaces, cards, inputs, and buttons
/// follow the "Modern Corporate / Minimalism" rules in
/// `project_docs/ui/samples/DESIGN.md` (flat fills, 1px borders, 8/12px radii).
ThemeData buildAppTheme() {
  final colourScheme = ColorScheme.fromSeed(
    seedColor: primaryNavy,
    primary: primaryNavy,
    secondary: accentRed,
    surface: surfaceWhite,
    error: accentRed,
    brightness: Brightness.light,
  );

  return ThemeData(
    useMaterial3: true,
    fontFamily: interFontFamily,
    colorScheme: colourScheme,
    scaffoldBackgroundColor: backgroundLight,
    textTheme: const TextTheme().copyWith(
      headlineLarge: headingLarge.copyWith(color: textPrimary),
      headlineMedium: headingMedium.copyWith(color: textPrimary),
      titleMedium: headingSmall.copyWith(color: textPrimary),
      bodyMedium: bodyRegular.copyWith(color: textPrimary),
      bodySmall: bodySmall.copyWith(color: textSecondary),
      labelSmall: labelBold.copyWith(color: textSecondary),
    ),
    appBarTheme: const AppBarTheme(
      backgroundColor: primaryNavy,
      foregroundColor: surfaceWhite,
      elevation: 0,
      centerTitle: false,
    ),
    cardTheme: CardThemeData(
      color: surfaceWhite,
      elevation: 0,
      margin: EdgeInsets.zero,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: const BorderSide(color: borderGrey),
      ),
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: true,
      fillColor: surfaceWhite,
      hintStyle: bodyRegular.copyWith(color: textSecondary),
      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(8),
        borderSide: const BorderSide(color: borderGrey),
      ),
      enabledBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(8),
        borderSide: const BorderSide(color: borderGrey),
      ),
      focusedBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(8),
        borderSide: const BorderSide(color: primaryNavy, width: 2),
      ),
    ),
    elevatedButtonTheme: ElevatedButtonThemeData(
      style: ElevatedButton.styleFrom(
        backgroundColor: primaryNavy,
        foregroundColor: surfaceWhite,
        elevation: 0,
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 14),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
        ),
      ),
    ),
    outlinedButtonTheme: OutlinedButtonThemeData(
      style: OutlinedButton.styleFrom(
        foregroundColor: primaryNavy,
        side: const BorderSide(color: primaryNavy),
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 14),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
        ),
      ),
    ),
    dividerTheme: const DividerThemeData(color: borderGrey, thickness: 1),
    chipTheme: ChipThemeData(
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(24),
      ),
    ),
    navigationBarTheme: NavigationBarThemeData(
      backgroundColor: surfaceWhite,
      indicatorColor: primaryNavy.withValues(alpha: 0.12),
    ),
  );
}
