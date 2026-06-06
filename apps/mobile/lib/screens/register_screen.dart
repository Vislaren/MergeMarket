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

/// Register screen — "Create Account" (USER_FLOWS Flow 1).
///
/// Standalone route (no bottom navigation bar). Validates email / password /
/// confirm-password, calls [AuthController.register], persists the session via
/// secure storage, and on success navigates to Home. Failures surface inline as
/// a banner using the [ApiException] message (e.g. 409 email-exists).
class RegisterScreen extends ConsumerStatefulWidget {
  const RegisterScreen({super.key});

  @override
  ConsumerState<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends ConsumerState<RegisterScreen> {
  final _formKey = GlobalKey<FormState>();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  final _confirmController = TextEditingController();

  bool _obscurePassword = true;
  bool _obscureConfirm = true;
  bool _submitting = false;
  String? _error;

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    _confirmController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    setState(() => _error = null);
    if (!_formKey.currentState!.validate()) return;

    setState(() => _submitting = true);
    try {
      await ref.read(authControllerProvider.notifier).register(
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
      backgroundColor: backgroundLight,
      appBar: AppBar(title: const Text('Create Account'), centerTitle: true),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(lg),
          child: Card(
            child: Padding(
              padding: const EdgeInsets.all(lg),
              child: Form(
                key: _formKey,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
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
                        hintText: 'name@company.com',
                      ),
                      validator: validateEmail,
                    ),
                    const SizedBox(height: md),
                    Text('Password',
                        style: headingSmall.copyWith(color: textPrimary)),
                    const SizedBox(height: sm),
                    TextFormField(
                      controller: _passwordController,
                      obscureText: _obscurePassword,
                      textInputAction: TextInputAction.next,
                      autofillHints: const [AutofillHints.newPassword],
                      decoration: InputDecoration(
                        prefixIcon: const Icon(Icons.lock_outline_rounded),
                        hintText: 'At least $kMinPasswordLength characters',
                        suffixIcon: IconButton(
                          icon: Icon(_obscurePassword
                              ? Icons.visibility_off_outlined
                              : Icons.visibility_outlined),
                          onPressed: () => setState(
                              () => _obscurePassword = !_obscurePassword),
                          tooltip: _obscurePassword
                              ? 'Show password'
                              : 'Hide password',
                        ),
                      ),
                      validator: validateNewPassword,
                    ),
                    const SizedBox(height: md),
                    Text('Confirm Password',
                        style: headingSmall.copyWith(color: textPrimary)),
                    const SizedBox(height: sm),
                    TextFormField(
                      controller: _confirmController,
                      obscureText: _obscureConfirm,
                      textInputAction: TextInputAction.done,
                      onFieldSubmitted: (_) => _submit(),
                      decoration: InputDecoration(
                        prefixIcon: const Icon(Icons.lock_reset_rounded),
                        hintText: 'Re-enter your password',
                        suffixIcon: IconButton(
                          icon: Icon(_obscureConfirm
                              ? Icons.visibility_off_outlined
                              : Icons.visibility_outlined),
                          onPressed: () => setState(
                              () => _obscureConfirm = !_obscureConfirm),
                          tooltip: _obscureConfirm
                              ? 'Show password'
                              : 'Hide password',
                        ),
                      ),
                      validator: (value) => validateConfirmPassword(
                          value, _passwordController.text),
                    ),
                    const SizedBox(height: lg),
                    SizedBox(
                      width: double.infinity,
                      child: ElevatedButton(
                        onPressed: _submitting ? null : _submit,
                        child: _submitting
                            ? const AuthButtonSpinner()
                            : const Text('Create Account'),
                      ),
                    ),
                    const SizedBox(height: lg),
                    const AuthOrDivider(label: 'OR CONTINUE WITH'),
                    const SizedBox(height: lg),
                    SizedBox(
                      width: double.infinity,
                      child: OutlinedButton.icon(
                        onPressed: () =>
                            _stub('Google sign-in is coming soon.'),
                        icon: const Icon(Icons.g_mobiledata_rounded, size: 28),
                        label: const Text('Google'),
                      ),
                    ),
                    const SizedBox(height: lg),
                    Center(
                      child: Wrap(
                        crossAxisAlignment: WrapCrossAlignment.center,
                        children: [
                          Text('Already have an account? ',
                              style:
                                  bodyRegular.copyWith(color: textSecondary)),
                          GestureDetector(
                            onTap: () => context.go(Routes.login),
                            child: Text('Log in',
                                style:
                                    headingSmall.copyWith(color: accentRed)),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
