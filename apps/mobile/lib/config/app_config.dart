/// Compile-time application configuration.
///
/// Values are read from `--dart-define` flags so the same build can point at
/// the B-02 mock server (default) or a real backend without code changes —
/// per the "config from environment, never hardcoded" standard shared with
/// Agent A's Go services.
///
/// Targets (see `.agents/Agent_B/PORTS_README.md`):
/// - Local dev (default): the B-02 mock server on `:8089`.
/// - Real backend: Kong gateway on `:8088` (e.g. `http://95.111.228.35:8088`).
/// - Android emulator: the host is reachable at `10.0.2.2`, not `localhost`.
///
/// Example:
/// ```
/// flutter run --dart-define=API_BASE_URL=http://95.111.228.35:8088   # real backend via Kong
/// flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8089        # mock from Android emulator
/// ```
class AppConfig {
  const AppConfig({required this.apiBaseUrl, required this.defaultLocation});

  /// Base URL of the API (Kong gateway in prod, mock-server in dev).
  ///
  /// Defaults to the B-02 mock server's assigned port (`8089` per
  /// PORTS_README — `8080` is a blocked coolify-proxy port). Override the
  /// `API_BASE_URL` define to point at Kong for the real backend.
  final String apiBaseUrl;

  /// ISO 3166-1 alpha-2 country code sent as the `location` query parameter.
  /// `CM` (Cameroon) matches the mock server's XAF sample locale.
  final String defaultLocation;

  /// The configuration resolved from `--dart-define` values at build time.
  static const AppConfig env = AppConfig(
    apiBaseUrl: String.fromEnvironment(
      'API_BASE_URL',
      defaultValue: 'http://localhost:8089',
    ),
    defaultLocation: String.fromEnvironment(
      'DEFAULT_LOCATION',
      defaultValue: 'CM',
    ),
  );
}
