// Unit tests for search ordering logic (B-03).
//
// Test artefacts: docs/testing/session-06/unit/.

import 'package:flutter_test/flutter_test.dart';
import 'package:mergemarket/models/product.dart';
import 'package:mergemarket/providers/search_providers.dart';

import '../mocks/search_mock_data.dart';

void main() {
  group('B-03 sortResults unit tests', () {
    final products = SearchResponse.fromJson(kSearchSuccessJson).results;

    List<String> storesIn(List<Product> p) => p.map((e) => e.store).toList();

    test('TC-06-U-012: bestPrice orders by ascending total cost', () {
      final sorted = sortResults(products, ResultSort.bestPrice);
      // Kilimall 247900 < Jumia 250000 < AfricShop 252000.
      expect(storesIn(sorted), ['Kilimall', 'Jumia', 'AfricShop']);
    });

    test('TC-06-U-013: topRated orders by descending deal score', () {
      final sorted = sortResults(products, ResultSort.topRated);
      // Jumia 88 > Kilimall 81 > AfricShop 74.
      expect(storesIn(sorted), ['Jumia', 'Kilimall', 'AfricShop']);
    });

    test('TC-06-U-014: fastestShip orders by ascending shipping', () {
      final sorted = sortResults(products, ResultSort.fastestShip);
      // AfricShop 0 < Jumia 5000 < Kilimall 8000.
      expect(storesIn(sorted), ['AfricShop', 'Jumia', 'Kilimall']);
    });

    test('TC-06-U-015: sortResults does not mutate the input list', () {
      final before = storesIn(products);
      sortResults(products, ResultSort.bestPrice);
      expect(storesIn(products), before);
    });
  });
}
