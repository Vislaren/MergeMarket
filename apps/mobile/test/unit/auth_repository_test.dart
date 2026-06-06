// Unit tests for AuthRepository (B-08) using package:http MockClient.
//
// Covers register (201 / 400 / 409), login (200 / 401), refresh (200 / 401),
// correct verbs + URLs + body, and the transport-error → network mapping.
//
// Test artefacts: docs/testing/session-11/unit/.

import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mergemarket/services/api_exception.dart';
import 'package:mergemarket/services/auth_repository.dart';

import '../mocks/auth_mock_data.dart';

void main() {
  group('B-08 AuthRepository.register', () {
    test('TC-11-U-009: 201 POSTs credentials and decodes the session',
        () async {
      late http.Request captured;
      final repo = AuthRepository(client: MockClient((req) async {
        captured = req;
        return http.Response(jsonEncode(authTokenJson()), 201);
      }));
      final session = await repo.register('new@example.com', 'secret123');

      expect(captured.method, 'POST');
      expect(captured.url.path, '/api/v1/auth/register');
      expect(jsonDecode(captured.body),
          {'email': 'new@example.com', 'password': 'secret123'});
      expect(session.token, 'mock.jwt.access-token');
      expect(session.isExpired, isFalse);
    });

    test('TC-11-U-010: 400 maps to ApiException(badRequest)', () async {
      final repo = AuthRepository(
          client: MockClient(
              (_) async => http.Response(jsonEncode(kInvalidInputJson), 400)));
      await expectLater(
        repo.register('', ''),
        throwsA(isA<ApiException>()
            .having((e) => e.kind, 'kind', ApiErrorKind.badRequest)),
      );
    });

    test('TC-11-U-011: 409 maps to ApiException(conflict)', () async {
      final repo = AuthRepository(
          client: MockClient(
              (_) async => http.Response(jsonEncode(kEmailExistsJson), 409)));
      await expectLater(
        repo.register('taken@mergemarket.app', 'secret123'),
        throwsA(isA<ApiException>()
            .having((e) => e.kind, 'kind', ApiErrorKind.conflict)),
      );
    });
  });

  group('B-08 AuthRepository.login', () {
    test('TC-11-U-012: 200 decodes the session', () async {
      late http.Request captured;
      final repo = AuthRepository(client: MockClient((req) async {
        captured = req;
        return http.Response(jsonEncode(authTokenJson()), 200);
      }));
      final session = await repo.login('user@example.com', 'secret123');

      expect(captured.url.path, '/api/v1/auth/login');
      expect(session.refreshToken, 'mock.jwt.refresh-token');
    });

    test('TC-11-U-013: 401 maps to ApiException(unauthorized)', () async {
      final repo = AuthRepository(
          client: MockClient((_) async =>
              http.Response(jsonEncode(kInvalidCredentialsJson), 401)));
      await expectLater(
        repo.login('user@example.com', 'wrongpassword'),
        throwsA(isA<ApiException>()
            .having((e) => e.kind, 'kind', ApiErrorKind.unauthorized)),
      );
    });

    test('TC-11-U-014: transport failure maps to ApiException(network)',
        () async {
      final repo = AuthRepository(
          client: MockClient((_) async => throw http.ClientException('x')));
      await expectLater(
        repo.login('user@example.com', 'secret123'),
        throwsA(isA<ApiException>()
            .having((e) => e.kind, 'kind', ApiErrorKind.network)),
      );
    });
  });

  group('B-08 AuthRepository.refresh', () {
    test('TC-11-U-015: 200 POSTs the refresh token and decodes a new session',
        () async {
      late http.Request captured;
      final repo = AuthRepository(client: MockClient((req) async {
        captured = req;
        return http.Response(jsonEncode(authTokenJson()), 200);
      }));
      await repo.refresh('mock.jwt.refresh-token');

      expect(captured.url.path, '/api/v1/auth/refresh');
      expect(jsonDecode(captured.body),
          {'refresh_token': 'mock.jwt.refresh-token'});
    });

    test('TC-11-U-016: 401 maps to ApiException(unauthorized)', () async {
      final repo = AuthRepository(
          client: MockClient((_) async => http.Response('{}', 401)));
      await expectLater(
        repo.refresh('expired'),
        throwsA(isA<ApiException>()
            .having((e) => e.kind, 'kind', ApiErrorKind.unauthorized)),
      );
    });
  });
}
