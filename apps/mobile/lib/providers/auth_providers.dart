import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/auth_session.dart';
import '../services/auth_repository.dart';
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
