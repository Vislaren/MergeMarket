import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/auth_providers.dart';
import '../router/app_router.dart';
import '../services/api_exception.dart';
import '../theme/colours.dart';
import '../theme/spacing.dart';
import '../theme/typography.dart';
import '../utils/validators.dart';
import '../widgets/auth_form_widgets.dart';

/// Login screen — "Welcome Back" (USER_FLOWS Flow 1).
///
/// Standalone route (no bottom navigation bar). Validates the email/password
/// form, calls [AuthController.login], persists the session via secure storage,
/// and on success navigates to Home. Failures surface inline as a banner using
/// the [ApiException] message.
class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  final _formKey = GlobalKey<FormState>();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();

  bool _obscure = true;
  bool _submitting = false;
  String? _error;

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    setState(() => _error = null);
    if (!_formKey.currentState!.validate()) return;

    setState(() => _submitting = true);
    try {
      await ref.read(authControllerProvider.notifier).login(
            _emailController.text.trim(),
            _passwordController.text,
          );
      if (mounted) context.go(Routes.home);
    } on ApiException catch (e) {
      if (mounted) setState(() => _error = e.message);
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  void _stub(String message) {
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(content: Text(message)));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: surfaceWhite,
      appBar: AppBar(
        backgroundColor: surfaceWhite,
        foregroundColor: primaryNavy,
        centerTitle: true,
        title: Text('Welcome Back',
            style: headingLarge.copyWith(color: primaryNavy)),
      ),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(lg),
          child: Form(
            key: _formKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const SizedBox(height: md),
                if (_error != null) ...[
                  AuthErrorBanner(message: _error!),
                  const SizedBox(height: md),
                ],
                Text('Email Address',
                    style: headingSmall.copyWith(color: textPrimary)),
                const SizedBox(height: sm),
                TextFormField(
                  controller: _emailController,
                  keyboardType: TextInputType.emailAddress,
                  textInputAction: TextInputAction.next,
                  autofillHints: const [AutofillHints.email],
                  decoration: const InputDecoration(
                    prefixIcon: Icon(Icons.mail_outline_rounded),
                    hintText: 'Enter your email',
                  ),
                  validator: validateEmail,
                ),
                const SizedBox(height: md),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text('Password',
                        style: headingSmall.copyWith(color: textPrimary)),
                    TextButton(
                      onPressed: () => _stub(
                          'Password reset is coming soon. Contact support to reset.'),
                      child: Text('Forgot password?',
                          style: headingSmall.copyWith(color: accentRed)),
                    ),
                  ],
                ),
                const SizedBox(height: sm),
                TextFormField(
                  controller: _passwordController,
                  obscureText: _obscure,
                  textInputAction: TextInputAction.done,
                  autofillHints: const [AutofillHints.password],
                  onFieldSubmitted: (_) => _submit(),
                  decoration: InputDecoration(
                    prefixIcon: const Icon(Icons.lock_outline_rounded),
                    hintText: 'Enter your password',
                    suffixIcon: IconButton(
                      icon: Icon(_obscure
                          ? Icons.visibility_off_outlined
                          : Icons.visibility_outlined),
                      onPressed: () => setState(() => _obscure = !_obscure),
                      tooltip: _obscure ? 'Show password' : 'Hide password',
                    ),
                  ),
                  validator: validateLoginPassword,
                ),
                const SizedBox(height: lg),
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    onPressed: _submitting ? null : _submit,
                    child: _submitting
                        ? const AuthButtonSpinner()
                        : const Text('Log In'),
                  ),
                ),
                const SizedBox(height: lg),
                const AuthOrDivider(label: 'OR'),
                const SizedBox(height: lg),
                SizedBox(
                  width: double.infinity,
                  child: OutlinedButton.icon(
                    onPressed: () => _stub('Google sign-in is coming soon.'),
                    icon: const Icon(Icons.g_mobiledata_rounded, size: 28),
                    label: const Text('Continue with Google'),
                  ),
                ),
                const SizedBox(height: xl),
                Center(
                  child: Wrap(
                    crossAxisAlignment: WrapCrossAlignment.center,
                    children: [
                      Text("Don't have an account? ",
                          style: bodyRegular.copyWith(color: textSecondary)),
                      GestureDetector(
                        onTap: () => context.go(Routes.register),
                        child: Text('Register',
                            style: headingSmall.copyWith(color: accentRed)),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
