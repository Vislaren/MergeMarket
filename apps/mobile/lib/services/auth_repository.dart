import 'dart:convert';

import 'package:http/http.dart' as http;

import '../config/app_config.dart';
import '../models/auth_session.dart';
import 'api_exception.dart';

/// Data-layer access to the Auth contract (`/api/v1/auth/*`).
///
/// The only component that knows the auth API shape: it registers, logs in, and
/// refreshes a session, mapping the contract's status codes to a typed
/// [ApiException]. The injected [http.Client] makes it unit testable with
/// `package:http`'s `MockClient`.
///
/// Until Agent A's real Auth service (A-08) is deployed, this runs against the
/// B-02 mock server, whose sentinel inputs are honoured here:
/// `password == "wrongpassword"` → 401, `email == "taken@mergemarket.app"` →
/// 409, `refresh_token == "expired"` → 401.
class AuthRepository {
  AuthRepository({
    required http.Client client,
    AppConfig config = AppConfig.env,
  })  : _client = client,
        _config = config;

  final http.Client _client;
  final AppConfig _config;

  Uri _uri(String path) => Uri.parse('${_config.apiBaseUrl}$path');

  static const Map<String, String> _jsonHeaders = {
    'Accept': 'application/json',
    'Content-Type': 'application/json',
  };

  /// Registers a new account. Returns the new [AuthSession] on 201; throws
  /// [ApiException] for 400 (invalid input), 409 (email already exists), any
  /// other status, or a transport error.
  Future<AuthSession> register(String email, String password) async {
    final http.Response response;
    try {
      response = await _client.post(
        _uri('/api/v1/auth/register'),
        headers: _jsonHeaders,
        body: jsonEncode({'email': email, 'password': password}),
      );
    } on Exception {
      throw _networkError;
    }
    switch (response.statusCode) {
      case 201:
        return _decode(response);
      case 400:
        throw ApiException(
          ApiErrorKind.badRequest,
          _messageOr(response, 'Please enter a valid email and password.'),
        );
      case 409:
        throw ApiException(
          ApiErrorKind.conflict,
          _messageOr(response, 'An account with this email already exists.'),
        );
      default:
        throw ApiException(
          ApiErrorKind.server,
          _messageOr(response, 'Could not create your account. Please try again.'),
        );
    }
  }

  /// Logs in with [email]/[password]. Returns the [AuthSession] on 200; throws
  /// [ApiException] for 401 (bad credentials), any other status, or a transport
  /// error.
  Future<AuthSession> login(String email, String password) async {
    final http.Response response;
    try {
      response = await _client.post(
        _uri('/api/v1/auth/login'),
        headers: _jsonHeaders,
        body: jsonEncode({'email': email, 'password': password}),
      );
    } on Exception {
      throw _networkError;
    }
    switch (response.statusCode) {
      case 200:
        return _decode(response);
      case 401:
        throw ApiException(
          ApiErrorKind.unauthorized,
          _messageOr(response, 'Email or password is incorrect.'),
        );
      default:
        throw ApiException(
          ApiErrorKind.server,
          _messageOr(response, 'Could not log you in. Please try again.'),
        );
    }
  }

  /// Exchanges a [refreshToken] for a fresh [AuthSession]. Returns it on 200;
  /// throws [ApiException] for 401 (token expired/invalid), any other status,
  /// or a transport error.
  Future<AuthSession> refresh(String refreshToken) async {
    final http.Response response;
    try {
      response = await _client.post(
        _uri('/api/v1/auth/refresh'),
        headers: _jsonHeaders,
        body: jsonEncode({'refresh_token': refreshToken}),
      );
    } on Exception {
      throw _networkError;
    }
    switch (response.statusCode) {
      case 200:
        return _decode(response);
      case 401:
        throw ApiException(
          ApiErrorKind.unauthorized,
          _messageOr(response, 'Your session has expired. Please log in again.'),
        );
      default:
        throw ApiException(
          ApiErrorKind.server,
          _messageOr(response, 'Could not refresh your session. Please try again.'),
        );
    }
  }

  AuthSession _decode(http.Response response) =>
      AuthSession.fromJson(jsonDecode(response.body) as Map<String, dynamic>);

  static const ApiException _networkError = ApiException(
    ApiErrorKind.network,
    'Could not reach the server. Check your connection and try again.',
  );

  /// Extracts the contract's `{error, message}` human message, falling back to
  /// [fallback] when the body is absent or not the expected error shape.
  String _messageOr(http.Response response, String fallback) {
    try {
      final body = jsonDecode(response.body) as Map<String, dynamic>;
      final message = body['message'] as String?;
      if (message != null && message.isNotEmpty) return message;
    } catch (_) {
      // Fall through to the fallback message.
    }
    return fallback;
  }
}
