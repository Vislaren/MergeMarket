// B-11 end-to-end integration scaffold (PENDING — needs a live backend).
//
// This drives the real journey over Kong: login → search → wishlist → alert.
// It self-skips unless API_BASE_URL points at a reachable backend, so it is safe
// to keep in the suite. To run for real once Agent A's services are deployed:
//
//   1. Bring up the stack (docker compose up -d) or target the VPS Kong.
//   2. flutter test \
//        --dart-define=API_BASE_URL=http://95.111.228.35:8088 \
//        docs/testing/session-13/integration/test_suite/b11_e2e_integration_test.dart
//
// NOTE: several legs (search, wishlist, alerts, truth-score) are additionally
// blocked until those real services exist — see ../../CONTRACT_AUDIT.md. Until
// then only the auth + history legs can pass.

import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;

const _baseUrl = String.fromEnvironment('API_BASE_URL');

void main() {
  final live = _baseUrl.isNotEmpty;

  group('B-11 E2E (real backend via Kong)', () {
    test('TC-13-I-001: login obtains a session over real auth', () async {
      if (!live) {
        markTestSkipped('API_BASE_URL unset — live backend not available');
        return;
      }
      final res = await http.post(
        Uri.parse('$_baseUrl/api/v1/auth/login'),
        headers: const {'Content-Type': 'application/json'},
        body: jsonEncode({'email': 'demo@mergemarket.app', 'password': 'secret123'}),
      );
      expect(res.statusCode, 200);
      final body = jsonDecode(res.body) as Map<String, dynamic>;
      expect(body['token'], isNotEmpty);
      expect(body['refresh_token'], isNotEmpty);
    });

    test('TC-13-I-002: a protected route without a token is rejected by Kong',
        () async {
      if (!live) {
        markTestSkipped('API_BASE_URL unset — live backend not available');
        return;
      }
      final res = await http.get(Uri.parse('$_baseUrl/api/v1/wishlist'));
      expect(res.statusCode, 401);
    });

    // TC-13-I-003…008 (search, wishlist, alert, refresh, notification, BFF
    // detail) are documented in ../test_cases.md and remain PENDING until the
    // corresponding real services exist and the stack is reachable.
  });
}
