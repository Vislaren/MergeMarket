// Shared fixtures + a fake SessionStore for the B-08 auth tests.
//
// JSON shapes mirror project_docs/api/API_CONTRACTS.md (auth) and the B-02
// mock server's token bundle.

import 'package:mergemarket/models/auth_session.dart';
import 'package:mergemarket/services/session_store.dart';

/// A token bundle whose access token expires [hours] from now (default +1h),
/// matching the mock server's `fixtures.Token()`.
Map<String, dynamic> authTokenJson({int hours = 1}) => {
      'token': 'mock.jwt.access-token',
      'refresh_token': 'mock.jwt.refresh-token',
      'expires_at':
          DateTime.now().add(Duration(hours: hours)).toIso8601String(),
    };

/// An already-expired token bundle (expiry one hour in the past).
Map<String, dynamic> expiredTokenJson() => {
      'token': 'mock.jwt.access-token',
      'refresh_token': 'mock.jwt.refresh-token',
      'expires_at':
          DateTime.now().subtract(const Duration(hours: 1)).toIso8601String(),
    };

/// 401 invalid-credentials body (login sentinel: password "wrongpassword").
const Map<String, dynamic> kInvalidCredentialsJson = {
  'error': 'invalid_credentials',
  'message': 'email or password is incorrect',
};

/// 409 email-exists body (register sentinel: email "taken@mergemarket.app").
const Map<String, dynamic> kEmailExistsJson = {
  'error': 'email_exists',
  'message': 'an account with this email already exists',
};

/// 400 invalid-input body (register: missing email/password).
const Map<String, dynamic> kInvalidInputJson = {
  'error': 'invalid_input',
  'message': 'email and password are required',
};

/// In-memory [SessionStore] for tests — no platform channel required.
class FakeSessionStore implements SessionStore {
  FakeSessionStore({AuthSession? initial, String? initialEmail})
      : _session = initial,
        _email = initialEmail;

  AuthSession? _session;
  String? _email;

  /// Number of times [clear] has been called — lets tests assert logout wiped
  /// storage.
  int clearCount = 0;

  @override
  Future<AuthSession?> read() async => _session;

  @override
  Future<String?> readEmail() async => _email;

  @override
  Future<void> save(AuthSession session, {String? email}) async {
    _session = session;
    if (email != null) _email = email;
  }

  @override
  Future<void> clear() async {
    clearCount++;
    _session = null;
    _email = null;
  }
}
