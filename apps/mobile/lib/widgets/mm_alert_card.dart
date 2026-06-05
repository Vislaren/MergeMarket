import 'package:flutter/material.dart';

import '../theme/colours.dart';
import '../theme/spacing.dart';
import '../theme/typography.dart';
import '../utils/money.dart';

/// A single price-alert row on the Alerts screen.
///
/// `COMPONENT_LIBRARY.md → MMAlertCard`. Shows the product title, the threshold
/// price in bold, an active/inactive status chip, and a delete swipe action.
class MMAlertCard extends StatelessWidget {
  const MMAlertCard({
    super.key,
    required this.alertId,
    required this.productTitle,
    required this.thresholdPrice,
    required this.currency,
    required this.isActive,
    required this.onDelete,
  });

  final String alertId;
  final String productTitle;
  final double thresholdPrice;
  final String currency;
  final bool isActive;
  final VoidCallback onDelete;

  @override
  Widget build(BuildContext context) {
    return Dismissible(
      key: ValueKey(alertId),
      direction: DismissDirection.endToStart,
      onDismissed: (_) => onDelete(),
      background: Container(
        margin: const EdgeInsets.only(bottom: md),
        alignment: Alignment.centerRight,
        padding: const EdgeInsets.only(right: lg),
        decoration: BoxDecoration(
          color: accentRed,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: const [
            Icon(Icons.delete_outline_rounded, color: surfaceWhite),
            Text('Delete',
                style: TextStyle(color: surfaceWhite, fontSize: 11)),
          ],
        ),
      ),
      child: Container(
        margin: const EdgeInsets.only(bottom: md),
        padding: const EdgeInsets.all(md),
        decoration: BoxDecoration(
          color: surfaceWhite,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: borderGrey),
        ),
        child: Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    productTitle,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: headingSmall.copyWith(color: textPrimary),
                  ),
                  const SizedBox(height: xs),
                  Row(
                    children: [
                      Text('Alert below ',
                          style: bodySmall.copyWith(color: textSecondary)),
                      Text(
                        formatMoney(thresholdPrice, currency),
                        style: bodyRegular.copyWith(
                          color: textPrimary,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
            const SizedBox(width: sm),
            _StatusChip(isActive: isActive),
          ],
        ),
      ),
    );
  }
}

/// Active / Inactive status pill.
class _StatusChip extends StatelessWidget {
  const _StatusChip({required this.isActive});

  final bool isActive;

  @override
  Widget build(BuildContext context) {
    final colour = isActive ? successGreen : textSecondary;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: sm, vertical: xs),
      decoration: BoxDecoration(
        color: colour.withValues(alpha: 0.15),
        borderRadius: BorderRadius.circular(100),
      ),
      child: Text(
        isActive ? 'Active' : 'Inactive',
        style: labelBold.copyWith(color: colour),
      ),
    );
  }
}
