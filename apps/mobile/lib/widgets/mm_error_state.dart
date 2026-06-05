import 'package:flutter/material.dart';

import '../theme/colours.dart';
import '../theme/spacing.dart';
import '../theme/typography.dart';

/// Centred error panel with a retry action.
///
/// `COMPONENT_LIBRARY.md → MMErrorState`. Shown on any screen whose data load
/// fails; [message] is a user-safe string (e.g. `ApiException.message`).
class MMErrorState extends StatelessWidget {
  const MMErrorState({
    super.key,
    required this.message,
    required this.onRetry,
    this.icon = Icons.wifi_off_rounded,
  });

  final String message;
  final VoidCallback onRetry;
  final IconData icon;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(lg),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 48, color: textSecondary),
            const SizedBox(height: md),
            Text(
              message,
              textAlign: TextAlign.center,
              style: bodyRegular.copyWith(color: textSecondary),
            ),
            const SizedBox(height: lg),
            ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: primaryNavy,
                foregroundColor: surfaceWhite,
              ),
              onPressed: onRetry,
              child: const Text('Try Again'),
            ),
          ],
        ),
      ),
    );
  }
}
