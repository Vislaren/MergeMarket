/// Form-field validators shared by the auth screens (B-08).
///
/// Each returns `null` when the value is acceptable, or a user-facing error
/// string otherwise — the shape `TextFormField.validator` expects. Kept pure
/// and UI-free so they are trivially unit testable.
library;

/// Minimum password length enforced when creating an account.
const int kMinPasswordLength = 8;

/// Permissive RFC-ish email check: something@something.tld with no spaces.
final RegExp _emailPattern = RegExp(r'^[^@\s]+@[^@\s]+\.[^@\s]+$');

/// Validates an email address field. Required + must look like an email.
String? validateEmail(String? value) {
  final email = value?.trim() ?? '';
  if (email.isEmpty) return 'Email is required.';
  if (!_emailPattern.hasMatch(email)) return 'Enter a valid email address.';
  return null;
}

/// Validates the password on the **login** form: presence only, since the
/// server is the authority on correctness (length rules are not re-checked on
/// an existing account).
String? validateLoginPassword(String? value) {
  if (value == null || value.isEmpty) return 'Password is required.';
  return null;
}

/// Validates the password on the **register** form: required and at least
/// [kMinPasswordLength] characters.
String? validateNewPassword(String? value) {
  final password = value ?? '';
  if (password.isEmpty) return 'Password is required.';
  if (password.length < kMinPasswordLength) {
    return 'Use at least $kMinPasswordLength characters.';
  }
  return null;
}

/// Validates the confirm-password field against [original]; must be non-empty
/// and identical.
String? validateConfirmPassword(String? value, String original) {
  if (value == null || value.isEmpty) return 'Please confirm your password.';
  if (value != original) return 'Passwords do not match.';
  return null;
}
