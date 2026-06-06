/// Typed transport/error model for API calls.
///
/// The data layer translates HTTP status codes and transport failures into a
/// single [ApiException] carrying a user-safe [message]. Screens render
/// `e.message` directly in `MMErrorState` — they never inspect raw status
/// codes or `http` exceptions.
library;

/// The category of an API failure, used by the UI to pick an icon/treatment.
enum ApiErrorKind {
  /// Could not reach the server (no connection, DNS, refused, etc.).
  network,

  /// The request itself was rejected as invalid (HTTP 400).
  badRequest,

  /// Authentication failed or the credentials/token are invalid (HTTP 401).
  unauthorized,

  /// The request conflicts with existing state, e.g. a duplicate email (HTTP 409).
  conflict,

  /// The requested resource does not exist (HTTP 404).
  notFound,

  /// The upstream took too long (HTTP 504) — search-specific timeout.
  timeout,

  /// Any 5xx or otherwise unexpected response.
  server,
}

/// An error surfaced from the data layer with a message safe to show users.
class ApiException implements Exception {
  const ApiException(this.kind, this.message);

  final ApiErrorKind kind;
  final String message;

  @override
  String toString() => 'ApiException(${kind.name}): $message';
}
