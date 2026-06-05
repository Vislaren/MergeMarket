// Unit tests for SearchRepository (B-03) using package:http's MockClient.
//
// Source of truth: project_docs/api/API_CONTRACTS.md (status-code contract).
// Test artefacts: docs/testing/session-06/unit/.

import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mergemarket/config/app_config.dart';
import 'package:mergemarket/services/api_exception.dart';
import 'package:mergemarket/services/search_repository.dart';

import '../mocks/search_mock_data.dart';

const _config = AppConfig(apiBaseUrl: 'http://test.local', defaultLocation: 'CM');

SearchRepository repoReturning(
  int status,
  Object body, {
  void Function(http.Request request)? onRequest,
}) {
  final client = MockClient((request) async {
    onRequest?.call(request);
    return http.Response(jsonEncode(body), status,
        headers: {'content-type': 'application/json'});
  });
  return SearchRepository(client: client, config: _config);
}

void main() {
  group('B-03 SearchRepository unit tests', () {
    test('TC-06-U-006: 200 decodes into a SearchResponse', () async {
      final repo = repoReturning(200, kSearchSuccessJson);
      final response = await repo.search('galaxy');
      expect(response.results, hasLength(3));
      expect(response.query, 'galaxy');
    });

    test('TC-06-U-007: request targets /api/v1/search with q and location',
        () async {
      late Uri captured;
      final repo = repoReturning(200, kSearchSuccessJson,
          onRequest: (r) => captured = r.url);
      await repo.search('galaxy', location: 'NG');

      expect(captured.path, '/api/v1/search');
      expect(captured.queryParameters['q'], 'galaxy');
      expect(captured.queryParameters['location'], 'NG');
    });

    test('TC-06-U-008: 400 throws badRequest with the contract message',
        () async {
      final repo = repoReturning(400, kMissingQueryJson);
      expect(
        () => repo.search(''),
        throwsA(isA<ApiException>()
            .having((e) => e.kind, 'kind', ApiErrorKind.badRequest)
            .having((e) => e.message, 'message', contains('required'))),
      );
    });

    test('TC-06-U-009: 504 throws a timeout ApiException', () async {
      final repo = repoReturning(504, kTimeoutJson);
      expect(
        () => repo.search('timeout'),
        throwsA(isA<ApiException>()
            .having((e) => e.kind, 'kind', ApiErrorKind.timeout)),
      );
    });

    test('TC-06-U-010: transport failure maps to a network ApiException',
        () async {
      final client = MockClient((_) async => throw http.ClientException('boom'));
      final repo = SearchRepository(client: client, config: _config);
      expect(
        () => repo.search('galaxy'),
        throwsA(isA<ApiException>()
            .having((e) => e.kind, 'kind', ApiErrorKind.network)),
      );
    });

    test('TC-06-U-011: unexpected 500 maps to a server ApiException', () async {
      final repo = repoReturning(500, {'error': 'server_error', 'message': 'x'});
      expect(
        () => repo.search('galaxy'),
        throwsA(isA<ApiException>()
            .having((e) => e.kind, 'kind', ApiErrorKind.server)),
      );
    });
  });
}
