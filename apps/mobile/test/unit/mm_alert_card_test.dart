// Widget tests for MMAlertCard (B-06).
//
// Test artefacts: docs/testing/session-09/unit/.

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mergemarket/widgets/mm_alert_card.dart';

Widget _wrap({required bool isActive, VoidCallback? onDelete}) {
  return MaterialApp(
    home: Scaffold(
      body: MMAlertCard(
        alertId: 'al-001',
        productTitle: 'Samsung Galaxy A54 128GB',
        thresholdPrice: 230000,
        currency: 'XAF',
        isActive: isActive,
        onDelete: onDelete ?? () {},
      ),
    ),
  );
}

void main() {
  group('B-06 MMAlertCard', () {
    testWidgets('TC-09-U-010: shows title, threshold, and Active status',
        (tester) async {
      await tester.pumpWidget(_wrap(isActive: true));
      expect(find.text('Samsung Galaxy A54 128GB'), findsOneWidget);
      expect(find.textContaining('230,000'), findsOneWidget);
      expect(find.text('Active'), findsOneWidget);
    });

    testWidgets('TC-09-U-011: shows Inactive when not active', (tester) async {
      await tester.pumpWidget(_wrap(isActive: false));
      expect(find.text('Inactive'), findsOneWidget);
    });

    testWidgets('TC-09-U-012: swipe-to-delete triggers onDelete',
        (tester) async {
      var deleted = false;
      await tester.pumpWidget(_wrap(isActive: true, onDelete: () => deleted = true));
      await tester.drag(
          find.text('Samsung Galaxy A54 128GB'), const Offset(-500, 0));
      await tester.pumpAndSettle();
      expect(deleted, isTrue);
    });
  });
}
