import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../models/savings.dart';
import '../providers/savings_providers.dart';
import '../router/app_router.dart';
import '../services/api_exception.dart';
import '../theme/colours.dart';
import '../theme/spacing.dart';
import '../theme/typography.dart';
import '../utils/money.dart';
import '../widgets/mm_error_state.dart';
import '../widgets/mm_savings_card.dart';
import '../widgets/mm_skeleton_loader.dart';

/// Savings Dashboard screen (USER_FLOWS Flow 7).
///
/// Watches [savingsProvider] and renders the four required states: loading
/// skeletons, success with [MMSavingsCard] plus savings events, empty, and error
/// with retry. All values come from `GET /api/v1/savings`.
class SavingsDashboardScreen extends ConsumerWidget {
  const SavingsDashboardScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final savings = ref.watch(savingsProvider);

    return Scaffold(
      backgroundColor: backgroundLight,
      appBar: AppBar(title: const Text('My Savings')),
      body: savings.when(
        loading: () => const _LoadingSavings(),
        error: (error, _) => MMErrorState(
          message: error is ApiException
              ? error.message
              : 'Something went wrong. Please try again.',
          onRetry: () => ref.invalidate(savingsProvider),
        ),
        data: (data) {
          if (data.transactions.isEmpty && data.totalSaved <= 0) {
            return const _EmptySavings();
          }
          return _SavingsContent(summary: data);
        },
      ),
    );
  }
}

class _SavingsContent extends StatelessWidget {
  const _SavingsContent({required this.summary});

  final SavingsSummary summary;

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: const EdgeInsets.all(md),
      children: [
        MMSavingsCard(
          totalSaved: summary.totalSaved,
          currency: summary.currency,
          savingsLevel: summary.savingsLevel,
          progressToNextLevel: summary.progressToNextLevel,
          remainingToNextLevel: summary.remainingToNextLevel,
          onShare: () => ScaffoldMessenger.of(context)
            ..hideCurrentSnackBar()
            ..showSnackBar(
              SnackBar(
                content: Text(
                  'Shared ${formatMoney(summary.totalSaved, summary.currency)} '
                  'in savings.',
                ),
              ),
            ),
        ),
        const SizedBox(height: md),
        _MomentumPanel(summary: summary),
        const SizedBox(height: xl),
        Text(
          'Recent savings',
          style: headingMedium.copyWith(color: textPrimary),
        ),
        const SizedBox(height: md),
        for (final transaction in summary.transactions)
          _SavingsEventTile(
            transaction: transaction,
            currency: summary.currency,
          ),
      ],
    );
  }
}

class _MomentumPanel extends StatelessWidget {
  const _MomentumPanel({required this.summary});

  final SavingsSummary summary;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(lg),
      decoration: BoxDecoration(
        color: surfaceWhite,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: borderGrey),
      ),
      child: Row(
        children: [
          Container(
            width: 48,
            height: 48,
            decoration: BoxDecoration(
              color: backgroundLight,
              borderRadius: BorderRadius.circular(100),
            ),
            child: const Icon(Icons.trending_up_rounded, color: primaryNavy),
          ),
          const SizedBox(width: md),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  '${summary.transactions.length} '
                  '${summary.transactions.length == 1 ? 'saving' : 'savings'} '
                  'recorded',
                  style: headingSmall.copyWith(color: textPrimary),
                ),
                const SizedBox(height: xs),
                Text(
                  summary.savingsLevel >= 10
                      ? 'You have reached the top savings level.'
                      : 'Keep tracking deals to reach Level '
                            '${summary.savingsLevel + 1}.',
                  style: bodySmall.copyWith(color: textSecondary),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _SavingsEventTile extends StatelessWidget {
  const _SavingsEventTile({required this.transaction, required this.currency});

  final SavingsTransaction transaction;
  final String currency;

  @override
  Widget build(BuildContext context) {
    final date = DateFormat('MMM yyyy').format(transaction.boughtAt);
    return Container(
      margin: const EdgeInsets.only(bottom: md),
      decoration: BoxDecoration(
        color: surfaceWhite,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: borderGrey),
      ),
      child: ListTile(
        contentPadding: const EdgeInsets.all(md),
        leading: Container(
          width: 44,
          height: 44,
          decoration: BoxDecoration(
            color: successGreen.withValues(alpha: 0.12),
            borderRadius: BorderRadius.circular(100),
          ),
          child: const Icon(Icons.sell_outlined, color: successGreen),
        ),
        title: Text(
          transaction.title,
          maxLines: 2,
          overflow: TextOverflow.ellipsis,
          style: headingSmall.copyWith(color: textPrimary),
        ),
        subtitle: Padding(
          padding: const EdgeInsets.only(top: xs),
          child: Text(
            'Saved ${formatMoney(transaction.saved, currency)} - $date',
            style: bodySmall.copyWith(color: textSecondary),
          ),
        ),
        trailing: const Icon(Icons.chevron_right_rounded),
        onTap: () => context.go('${Routes.product}/${transaction.productId}'),
      ),
    );
  }
}

class _LoadingSavings extends StatelessWidget {
  const _LoadingSavings();

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: const EdgeInsets.all(md),
      children: const [
        MMSkeletonLoader(width: double.infinity, height: 224, borderRadius: 12),
        SizedBox(height: md),
        MMSkeletonLoader(width: double.infinity, height: 104, borderRadius: 12),
        SizedBox(height: xl),
        MMSkeletonLoader(width: 160, height: 24, borderRadius: 8),
        SizedBox(height: md),
        MMSkeletonLoader(width: double.infinity, height: 88, borderRadius: 12),
        SizedBox(height: md),
        MMSkeletonLoader(width: double.infinity, height: 88, borderRadius: 12),
      ],
    );
  }
}

class _EmptySavings extends StatelessWidget {
  const _EmptySavings();

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(lg),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.savings_outlined, size: 48, color: textSecondary),
            const SizedBox(height: md),
            Text(
              'No savings yet',
              style: headingMedium.copyWith(color: textPrimary),
            ),
            const SizedBox(height: sm),
            Text(
              'Add products to your wishlist and buy when prices drop to start '
              'building your savings streak.',
              textAlign: TextAlign.center,
              style: bodyRegular.copyWith(color: textSecondary),
            ),
            const SizedBox(height: lg),
            ElevatedButton.icon(
              style: ElevatedButton.styleFrom(
                backgroundColor: primaryNavy,
                foregroundColor: surfaceWhite,
                padding: const EdgeInsets.symmetric(
                  horizontal: lg,
                  vertical: md,
                ),
              ),
              onPressed: () => context.go(Routes.wishlist),
              icon: const Icon(Icons.favorite_border_rounded),
              label: const Text('Open Wishlist'),
            ),
          ],
        ),
      ),
    );
  }
}
