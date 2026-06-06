// Unit tests for the AuthSession model (B-08).
//
// Test artefacts: docs/testing/session-11/unit/.

import 'package:flutter_test/flutter_test.dart';
import 'package:mergemarket/models/auth_session.dart';

void main() {
  group('B-08 AuthSession.fromJson', () {
    test('TC-11-U-001: decodes the token bundle from the contract shape', () {
      final session = AuthSession.fromJson({
        'token': 'access',
        'refresh_token': 'refresh',
        'expires_at': '2999-01-01T00:00:00Z',
      });
      expect(session.token, 'access');
      expect(session.refreshToken, 'refresh');
      expect(session.isExpired, isFalse);
    });

    test('TC-11-U-002: a past expiry is reported as expired', () {
      final session = AuthSession.fromJson({
        'token': 'access',
        'refresh_token': 'refresh',
        'expires_at': '2000-01-01T00:00:00Z',
      });
      expect(session.isExpired, isTrue);
    });

    test('TC-11-U-003: missing/invalid fields fall back to safe (expired) values',
        () {
      final session = AuthSession.fromJson(const {});
      expect(session.token, '');
      expect(session.refreshToken, '');
      expect(session.isExpired, isTrue); // epoch fallback
    });
  });
}
