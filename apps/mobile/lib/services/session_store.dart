import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../models/auth_session.dart';

/// Persists the authenticated session across app launches.
///
/// Defined as an interface so the auth controller depends on the abstraction,
/// not on `flutter_secure_storage` directly — tests inject an in-memory fake
/// (the secure-storage plugin needs a platform channel that is unavailable in
/// unit tests). The production implementation is [SecureSessionStore].
abstract interface class SessionStore {
  /// Reads the persisted session, or `null` if none is stored / it is
  /// unreadable. Never throws — a storage failure is treated as "signed out".
  Future<AuthSession?> read();

  /// The email of the persisted session, for pre-filling the login form.
  Future<String?> readEmail();

  /// Persists [session] (and optionally the [email] that created it).
  Future<void> save(AuthSession session, {String? email});

  /// Removes all persisted auth state (logout).
  Future<void> clear();
}

/// [SessionStore] backed by the OS keystore/keychain via `flutter_secure_storage`.
///
/// Tokens are sensitive, so they live in encrypted platform storage rather than
/// `SharedPreferences`. Reads degrade to `null` on any storage error so a
/// corrupt/locked store never blocks app startup.
class SecureSessionStore implements SessionStore {
  SecureSessionStore({FlutterSecureStorage? storage})
      : _storage = storage ?? const FlutterSecureStorage();

  final FlutterSecureStorage _storage;

  static const _kToken = 'auth_token';
  static const _kRefresh = 'auth_refresh_token';
  static const _kExpires = 'auth_expires_at';
  static const _kEmail = 'auth_email';

  @override
  Future<AuthSession?> read() async {
    try {
      final token = await _storage.read(key: _kToken);
      final refresh = await _storage.read(key: _kRefresh);
      final expires = await _storage.read(key: _kExpires);
      if (token == null || refresh == null || expires == null) return null;
      return AuthSession.fromJson({
        'token': token,
        'refresh_token': refresh,
        'expires_at': expires,
      });
    } catch (_) {
      return null;
    }
  }

  @override
  Future<String?> readEmail() async {
    try {
      return await _storage.read(key: _kEmail);
    } catch (_) {
      return null;
    }
  }

  @override
  Future<void> save(AuthSession session, {String? email}) async {
    await _storage.write(key: _kToken, value: session.token);
    await _storage.write(key: _kRefresh, value: session.refreshToken);
    await _storage.write(
        key: _kExpires, value: session.expiresAt.toIso8601String());
    if (email != null) {
      await _storage.write(key: _kEmail, value: email);
    }
  }

  @override
  Future<void> clear() async {
    await _storage.delete(key: _kToken);
    await _storage.delete(key: _kRefresh);
    await _storage.delete(key: _kExpires);
    await _storage.delete(key: _kEmail);
  }
}
