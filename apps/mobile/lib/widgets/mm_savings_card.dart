import 'package:flutter/material.dart';

import '../theme/colours.dart';
import '../theme/spacing.dart';
import '../theme/typography.dart';
import '../utils/money.dart';

/// Gamified cumulative savings card for the Savings Dashboard.
///
/// `COMPONENT_LIBRARY.md -> MMSavingsCard`. Shows a large total saved amount,
/// the user's savings level, progress to the next level, and a share action.
class MMSavingsCard extends StatelessWidget {
  const MMSavingsCard({
    super.key,
    required this.totalSaved,
    required this.currency,
    required this.savingsLevel,
    required this.progressToNextLevel,
    this.remainingToNextLevel,
    this.onShare,
  });

  final double totalSaved;
  final String currency;
  final int savingsLevel;
  final double progressToNextLevel;
  final double? remainingToNextLevel;
  final VoidCallback? onShare;

  @override
  Widget build(BuildContext context) {
    final progress = progressToNextLevel.clamp(0.0, 1.0);
    return Container(
      padding: const EdgeInsets.all(lg),
      decoration: BoxDecoration(
        color: primaryNavy,
        borderRadius: BorderRadius.circular(12),
        boxShadow: [
          BoxShadow(
            color: primaryNavy.withValues(alpha: 0.18),
            blurRadius: 12,
            offset: const Offset(0, 6),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Total lifetime savings',
                      style: bodyRegular.copyWith(
                        color: surfaceWhite.withValues(alpha: 0.68),
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                    const SizedBox(height: sm),
                    Text(
                      formatMoney(totalSaved, currency),
                      style: headingLarge.copyWith(
                        color: surfaceWhite,
                        fontSize: 34,
                      ),
                    ),
                  ],
                ),
              ),
              IconButton.filled(
                tooltip: 'Share savings',
                onPressed: onShare,
                style: IconButton.styleFrom(
                  backgroundColor: surfaceWhite.withValues(alpha: 0.12),
                  foregroundColor: surfaceWhite,
                ),
                icon: const Icon(Icons.ios_share_rounded),
              ),
            ],
          ),
          const SizedBox(height: xl),
          Row(
            children: [
              _LevelBadge(level: savingsLevel),
              const SizedBox(width: sm),
              Expanded(
                child: Text(
                  _remainingLabel(),
                  textAlign: TextAlign.end,
                  style: labelBold.copyWith(color: dealGold),
                ),
              ),
            ],
          ),
          const SizedBox(height: sm),
          ClipRRect(
            borderRadius: BorderRadius.circular(100),
            child: LinearProgressIndicator(
              value: progress,
              minHeight: 10,
              color: dealGold,
              backgroundColor: surfaceWhite.withValues(alpha: 0.12),
            ),
          ),
        ],
      ),
    );
  }

  String _remainingLabel() {
    if (savingsLevel >= 10) return 'Top level reached';
    final remaining = remainingToNextLevel;
    if (remaining == null) return 'Next level';
    return '${formatMoney(remaining, currency)} to next level';
  }
}

class _LevelBadge extends StatelessWidget {
  const _LevelBadge({required this.level});

  final int level;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: sm, vertical: xs),
      decoration: BoxDecoration(
        color: successGreen.withValues(alpha: 0.18),
        borderRadius: BorderRadius.circular(100),
      ),
      child: Text('Level $level', style: labelBold.copyWith(color: dealGold)),
    );
  }
}
