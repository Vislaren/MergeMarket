// Unit tests for the Product Detail data models (B-04).
//
// Verifies PriceHistory / PricePoint / TruthScore decode the contract shapes
// and that the derived helpers (latestPrice, currency, recordedAtDate) behave.
//
// Test artefacts: docs/testing/session-07/unit/.

import 'package:flutter_test/flutter_test.dart';
import 'package:mergemarket/models/price_history.dart';
import 'package:mergemarket/models/truth_score.dart';

import '../mocks/product_detail_mock_data.dart';

void main() {
  group('B-04 PriceHistory model', () {
    test('TC-07-U-001: decodes the history contract', () {
      final h = PriceHistory.fromJson(kHistoryJson);
      expect(h.productId, 'prod-001');
      expect(h.title, 'Samsung Galaxy A54 128GB');
      expect(h.history, hasLength(6));
      expect(h.average6m, 256000);
      expect(h.lowest30d, 245000);
    });

    test('TC-07-U-002: latestPrice and currency come from the last point', () {
      final h = PriceHistory.fromJson(kHistoryJson);
      expect(h.latestPrice, 245000);
      expect(h.currency, 'XAF');
      // History is oldest-first per the contract.
      expect(h.history.first.price, 268000);
      expect(h.history.last.price, 245000);
    });

    test('TC-07-U-003: missing history decodes to an empty, safe series', () {
      final h = PriceHistory.fromJson({'product_id': 'x', 'title': 't'});
      expect(h.history, isEmpty);
      expect(h.latestPrice, isNull);
      expect(h.currency, '');
      expect(h.average6m, 0);
    });

    test('TC-07-U-004: PricePoint parses its timestamp; bad dates are safe', () {
      final p = PricePoint.fromJson(
          {'price': 100, 'currency': 'XAF', 'recorded_at': '2026-06-05T09:00:00Z'});
      expect(p.recordedAtDate.year, 2026);
      final bad = PricePoint.fromJson({'price': 1, 'recorded_at': 'not-a-date'});
      expect(bad.recordedAtDate, DateTime.fromMillisecondsSinceEpoch(0));
    });
  });

  group('B-04 TruthScore model', () {
    test('TC-07-U-005: decodes the truth-score contract', () {
      final t = TruthScore.fromJson(kTruthScoreJson);
      expect(t.score, 82);
      expect(t.sentiment, 'positive');
      expect(t.fakeReviewRisk, 'low');
      expect(t.summary, isNotEmpty);
    });

    test('TC-07-U-006: missing fields fall back to safe defaults', () {
      final t = TruthScore.fromJson({'product_id': 'x'});
      expect(t.score, 0);
      expect(t.sentiment, 'mixed');
      expect(t.fakeReviewRisk, 'medium');
      expect(t.summary, '');
    });
  });
}
