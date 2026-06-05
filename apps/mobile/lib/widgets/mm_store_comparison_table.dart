import 'package:flutter/material.dart';

import '../models/store_result.dart';
import '../theme/colours.dart';
import '../theme/spacing.dart';
import '../theme/typography.dart';
import '../utils/money.dart';

/// Store comparison table for the Product Detail screen.
///
/// `COMPONENT_LIBRARY.md → MMStoreComparisonTable`. Rows are sorted by total
/// cost ascending; the cheapest (best-deal) row gets a subtle green background,
/// and every row has a "Go to Store" action. [onGoToStore] follows the
/// documented signature: it takes a [StoreResult] and returns the callback to
/// run for that row.
class MMStoreComparisonTable extends StatelessWidget {
  const MMStoreComparisonTable({
    super.key,
    required this.stores,
    required this.currency,
    required this.onGoToStore,
  });

  final List<StoreResult> stores;
  final String currency;
  final VoidCallback Function(StoreResult) onGoToStore;

  @override
  Widget build(BuildContext context) {
    if (stores.isEmpty) {
      return Text(
        'No store offers available for this product.',
        style: bodySmall.copyWith(color: textSecondary),
      );
    }

    final sorted = [...stores]
      ..sort((a, b) => a.totalCost.compareTo(b.totalCost));

    return Column(
      children: [
        for (var i = 0; i < sorted.length; i++)
          _StoreRow(
            store: sorted[i],
            currency: currency,
            isBest: i == 0,
            onGoToStore: onGoToStore(sorted[i]),
          ),
      ],
    );
  }
}

/// One store row: name + delivery hint on the left, total cost and a
/// "Go to Store" chip on the right. Highlighted when it is the best deal.
class _StoreRow extends StatelessWidget {
  const _StoreRow({
    required this.store,
    required this.currency,
    required this.isBest,
    required this.onGoToStore,
  });

  final StoreResult store;
  final String currency;
  final bool isBest;
  final VoidCallback onGoToStore;

  @override
  Widget build(BuildContext context) {
    final shippingLabel = store.shipping <= 0
        ? 'Free shipping'
        : '+ ${formatMoney(store.shipping, currency)} shipping';

    return Container(
      margin: const EdgeInsets.only(bottom: sm),
      padding: const EdgeInsets.all(md),
      decoration: BoxDecoration(
        color: isBest ? successGreen.withValues(alpha: 0.10) : surfaceWhite,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: isBest ? successGreen : borderGrey),
      ),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Text(store.store,
                        style: headingSmall.copyWith(color: textPrimary)),
                    if (isBest) ...[
                      const SizedBox(width: sm),
                      _BestChip(),
                    ],
                  ],
                ),
                const SizedBox(height: xs),
                Text(shippingLabel,
                    style: bodySmall.copyWith(color: textSecondary)),
              ],
            ),
          ),
          const SizedBox(width: sm),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Text(
                formatMoney(store.totalCost, currency),
                style: headingMedium.copyWith(
                    color: isBest ? successGreen : textPrimary),
              ),
              const SizedBox(height: xs),
              GestureDetector(
                onTap: onGoToStore,
                child: Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: md, vertical: xs),
                  decoration: BoxDecoration(
                    color: primaryNavy,
                    borderRadius: BorderRadius.circular(100),
                  ),
                  child: Text('Go to Store',
                      style: labelBold.copyWith(color: surfaceWhite)),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

/// "Best deal" pill shown on the cheapest row.
class _BestChip extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: sm, vertical: 2),
      decoration: BoxDecoration(
        color: successGreen,
        borderRadius: BorderRadius.circular(100),
      ),
      child: Text('Best deal', style: labelBold.copyWith(color: surfaceWhite)),
    );
  }
}
