// Widget tests for MMSavingsCard (B-07).
//
// Test artefacts: docs/testing/session-10/unit/.

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mergemarket/widgets/mm_savings_card.dart';

Widget _app(Widget child) => MaterialApp(home: Scaffold(body: child));

void main() {
  group('B-07 MMSavingsCard', () {
    testWidgets('TC-10-U-009: renders total, level, progress hint, and share', (
      tester,
    ) async {
      var shared = false;

      await tester.pumpWidget(
        _app(
          MMSavingsCard(
            totalSaved: 33500,
            currency: 'XAF',
            savingsLevel: 1,
            progressToNextLevel: 0.67,
            remainingToNextLevel: 16500,
            onShare: () => shared = true,
          ),
        ),
      );

      expect(find.text('Total lifetime savings'), findsOneWidget);
      expect(find.text('XAF 33,500'), findsOneWidget);
      expect(find.text('Level 1'), findsOneWidget);
      expect(find.text('XAF 16,500 to next level'), findsOneWidget);

      await tester.tap(find.byIcon(Icons.ios_share_rounded));
      expect(shared, isTrue);
    });

    testWidgets('TC-10-U-010: top level shows completion copy', (tester) async {
      await tester.pumpWidget(
        _app(
          const MMSavingsCard(
            totalSaved: 510000,
            currency: 'XAF',
            savingsLevel: 10,
            progressToNextLevel: 1,
            remainingToNextLevel: 0,
          ),
        ),
      );

      expect(find.text('Level 10'), findsOneWidget);
      expect(find.text('Top level reached'), findsOneWidget);
    });
  });
}
