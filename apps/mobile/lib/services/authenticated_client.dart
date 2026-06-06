import 'dart:async';

import 'package:http/http.dart' as http;

/// Reads the current access token, or `null` when signed out.
typedef TokenReader = String? Function();

/// Refreshes the session, returning `true` on success (a new token is now
/// available via the [TokenReader]) or `false` when the session is gone.
typedef TokenRefresher = Future<bool> Function();

/// An [http.Client] decorator that makes the app talk to the **real backend**
/// (Kong gateway, B-11) instead of the auth-free B-02 mock server.
///
/// It does two things the mock never required but Kong does:
///
/// 1. **Bearer auth** — attaches `Authorization: Bearer <token>` to every
///    request when a session exists. Kong's JWT plugin (A-09) rejects
///    unauthenticated calls to all non-auth routes with 401.
/// 2. **Refresh on 401** — when a protected request comes back 401 *and* a
///    token was sent, it refreshes the session once (via [TokenRefresher]) and
///    replays the original request with the new token. A failed refresh leaves
///    the 401 to surface so the UI can bounce the user to login. This is the
///    token-refresh interceptor deferred from B-08.
///
/// Auth endpoints (`/api/v1/auth/*`) are never refreshed or auth-decorated:
/// they mint tokens, so a 401 there is a real credential failure, not an
/// expired session.
///
/// The dependencies are plain closures (not Riverpod) so the client is unit
/// testable with `package:http`'s `MockClient` and a fake token store.
class AuthenticatedClient extends http.BaseClient {
  /// Wraps [inner], reading the access token via [readToken] and refreshing via
  /// [refreshToken].
  AuthenticatedClient({
    required http.Client inner,
    required TokenReader readToken,
    required TokenRefresher refreshToken,
  })  : _inner = inner,
        _readToken = readToken,
        _refreshToken = refreshToken;

  final http.Client _inner;
  final TokenReader _readToken;
  final TokenRefresher _refreshToken;

  /// Coalesces concurrent refreshes so a burst of parallel 401s triggers a
  /// single refresh rather than one per in-flight request.
  Future<bool>? _inFlightRefresh;

  static const String _authPathFragment = '/api/v1/auth/';

  @override
  Future<http.StreamedResponse> send(http.BaseRequest request) async {
    // Buffer the body once so the request can be safely replayed after a
    // refresh (a finalized BaseRequest cannot be re-sent).
    final List<int>? bytes = request is http.Request ? request.bodyBytes : null;

    final String? token = _readToken();
    final bool isAuthRoute = request.url.path.contains(_authPathFragment);

    var response = await _inner.send(_clone(request, bytes, token));

    if (response.statusCode != 401 || token == null || isAuthRoute) {
      return response;
    }

    // Free the connection: the 401 body is discarded before any replay.
    await response.stream.drain<void>();

    // Another request may have already refreshed while this one was in flight;
    // if the current token differs, just retry with it rather than spending a
    // (rotated) refresh token a second time.
    if (_readToken() != token) {
      return _inner.send(_clone(request, bytes, _readToken()));
    }

    final bool refreshed = await _coalescedRefresh();
    if (!refreshed) {
      // Re-issue the original request so the caller still receives a 401 to act
      // on (the first response's stream was already drained).
      return _inner.send(_clone(request, bytes, token));
    }
    return _inner.send(_clone(request, bytes, _readToken()));
  }

  Future<bool> _coalescedRefresh() {
    return _inFlightRefresh ??=
        _refreshToken().whenComplete(() => _inFlightRefresh = null);
  }

  /// Builds a fresh, replayable request from [original], optionally attaching a
  /// Bearer [token].
  http.BaseRequest _clone(
    http.BaseRequest original,
    List<int>? bytes,
    String? token,
  ) {
    if (original is http.Request) {
      final copy = http.Request(original.method, original.url)
        ..followRedirects = original.followRedirects
        ..maxRedirects = original.maxRedirects
        ..persistentConnection = original.persistentConnection
        ..headers.addAll(original.headers);
      if (bytes != null) copy.bodyBytes = bytes;
      if (token != null) copy.headers['Authorization'] = 'Bearer $token';
      return copy;
    }
    // Streamed/multipart requests are not used in this app and cannot be
    // replayed; forward as-is with the auth header attached.
    if (token != null) original.headers['Authorization'] = 'Bearer $token';
    return original;
  }

  @override
  void close() => _inner.close();
}
