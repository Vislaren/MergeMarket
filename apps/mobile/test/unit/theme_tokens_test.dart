// Unit tests for B-01 design tokens and global theme.
//
// Source of truth: project_docs/ui/COMPONENT_LIBRARY.md (Design Tokens).
// Test artefacts: docs/testing/session-03/unit/.

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:mergemarket/theme/app_theme.dart';
import 'package:mergemarket/theme/colours.dart';
import 'package:mergemarket/theme/spacing.dart';
import 'package:mergemarket/theme/typography.dart';

void main() {
  group('B-01 Theme Tokens Unit Tests', () {
    test('TC-03-U-001: colour tokens match COMPONENT_LIBRARY hex values', () {
      expect(primaryNavy, const Color(0xFF1A2B4A));
      expect(accentRed, const Color(0xFFC0392B));
      expect(backgroundLight, const Color(0xFFF4F6F9));
      expect(surfaceWhite, const Color(0xFFFFFFFF));
      expect(borderGrey, const Color(0xFFD5DCE8));
      expect(successGreen, const Color(0xFF27AE60));
      expect(warningAmber, const Color(0xFFF39C12));
      expect(dealGold, const Color(0xFFFFD700));
    });

    test('TC-03-U-002: typography scale matches the documented sizes/weights',
        () {
      expect(headingLarge.fontSize, 26);
      expect(headingLarge.fontWeight, FontWeight.w700);
      expect(headingMedium.fontSize, 18);
      expect(headingMedium.fontWeight, FontWeight.w700);
      expect(headingSmall.fontSize, 15);
      expect(headingSmall.fontWeight, FontWeight.w600);
      expect(bodyRegular.fontSize, 14);
      expect(bodySmall.fontSize, 12);
      expect(labelBold.fontSize, 11);
      expect(labelBold.letterSpacing, 0.5);
    });

    test('TC-03-U-003: spacing tokens follow the 8pt grid', () {
      expect([xs, sm, md, lg, xl, xxl], [4.0, 8.0, 16.0, 24.0, 32.0, 48.0]);
    });

    test('TC-03-U-004: buildAppTheme is Material 3 with the navy/red scheme',
        () {
      final theme = buildAppTheme();
      expect(theme.useMaterial3, isTrue);
      expect(theme.scaffoldBackgroundColor, backgroundLight);
      expect(theme.colorScheme.primary, primaryNavy);
      expect(theme.colorScheme.secondary, accentRed);
      expect(theme.appBarTheme.backgroundColor, primaryNavy);
    });
  });
}
