// Unit tests for SavingsRepository (B-07) using package:http MockClient.
//
// Covers 200 decoding, correct verb/path, server error mapping, and transport
// error mapping.
//
// Test artefacts: docs/testing/session-10/unit/.

import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mergemarket/services/api_exception.dart';
import 'package:mergemarket/services/savings_repository.dart';

import '../mocks/savings_mock_data.dart';

void main() {
  group('B-07 SavingsRepository.getSavings', () {
    test('TC-10-U-005: 200 decodes the savings summary', () async {
      final repo = SavingsRepository(
        client: MockClient(
          (_) async => http.Response(jsonEncode(kSavingsJson), 200),
        ),
      );

      final summary = await repo.getSavings();

      expect(summary.totalSaved, 33500);
      expect(summary.transactions, hasLength(3));
    });

    test('TC-10-U-006: GETs /api/v1/savings with JSON accept header', () async {
      late http.Request captured;
      final repo = SavingsRepository(
        client: MockClient((req) async {
          captured = req;
          return http.Response(jsonEncode(kSavingsJson), 200);
        }),
      );

      await repo.getSavings();

      expect(captured.method, 'GET');
      expect(captured.url.path, '/api/v1/savings');
      expect(captured.headers['Accept'], 'application/json');
    });

    test('TC-10-U-007: non-200 maps to ApiException(server)', () async {
      final repo = SavingsRepository(
        client: MockClient(
          (_) async => http.Response(jsonEncode(kSavingsErrorJson), 500),
        ),
      );

      await expectLater(
        repo.getSavings(),
        throwsA(
          isA<ApiException>()
              .having((e) => e.kind, 'kind', ApiErrorKind.server)
              .having((e) => e.message, 'message', contains('unavailable')),
        ),
      );
    });

    test(
      'TC-10-U-008: transport failure maps to ApiException(network)',
      () async {
        final repo = SavingsRepository(
          client: MockClient((_) async => throw http.ClientException('x')),
        );

        await expectLater(
          repo.getSavings(),
          throwsA(
            isA<ApiException>().having(
              (e) => e.kind,
              'kind',
              ApiErrorKind.network,
            ),
          ),
        );
      },
    );
  });
}
