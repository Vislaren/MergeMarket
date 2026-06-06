/// Data model for an authenticated session (`/api/v1/auth/*`).
///
/// JSON keys mirror `project_docs/api/API_CONTRACTS.md` exactly, so this decodes
/// identically against the B-02 mock server and the real Auth service (A-08)
/// later. The data layer is the only place that parses JSON; screens and the
/// auth controller consume the typed model only.
library;

/// A token bundle returned by register / login / refresh.
class AuthSession {
  const AuthSession({
    required this.token,
    required this.refreshToken,
    required this.expiresAt,
  });

  /// Short-lived access token sent as `Authorization: Bearer <token>`.
  final String token;

  /// Long-lived token used to obtain a new access token at `/auth/refresh`.
  final String refreshToken;

  /// Absolute expiry of [token]. A session is treated as signed-out once past.
  final DateTime expiresAt;

  /// Whether [token] has expired (and so a refresh/login is required).
  bool get isExpired => !DateTime.now().isBefore(expiresAt);

  /// Decodes a token bundle. Missing or mistyped fields fall back to safe
  /// values (an unparseable `expires_at` becomes the epoch, i.e. expired).
  factory AuthSession.fromJson(Map<String, dynamic> json) {
    return AuthSession(
      token: json['token'] as String? ?? '',
      refreshToken: json['refresh_token'] as String? ?? '',
      expiresAt: DateTime.tryParse(json['expires_at'] as String? ?? '') ??
          DateTime.fromMillisecondsSinceEpoch(0),
    );
  }
}
