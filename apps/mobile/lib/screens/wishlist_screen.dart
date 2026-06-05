import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../models/wishlist.dart';
import '../providers/wishlist_providers.dart';
import '../router/app_router.dart';
import '../services/api_exception.dart';
import '../theme/colours.dart';
import '../theme/spacing.dart';
import '../theme/typography.dart';
import '../widgets/mm_error_state.dart';
import '../widgets/mm_set_alert_sheet.dart';
import '../widgets/mm_skeleton_loader.dart';
import '../widgets/mm_wishlist_board.dart';

/// Wishlist screen (USER_FLOWS Flow 4) — the visual board of tracked products.
///
/// A primary bottom-nav destination. Watches [wishlistProvider] and renders the
/// four states: loading (skeleton grid), success ([MMWishlistBoard]), empty
/// (prompt to start searching), and error ([MMErrorState] with retry). Tapping a
/// tile opens Product Detail; the bell starts the Set-Alert flow (Flow 5, wired
/// in B-06); swipe removes the entry.
class WishlistScreen extends ConsumerWidget {
  const WishlistScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final wishlist = ref.watch(wishlistProvider);

    return Scaffold(
      backgroundColor: backgroundLight,
      appBar: AppBar(title: const Text('My Wishlist')),
      body: wishlist.when(
        loading: () => const _LoadingGrid(),
        error: (error, _) => MMErrorState(
          message: error is ApiException
              ? error.message
              : 'Something went wrong. Please try again.',
          onRetry: () => ref.invalidate(wishlistProvider),
        ),
        data: (data) {
          if (data.items.isEmpty) return const _EmptyWishlist();
          return MMWishlistBoard(
            items: data.items,
            onTap: (productId) =>
                context.go('${Routes.product}/$productId'),
            onRemove: (wishlistId) => _remove(context, ref, wishlistId),
            onSetAlert: (productId) => _setAlert(
              context,
              ref,
              data.items.firstWhere((i) => i.productId == productId),
            ),
          );
        },
      ),
    );
  }

  /// Removes a wishlist entry and reports the outcome via a SnackBar.
  Future<void> _remove(
      BuildContext context, WidgetRef ref, String wishlistId) async {
    final messenger = ScaffoldMessenger.of(context);
    try {
      await ref.read(wishlistActionsProvider).remove(wishlistId);
      messenger
        ..hideCurrentSnackBar()
        ..showSnackBar(const SnackBar(content: Text('Removed from wishlist.')));
    } on ApiException catch (e) {
      messenger
        ..hideCurrentSnackBar()
        ..showSnackBar(SnackBar(content: Text(e.message)));
    }
  }

  /// Opens the Set-Alert sheet for a wishlist item (USER_FLOWS Flow 5) and, on
  /// success, takes the user to the Alerts tab to see the new alert.
  Future<void> _setAlert(
      BuildContext context, WidgetRef ref, WishlistItem item) async {
    final current = item.bestTotalCost ?? 0;
    // The wishlist contract carries no currency; the sample locale is XAF (CM).
    final threshold = await showSetAlertSheet(
      context,
      ref,
      productId: item.productId,
      productTitle: item.title,
      currentPrice: current,
      currency: 'XAF',
      imageUrl: item.imageUrl,
    );
    if (threshold != null && context.mounted) {
      context.go(Routes.alerts);
    }
  }
}

/// Skeleton grid shown while the wishlist loads.
class _LoadingGrid extends StatelessWidget {
  const _LoadingGrid();

  @override
  Widget build(BuildContext context) {
    return GridView.count(
      padding: const EdgeInsets.all(md),
      crossAxisCount: 2,
      crossAxisSpacing: md,
      mainAxisSpacing: md,
      childAspectRatio: 0.66,
      children: const [
        MMSkeletonLoader(width: double.infinity, height: 240, borderRadius: 12),
        MMSkeletonLoader(width: double.infinity, height: 240, borderRadius: 12),
        MMSkeletonLoader(width: double.infinity, height: 240, borderRadius: 12),
        MMSkeletonLoader(width: double.infinity, height: 240, borderRadius: 12),
      ],
    );
  }
}

/// Empty-state panel inviting the user to search and start tracking products.
class _EmptyWishlist extends StatelessWidget {
  const _EmptyWishlist();

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(lg),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 96,
              height: 96,
              decoration: const BoxDecoration(
                color: surfaceWhite,
                shape: BoxShape.circle,
              ),
              child: const Icon(Icons.favorite_border_rounded,
                  size: 40, color: textSecondary),
            ),
            const SizedBox(height: lg),
            Text('Your wishlist is empty',
                style: headingMedium.copyWith(color: textPrimary)),
            const SizedBox(height: sm),
            Text(
              'Save items you want to track. We\'ll monitor prices across '
              'multiple stores and alert you when they drop.',
              textAlign: TextAlign.center,
              style: bodyRegular.copyWith(color: textSecondary),
            ),
            const SizedBox(height: lg),
            ElevatedButton.icon(
              style: ElevatedButton.styleFrom(
                backgroundColor: primaryNavy,
                foregroundColor: surfaceWhite,
                padding:
                    const EdgeInsets.symmetric(horizontal: lg, vertical: md),
              ),
              onPressed: () => context.go(Routes.home),
              icon: const Icon(Icons.search_rounded),
              label: const Text('Start Searching'),
            ),
          ],
        ),
      ),
    );
  }
}
