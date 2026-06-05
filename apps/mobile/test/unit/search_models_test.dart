// Unit tests for the Search data models (B-03).
//
// Source of truth: project_docs/api/API_CONTRACTS.md (GET /api/v1/search).
// Test artefacts: docs/testing/session-06/unit/.

import 'package:flutter_test/flutter_test.dart';
import 'package:mergemarket/models/product.dart';

import '../mocks/search_mock_data.dart';

void main() {
  group('B-03 Search model unit tests', () {
    test('TC-06-U-001: SearchResponse.fromJson decodes all fields', () {
      final response = SearchResponse.fromJson(kSearchSuccessJson);

      expect(response.query, 'galaxy');
      expect(response.results, hasLength(3));
      expect(response.cached, isTrue);
      expect(response.latencyMs, 42);
    });

    test('TC-06-U-002: total_cost equals price + shipping for every offer', () {
      final response = SearchResponse.fromJson(kSearchSuccessJson);

      for (final offer in response.results) {
        expect(offer.totalCost, offer.price + offer.shipping,
            reason: '${offer.store} total_cost must equal price + shipping');
      }
    });

    test('TC-06-U-003: storeCount counts distinct stores', () {
      final response = SearchResponse.fromJson(kSearchSuccessJson);
      expect(response.storeCount, 3);
    });

    test('TC-06-U-004: empty results decode to an empty list, not an error',
        () {
      final response = SearchResponse.fromJson(kSearchEmptyJson);
      expect(response.results, isEmpty);
      expect(response.cached, isFalse);
    });

    test('TC-06-U-005: Product.fromJson tolerates missing fields', () {
      final product = Product.fromJson(const {'product_id': 'p1'});
      expect(product.productId, 'p1');
      expect(product.title, '');
      expect(product.price, 0);
      expect(product.dealScore, 0);
    });
  });
}
