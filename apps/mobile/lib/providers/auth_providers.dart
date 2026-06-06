import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:http/http.dart' as http;

import '../models/auth_session.dart';
import '../services/api_exception.dart';
import '../services/auth_repository.dart';
import '../services/authenticated_client.dart';
import '../services/session_store.dart';
import 'search_providers.dart';

/// Where the auth session is persisted. Overridden in tests with an in-memory
/// fake, since `flutter_secure_storage` needs a platform channel.
final sessionStoreProvider = Provider<SessionStore>((ref) {
  return SecureSessionStore();
});

/// The Auth data-layer repository, wired to the shared HTTP client from
/// [httpClientProvider].
final authRepositoryProvider = Provider<AuthRepository>((ref) {
  return AuthRepository(client: ref.watch(httpClientProvider));
});

/// Immutable snapshot of who is signed in.
///
/// A `null` [session] (or an expired one) means signed out. [email] is kept for
/// pre-filling the login form and showing the account affordance.
class AuthState {
  const AuthState({this.session, this.email});

  final AuthSession? session;
  final String? email;

  /// True when a non-expired session is present.
  bool get isAuthenticated => session != null && !session!.isExpired;

  /// The signed-out state.
  static const AuthState signedOut = AuthState();
}

/// Owns the authentication lifecycle: restoring a persisted session on startup,
/// logging in / registering / out, and keeping [sessionStoreProvider] in sync.
///
/// Exposed as an [AsyncNotifier] so the UI (and the router guard) can react to
/// the restore happening asynchronously at launch. Mutations rethrow
/// [ApiException] to the caller so screens can surface the message inline.
class AuthController extends AsyncNotifier<AuthState> {
  @override
  Future<AuthState> build() async {
    final store = ref.read(sessionStoreProvider);
    final session = await store.read();
    if (session == null || session.isExpired) {
      return AuthState(email: await store.readEmail());
    }
    return AuthState(session: session, email: await store.readEmail());
  }

  /// Logs in and persists the session. Rethrows [ApiException] on failure.
  Future<void> login(String email, String password) async {
    final session = await ref.read(authRepositoryProvider).login(email, password);
    await ref.read(sessionStoreProvider).save(session, email: email);
    state = AsyncData(AuthState(session: session, email: email));
  }

  /// Registers, persists the session, and signs the new user in. Rethrows
  /// [ApiException] on failure.
  Future<void> register(String email, String password) async {
    final session =
        await ref.read(authRepositoryProvider).register(email, password);
    await ref.read(sessionStoreProvider).save(session, email: email);
    state = AsyncData(AuthState(session: session, email: email));
  }

  /// Exchanges the current refresh token for a fresh session, persisting it and
  /// updating [state]. Returns `true` on success.
  ///
  /// On failure (the refresh token is expired/invalid) it clears the session
  /// and returns to the signed-out state, returning `false` — the caller (the
  /// [AuthenticatedClient] interceptor) then lets the original 401 surface so
  /// the router guard can route to login. Used by the refresh-on-401 path; not
  /// called directly by screens.
  Future<bool> refreshSession() async {
    final current = state.value?.session;
    if (current == null) return false;
    final email = state.value?.email;
    try {
      final session =
          await ref.read(authRepositoryProvider).refresh(current.refreshToken);
      await ref.read(sessionStoreProvider).save(session, email: email);
      state = AsyncData(AuthState(session: session, email: email));
      return true;
    } on ApiException {
      await ref.read(sessionStoreProvider).clear();
      state = AsyncData(AuthState(email: email));
      return false;
    }
  }

  /// Clears the persisted session and returns to the signed-out state.
  Future<void> logout() async {
    await ref.read(sessionStoreProvider).clear();
    state = const AsyncData(AuthState.signedOut);
  }
}

/// The app-wide authentication state.
final authControllerProvider =
    AsyncNotifierProvider<AuthController, AuthState>(AuthController.new);

/// Convenience: `true` once a non-expired session has been restored or created.
/// Defaults to `false` while the session is still loading or on any error, so
/// the app is safely treated as signed-out until proven otherwise.
final isAuthenticatedProvider = Provider<bool>((ref) {
  return ref.watch(authControllerProvider).value?.isAuthenticated ?? false;
});

/// The HTTP client for **protected** data-layer calls (everything except auth):
/// the shared [httpClientProvider] wrapped in an [AuthenticatedClient] so every
/// request carries the Bearer token and a 401 triggers a single refresh + replay.
///
/// This is the seam B-11 swaps in: against the B-02 mock (no auth) it is a
/// harmless pass-through when signed out; against Kong (A-09) it satisfies the
/// JWT plugin and survives access-token expiry. The auth repository keeps using
/// the bare [httpClientProvider] — it must not require or refresh a token.
final authedHttpClientProvider = Provider<http.Client>((ref) {
  return AuthenticatedClient(
    inner: ref.watch(httpClientProvider),
    readToken: () => ref.read(authControllerProvider).value?.session?.token,
    refreshToken: () => ref.read(authControllerProvider.notifier).refreshSession(),
  );
});
