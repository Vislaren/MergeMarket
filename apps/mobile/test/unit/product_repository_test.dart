// Unit tests for ProductRepository (B-04) using package:http MockClient.
//
// Covers history + truth-score 200 decode and the contract's 404 → notFound and
// transport-error → network mappings.
//
// Test artefacts: docs/testing/session-07/unit/.

import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mergemarket/services/api_exception.dart';
import 'package:mergemarket/services/product_repository.dart';

import '../mocks/product_detail_mock_data.dart';

http.Client _json(int status, Object body) => MockClient(
    (_) async => http.Response(jsonEncode(body), status,
        headers: {'content-type': 'application/json'}));

void main() {
  group('B-04 ProductRepository.history', () {
    test('TC-07-U-007: 200 decodes a PriceHistory', () async {
      final repo = ProductRepository(client: _json(200, kHistoryJson));
      final h = await repo.history('prod-001');
      expect(h.title, 'Samsung Galaxy A54 128GB');
      expect(h.history, hasLength(6));
    });

    test('TC-07-U-008: builds the correct history URL', () async {
      late Uri captured;
      final client = MockClient((req) async {
        captured = req.url;
        return http.Response(jsonEncode(kHistoryJson), 200);
      });
      await ProductRepository(client: client).history('prod-099');
      expect(captured.path, '/api/v1/products/prod-099/history');
    });

    test('TC-07-U-009: 404 maps to ApiException(notFound)', () async {
      final repo = ProductRepository(client: _json(404, kNotFoundJson));
      await expectLater(
        repo.history('unknown'),
        throwsA(isA<ApiException>()
            .having((e) => e.kind, 'kind', ApiErrorKind.notFound)),
      );
    });

    test('TC-07-U-010: transport failure maps to ApiException(network)',
        () async {
      final client = MockClient((_) async => throw http.ClientException('x'));
      final repo = ProductRepository(client: client);
      await expectLater(
        repo.history('prod-001'),
        throwsA(isA<ApiException>()
            .having((e) => e.kind, 'kind', ApiErrorKind.network)),
      );
    });
  });

  group('B-04 ProductRepository.truthScore', () {
    test('TC-07-U-011: 200 decodes a TruthScore', () async {
      final repo = ProductRepository(client: _json(200, kTruthScoreJson));
      final t = await repo.truthScore('prod-001');
      expect(t.score, 82);
      expect(t.sentiment, 'positive');
    });

    test('TC-07-U-012: builds the correct truth-score URL', () async {
      late Uri captured;
      final client = MockClient((req) async {
        captured = req.url;
        return http.Response(jsonEncode(kTruthScoreJson), 200);
      });
      await ProductRepository(client: client).truthScore('prod-007');
      expect(captured.path, '/api/v1/products/prod-007/truth-score');
    });
  });
}
