import 'package:flutter/material.dart';

import '../theme/colours.dart';
import '../theme/spacing.dart';
import '../theme/typography.dart';

/// Small presentational helpers shared by the Login and Register screens (B-08).
///
/// These are auth-specific compositions (not reusable design-system components),
/// so they live here rather than in `COMPONENT_LIBRARY.md`.

/// Inline error banner shown above an auth form when a submit fails. [message]
/// is a user-safe string, typically an `ApiException.message`.
class AuthErrorBanner extends StatelessWidget {
  const AuthErrorBanner({super.key, required this.message});

  final String message;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(md),
      decoration: BoxDecoration(
        color: accentRed.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: accentRed),
      ),
      child: Row(
        children: [
          const Icon(Icons.error_outline_rounded, color: accentRed, size: 20),
          const SizedBox(width: sm),
          Expanded(
            child:
                Text(message, style: bodyRegular.copyWith(color: accentRed)),
          ),
        ],
      ),
    );
  }
}

/// Horizontal divider with a centred [label] (e.g. "OR", "OR CONTINUE WITH").
class AuthOrDivider extends StatelessWidget {
  const AuthOrDivider({super.key, required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        const Expanded(child: Divider()),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: md),
          child: Text(label, style: bodySmall.copyWith(color: textSecondary)),
        ),
        const Expanded(child: Divider()),
      ],
    );
  }
}

/// White spinner sized to sit inside a primary button while a request is
/// in flight (replaces the button label).
class AuthButtonSpinner extends StatelessWidget {
  const AuthButtonSpinner({super.key});

  @override
  Widget build(BuildContext context) {
    return const SizedBox(
      height: 20,
      width: 20,
      child: CircularProgressIndicator(strokeWidth: 2, color: surfaceWhite),
    );
  }
}
