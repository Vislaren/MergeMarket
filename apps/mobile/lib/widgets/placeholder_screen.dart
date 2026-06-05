import 'package:flutter/material.dart';

import '../theme/colours.dart';
import '../theme/spacing.dart';
import '../theme/typography.dart';

/// A temporary stub used by every screen that B-01 wires into the router but
/// whose real UI lands in a later task (B-03–B-08).
///
/// It renders the screen title and the task that will replace it, so the app
/// is fully navigable end-to-end immediately after project setup. Replace the
/// whole screen file (not this widget) when implementing the real screen.
class PlaceholderScreen extends StatelessWidget {
  const PlaceholderScreen({
    super.key,
    required this.title,
    required this.implementedIn,
    this.icon = Icons.construction_rounded,
  });

  /// Human-readable screen name (matches `USER_FLOWS.md`).
  final String title;

  /// The task id that will replace this placeholder, e.g. `B-03`.
  final String implementedIn;

  final IconData icon;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(lg),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 48, color: primaryNavy),
            const SizedBox(height: md),
            Text(title, style: headingMedium.copyWith(color: textPrimary)),
            const SizedBox(height: xs),
            Text(
              'Coming soon — implemented in $implementedIn.',
              textAlign: TextAlign.center,
              style: bodySmall.copyWith(color: textSecondary),
            ),
          ],
        ),
      ),
    );
  }
}
