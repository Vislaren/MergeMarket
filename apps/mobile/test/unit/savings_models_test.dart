// Unit tests for SavingsSummary and SavingsTransaction (B-07).
//
// Test artefacts: docs/testing/session-10/unit/.

import 'package:flutter_test/flutter_test.dart';
import 'package:mergemarket/models/savings.dart';

import '../mocks/savings_mock_data.dart';

void main() {
  group('B-07 Savings models', () {
    test('TC-10-U-001: decodes the savings contract', () {
      final summary = SavingsSummary.fromJson(kSavingsJson);

      expect(summary.totalSaved, 33500);
      expect(summary.currency, 'XAF');
      expect(summary.transactions, hasLength(3));
      expect(summary.transactions.first.productId, 'prod-001');
      expect(summary.transactions.first.saved, 23000);
      expect(summary.transactions.first.boughtAt.year, 2026);
    });

    test('TC-10-U-002: derives level, progress, and remaining amount', () {
      final summary = SavingsSummary.fromJson(kSavingsJson);

      expect(summary.savingsLevel, 1);
      expect(summary.progressToNextLevel, closeTo(0.67, 0.001));
      expect(summary.remainingToNextLevel, 16500);
    });

    test('TC-10-U-003: clamps level 10 progress at the top level', () {
      final summary = SavingsSummary.fromJson(kSavingsTopLevelJson);

      expect(summary.savingsLevel, 10);
      expect(summary.progressToNextLevel, 1);
      expect(summary.remainingToNextLevel, 0);
    });

    test('TC-10-U-004: malformed or missing fields fall back safely', () {
      final summary = SavingsSummary.fromJson({
        'transactions': [
          {'saved': 'not-a-number', 'bought_at': 'bad-date'},
        ],
      });

      expect(summary.totalSaved, 0);
      expect(summary.currency, '');
      expect(summary.transactions.single.productId, '');
      expect(summary.transactions.single.saved, 0);
      expect(summary.transactions.single.boughtAt.year, 1970);
    });
  });
}
