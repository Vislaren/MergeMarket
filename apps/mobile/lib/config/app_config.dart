/// Compile-time application configuration.
///
/// Values are read from `--dart-define` flags so the same build can point at
/// the B-02 mock server (default) or a real backend without code changes —
/// per the "config from environment, never hardcoded" standard shared with
/// Agent A's Go services.
///
/// Example:
/// ```
/// flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8080
/// ```
class AppConfig {
  const AppConfig({required this.apiBaseUrl, required this.defaultLocation});

  /// Base URL of the API (Kong gateway in prod, mock-server in dev).
  ///
  /// Defaults to the B-02 mock server's documented port (8080). On the Android
  /// emulator the host machine is reachable at `10.0.2.2`, so override this
  /// define when running on Android.
  final String apiBaseUrl;

  /// ISO 3166-1 alpha-2 country code sent as the `location` query parameter.
  /// `CM` (Cameroon) matches the mock server's XAF sample locale.
  final String defaultLocation;

  /// The configuration resolved from `--dart-define` values at build time.
  static const AppConfig env = AppConfig(
    apiBaseUrl: String.fromEnvironment(
      'API_BASE_URL',
      defaultValue: 'http://localhost:8080',
    ),
    defaultLocation: String.fromEnvironment(
      'DEFAULT_LOCATION',
      defaultValue: 'CM',
    ),
  );
}
