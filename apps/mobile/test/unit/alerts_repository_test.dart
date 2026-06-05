// Unit tests for AlertsRepository (B-06) using package:http MockClient.
//
// Covers list (200), create (201 / 400), remove (204 / 404), correct verbs +
// URLs + body, and the transport-error → network mapping.
//
// Test artefacts: docs/testing/session-09/unit/.

import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mergemarket/services/alerts_repository.dart';
import 'package:mergemarket/services/api_exception.dart';

import '../mocks/alerts_mock_data.dart';

void main() {
  group('B-06 AlertsRepository.list', () {
    test('TC-09-U-004: 200 decodes the alert list', () async {
      final repo = AlertsRepository(
          client: MockClient(
              (_) async => http.Response(jsonEncode(kAlertsJson), 200)));
      final list = await repo.list();
      expect(list.alerts, hasLength(2));
    });

    test('TC-09-U-005: transport failure maps to ApiException(network)',
        () async {
      final repo = AlertsRepository(
          client: MockClient((_) async => throw http.ClientException('x')));
      await expectLater(
        repo.list(),
        throwsA(isA<ApiException>()
            .having((e) => e.kind, 'kind', ApiErrorKind.network)),
      );
    });
  });

  group('B-06 AlertsRepository.create', () {
    test('TC-09-U-006: 201 POSTs the alert payload', () async {
      late http.Request captured;
      final repo = AlertsRepository(client: MockClient((req) async {
        captured = req;
        return http.Response(jsonEncode(kAlertCreatedJson), 201);
      }));
      await repo.create(
          productId: 'prod-001', thresholdPrice: 230000, currency: 'XAF');
      expect(captured.method, 'POST');
      expect(captured.url.path, '/api/v1/alerts');
      expect(jsonDecode(captured.body), {
        'product_id': 'prod-001',
        'threshold_price': 230000,
        'currency': 'XAF',
      });
    });

    test('TC-09-U-007: 400 maps to ApiException(badRequest)', () async {
      final repo = AlertsRepository(
          client: MockClient(
              (_) async => http.Response(jsonEncode(kAlertInvalidJson), 400)));
      await expectLater(
        repo.create(productId: '', thresholdPrice: 0, currency: 'XAF'),
        throwsA(isA<ApiException>()
            .having((e) => e.kind, 'kind', ApiErrorKind.badRequest)),
      );
    });
  });

  group('B-06 AlertsRepository.remove', () {
    test('TC-09-U-008: 204 DELETEs the alert by id', () async {
      late http.Request captured;
      final repo = AlertsRepository(client: MockClient((req) async {
        captured = req;
        return http.Response('', 204);
      }));
      await repo.remove('al-002');
      expect(captured.method, 'DELETE');
      expect(captured.url.path, '/api/v1/alerts/al-002');
    });

    test('TC-09-U-009: 404 maps to ApiException(notFound)', () async {
      final repo = AlertsRepository(
          client: MockClient((_) async => http.Response('{}', 404)));
      await expectLater(
        repo.remove('unknown'),
        throwsA(isA<ApiException>()
            .having((e) => e.kind, 'kind', ApiErrorKind.notFound)),
      );
    });
  });
}
