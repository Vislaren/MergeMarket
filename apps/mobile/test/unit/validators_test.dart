// Unit tests for the auth form validators (B-08).
//
// Test artefacts: docs/testing/session-11/unit/.

import 'package:flutter_test/flutter_test.dart';
import 'package:mergemarket/utils/validators.dart';

void main() {
  group('B-08 validateEmail', () {
    test('TC-11-U-004: empty and malformed emails are rejected', () {
      expect(validateEmail(''), isNotNull);
      expect(validateEmail('   '), isNotNull);
      expect(validateEmail('not-an-email'), isNotNull);
      expect(validateEmail('missing@tld'), isNotNull);
    });

    test('TC-11-U-005: a well-formed email passes (and is trimmed)', () {
      expect(validateEmail('user@example.com'), isNull);
      expect(validateEmail('  user@example.com  '), isNull);
    });
  });

  group('B-08 password validators', () {
    test('TC-11-U-006: login password only requires presence', () {
      expect(validateLoginPassword(''), isNotNull);
      expect(validateLoginPassword('x'), isNull);
    });

    test('TC-11-U-007: new password enforces the minimum length', () {
      expect(validateNewPassword('short'), isNotNull);
      expect(validateNewPassword('12345678'), isNull);
      expect(validateNewPassword(''), isNotNull);
    });

    test('TC-11-U-008: confirm-password must match the original', () {
      expect(validateConfirmPassword('', 'secret123'), isNotNull);
      expect(validateConfirmPassword('different', 'secret123'), isNotNull);
      expect(validateConfirmPassword('secret123', 'secret123'), isNull);
    });
  });
}
