// Widget tests for the Product Detail widgets (B-04):
// MMDealMeter, MMStoreComparisonTable, MMTruthScore.
//
// Test artefacts: docs/testing/session-07/unit/.

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mergemarket/models/store_result.dart';
import 'package:mergemarket/widgets/mm_deal_meter.dart';
import 'package:mergemarket/widgets/mm_store_comparison_table.dart';
import 'package:mergemarket/widgets/mm_truth_score.dart';

Widget _wrap(Widget child) =>
    MaterialApp(home: Scaffold(body: SingleChildScrollView(child: child)));

void main() {
  group('B-04 MMDealMeter', () {
    testWidgets('TC-07-U-013: shows the score and band label', (tester) async {
      await tester.pumpWidget(_wrap(const MMDealMeter(
        score: 88,
        currentPrice: 247900,
        averagePrice: 256000,
        currency: 'XAF',
      )));
      expect(find.text('88/100'), findsOneWidget);
      expect(find.text('Exceptional'), findsOneWidget);
    });

    testWidgets('TC-07-U-014: comparison text reflects below-average price',
        (tester) async {
      await tester.pumpWidget(_wrap(const MMDealMeter(
        score: 70,
        currentPrice: 240000,
        averagePrice: 256000,
        currency: 'XAF',
      )));
      // (256000-240000)/256000 ≈ 6% below.
      expect(find.textContaining('below the 6-month average'), findsOneWidget);
    });

    testWidgets('TC-07-U-015: handles missing average gracefully',
        (tester) async {
      await tester.pumpWidget(_wrap(const MMDealMeter(
        score: 50,
        currentPrice: 0,
        averagePrice: 0,
        currency: 'XAF',
      )));
      expect(find.textContaining('Not enough price history'), findsOneWidget);
    });
  });

  group('B-04 MMStoreComparisonTable', () {
    final stores = const [
      StoreResult(store: 'AfricShop', price: 252000, shipping: 0, totalCost: 252000),
      StoreResult(store: 'Jumia', price: 245000, shipping: 5000, totalCost: 250000),
      StoreResult(store: 'Kilimall', price: 239900, shipping: 8000, totalCost: 247900),
    ];

    testWidgets('TC-07-U-016: rows sort by total cost ascending', (tester) async {
      await tester.pumpWidget(_wrap(MMStoreComparisonTable(
        stores: stores,
        currency: 'XAF',
        onGoToStore: (_) => () {},
      )));
      // Cheapest (Kilimall, 247,900) is the best deal and the first row.
      expect(find.text('Best deal'), findsOneWidget);
      final kilimall = tester.getTopLeft(find.text('Kilimall'));
      final africshop = tester.getTopLeft(find.text('AfricShop'));
      expect(kilimall.dy, lessThan(africshop.dy));
    });

    testWidgets('TC-07-U-017: Go to Store invokes the callback for its row',
        (tester) async {
      String? tapped;
      await tester.pumpWidget(_wrap(MMStoreComparisonTable(
        stores: stores,
        currency: 'XAF',
        onGoToStore: (s) => () => tapped = s.store,
      )));
      await tester.tap(find.text('Go to Store').first);
      expect(tapped, 'Kilimall'); // first (best) row
    });
  });

  group('B-04 MMTruthScore', () {
    testWidgets('TC-07-U-018: renders score, sentiment, and risk', (tester) async {
      await tester.pumpWidget(_wrap(const MMTruthScore(
        score: 82,
        sentiment: 'positive',
        fakeReviewRisk: 'low',
        summary: 'Reviews are consistent across stores.',
      )));
      expect(find.text('82'), findsOneWidget);
      expect(find.text('Positive sentiment'), findsOneWidget);
      expect(find.text('Low fake-review risk'), findsOneWidget);
    });

    testWidgets('TC-07-U-019: summary expands on Read more', (tester) async {
      await tester.pumpWidget(_wrap(const MMTruthScore(
        score: 60,
        sentiment: 'mixed',
        fakeReviewRisk: 'medium',
        summary: 'A long summary that is collapsed to two lines by default.',
      )));
      expect(find.text('Read more'), findsOneWidget);
      await tester.tap(find.text('Read more'));
      await tester.pump();
      expect(find.text('Show less'), findsOneWidget);
    });
  });
}
