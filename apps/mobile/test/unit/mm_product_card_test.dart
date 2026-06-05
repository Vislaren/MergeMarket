// Widget tests for MMProductCard (B-03).
//
// Source of truth: project_docs/ui/COMPONENT_LIBRARY.md → MMProductCard.
// Test artefacts: docs/testing/session-06/unit/.

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mergemarket/widgets/mm_product_card.dart';

Widget _wrap(Widget child) => MaterialApp(home: Scaffold(body: child));

void main() {
  group('B-03 MMProductCard widget tests', () {
    testWidgets('TC-06-U-016: renders store, title and formatted total cost',
        (tester) async {
      var tapped = false;
      await tester.pumpWidget(_wrap(MMProductCard(
        productId: 'prod-001',
        title: 'Samsung Galaxy A54 128GB',
        imageUrl: 'https://img.example/x.jpg',
        price: 245000,
        shipping: 5000,
        totalCost: 250000,
        store: 'Jumia',
        currency: 'XAF',
        dealScore: 88,
        onTap: () => tapped = true,
      )));

      expect(find.text('Jumia'), findsOneWidget);
      expect(find.text('Samsung Galaxy A54 128GB'), findsOneWidget);
      expect(find.text('XAF 250,000'), findsOneWidget);
      expect(find.text('Hot Deal'), findsOneWidget); // score 88 → exceptional

      await tester.tap(find.text('Samsung Galaxy A54 128GB'));
      expect(tapped, isTrue);
    });

    testWidgets('TC-06-U-017: free shipping is labelled, not shown as 0',
        (tester) async {
      await tester.pumpWidget(_wrap(MMProductCard(
        productId: 'prod-001',
        title: 'Galaxy',
        imageUrl: '',
        price: 252000,
        shipping: 0,
        totalCost: 252000,
        store: 'AfricShop',
        currency: 'XAF',
        dealScore: 74,
        onTap: () {},
      )));

      expect(find.textContaining('Free shipping'), findsOneWidget);
      expect(find.text('Good Value'), findsOneWidget); // score 74 → good
    });
  });
}
