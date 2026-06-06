import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/product.dart';
import '../models/store_result.dart';
import '../providers/product_providers.dart';
import '../providers/wishlist_providers.dart';
import '../services/api_exception.dart';
import '../theme/colours.dart';
import '../theme/spacing.dart';
import '../theme/typography.dart';
import '../utils/money.dart';
import '../widgets/mm_deal_meter.dart';
import '../widgets/mm_error_state.dart';
import '../widgets/mm_price_chart.dart';
import '../widgets/mm_skeleton_loader.dart';
import '../widgets/mm_store_comparison_table.dart';
import '../widgets/mm_truth_score.dart';

/// Product Detail screen (USER_FLOWS Flows 2, 3, 4, 6).
///
/// Standalone route (no bottom navigation bar). Reached via `/product/:id`,
/// including as the deep-link target of a price-drop notification (Flow 6), so
/// it loads everything it needs from the product id alone: price history, the
/// AI Deal Meter, the multi-store comparison, and the Truth Score. Renders the
/// loading / success / empty / error states from a single
/// [productDetailProvider] source of truth.
class ProductDetailScreen extends ConsumerWidget {
  const ProductDetailScreen({super.key, required this.productId});

  final String productId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final detail = ref.watch(productDetailProvider(productId));

    return Scaffold(
      backgroundColor: backgroundLight,
      appBar: AppBar(
        title: const Text('Product Detail'),
        actions: [
          IconButton(
            icon: const Icon(Icons.favorite_border_rounded),
            tooltip: 'Add to wishlist',
            onPressed: () => _addToWishlist(context, ref),
          ),
          IconButton(
            icon: const Icon(Icons.ios_share_rounded),
            tooltip: 'Share this deal',
            // Outbound share / inbound Share-to-Scrape (Flow 3) is finalised at
            // integration (B-11); confirm intent for now.
            onPressed: () => _snack(context, 'Share link copied (mock).'),
          ),
        ],
      ),
      body: detail.when(
        loading: () => const _LoadingDetail(),
        error: (error, _) => MMErrorState(
          message: error is ApiException
              ? error.message
              : 'Something went wrong. Please try again.',
          icon: error is ApiException && error.kind == ApiErrorKind.notFound
              ? Icons.search_off_rounded
              : Icons.wifi_off_rounded,
          onRetry: () => ref.invalidate(productDetailProvider(productId)),
        ),
        data: (data) => _DetailBody(detail: data),
      ),
    );
  }

  /// Adds this product to the wishlist (USER_FLOWS Flow 4) and reports the
  /// outcome via a SnackBar.
  Future<void> _addToWishlist(BuildContext context, WidgetRef ref) async {
    final messenger = ScaffoldMessenger.of(context);
    try {
      await ref.read(wishlistActionsProvider).add(productId);
      messenger
        ..hideCurrentSnackBar()
        ..showSnackBar(const SnackBar(content: Text('Added to your wishlist.')));
    } on ApiException catch (e) {
      messenger
        ..hideCurrentSnackBar()
        ..showSnackBar(SnackBar(content: Text(e.message)));
    }
  }

  /// Shows a transient confirmation message.
  static void _snack(BuildContext context, String message) {
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(content: Text(message)));
  }
}

/// The populated detail view: hero image, headline price, deal meter, chart,
/// store comparison, and truth score, with a sticky "Go to Best Store" CTA.
class _DetailBody extends StatelessWidget {
  const _DetailBody({required this.detail});

  final ProductDetail detail;

  @override
  Widget build(BuildContext context) {
    final best = detail.bestOffer;
    final currency = best?.currency ?? detail.history.currency;
    final imageUrl = best?.imageUrl ?? '';

    return Column(
      children: [
        Expanded(
          child: SingleChildScrollView(
            padding: const EdgeInsets.only(bottom: lg),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _HeroImage(imageUrl: imageUrl),
                Padding(
                  padding: const EdgeInsets.all(md),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _UpdatedBadge(scrapedAt: best?.scrapedAt ?? ''),
                      const SizedBox(height: sm),
                      Text(detail.history.title,
                          style: headingLarge.copyWith(color: textPrimary)),
                      const SizedBox(height: md),
                      _BestPriceCard(
                        store: best?.store ?? '—',
                        totalCost: best?.totalCost ?? 0,
                        currency: currency,
                        onBuy: () => _snack(
                            context, 'Opening ${best?.store ?? 'store'} (mock).'),
                      ),
                      const SizedBox(height: lg),
                      _Section(
                        title: 'AI Deal Analysis',
                        child: MMDealMeter(
                          score: detail.dealScore,
                          currentPrice: best?.totalCost ?? 0,
                          averagePrice: detail.history.average6m,
                          currency: currency,
                        ),
                      ),
                      const SizedBox(height: lg),
                      _Section(
                        title: 'Price History',
                        child: MMPriceChart(
                          history: detail.history.history,
                          currency: currency,
                        ),
                      ),
                      const SizedBox(height: lg),
                      _Section(
                        title: 'Compare Stores',
                        child: MMStoreComparisonTable(
                          stores: detail.offers
                              .map(StoreResult.fromProduct)
                              .toList(),
                          currency: currency,
                          onGoToStore: (store) => () =>
                              _snack(context, 'Opening ${store.store} (mock).'),
                        ),
                      ),
                      const SizedBox(height: lg),
                      _Section(
                        title: 'Truth Score',
                        child: MMTruthScore(
                          score: detail.truthScore.score,
                          sentiment: detail.truthScore.sentiment,
                          fakeReviewRisk: detail.truthScore.fakeReviewRisk,
                          summary: detail.truthScore.summary,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
        if (best != null) _GoToStoreBar(best: best, currency: currency),
      ],
    );
  }

  void _snack(BuildContext context, String message) =>
      ProductDetailScreen._snack(context, message);
}

/// Hero product image with a shimmer placeholder and a neutral fallback.
class _HeroImage extends StatelessWidget {
  const _HeroImage({required this.imageUrl});

  final String imageUrl;

  @override
  Widget build(BuildContext context) {
    const height = 240.0;
    return CachedNetworkImage(
      imageUrl: imageUrl,
      width: double.infinity,
      height: height,
      fit: BoxFit.cover,
      placeholder: (context, url) =>
          const MMSkeletonLoader(width: double.infinity, height: height),
      errorWidget: (context, url, error) => Container(
        width: double.infinity,
        height: height,
        color: surfaceWhite,
        child: const Icon(Icons.image_not_supported_outlined,
            size: 48, color: borderGrey),
      ),
    );
  }
}

/// "Top deal · Updated Xm ago" pill above the title.
class _UpdatedBadge extends StatelessWidget {
  const _UpdatedBadge({required this.scrapedAt});

  final String scrapedAt;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Container(
          padding: const EdgeInsets.symmetric(horizontal: sm, vertical: xs),
          decoration: BoxDecoration(
            color: dealGold,
            borderRadius: BorderRadius.circular(100),
          ),
          child: Text('TOP DEAL',
              style: labelBold.copyWith(color: textPrimary)),
        ),
        const SizedBox(width: sm),
        Text('Updated ${_relative(scrapedAt)}',
            style: bodySmall.copyWith(color: textSecondary)),
      ],
    );
  }

  /// Coarse "Xm/Xh/Xd ago" string from an ISO 8601 timestamp.
  String _relative(String iso) {
    final t = DateTime.tryParse(iso);
    if (t == null) return 'recently';
    final d = DateTime.now().difference(t);
    if (d.inMinutes < 1) return 'just now';
    if (d.inMinutes < 60) return '${d.inMinutes}m ago';
    if (d.inHours < 24) return '${d.inHours}h ago';
    return '${d.inDays}d ago';
  }
}

/// Navy headline card showing the current best total cost and a Buy Now action.
class _BestPriceCard extends StatelessWidget {
  const _BestPriceCard({
    required this.store,
    required this.totalCost,
    required this.currency,
    required this.onBuy,
  });

  final String store;
  final double totalCost;
  final String currency;
  final VoidCallback onBuy;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(md),
      decoration: BoxDecoration(
        color: primaryNavy,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Current Best Price',
              style: bodySmall.copyWith(color: borderGrey)),
          const SizedBox(height: xs),
          Text(formatMoney(totalCost, currency),
              style: headingLarge.copyWith(color: surfaceWhite)),
          const SizedBox(height: xs),
          Text('on $store', style: bodySmall.copyWith(color: borderGrey)),
          const SizedBox(height: md),
          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: accentRed,
                foregroundColor: surfaceWhite,
                padding: const EdgeInsets.symmetric(vertical: md),
              ),
              onPressed: onBuy,
              child: const Text('Buy Now'),
            ),
          ),
        ],
      ),
    );
  }
}

/// A titled white card section wrapping one piece of detail content.
class _Section extends StatelessWidget {
  const _Section({required this.title, required this.child});

  final String title;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(md),
      decoration: BoxDecoration(
        color: surfaceWhite,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: borderGrey),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(title, style: headingMedium.copyWith(color: textPrimary)),
          const SizedBox(height: md),
          child,
        ],
      ),
    );
  }
}

/// Sticky bottom CTA jumping straight to the cheapest store.
class _GoToStoreBar extends StatelessWidget {
  const _GoToStoreBar({required this.best, required this.currency});

  final Product best;
  final String currency;

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.all(md),
        child: SizedBox(
          width: double.infinity,
          child: ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: primaryNavy,
              foregroundColor: surfaceWhite,
              padding: const EdgeInsets.symmetric(vertical: md),
            ),
            onPressed: () => ProductDetailScreen._snack(
                context, 'Opening ${best.store} (mock).'),
            child: Text(
                'Go to Best Store — ${formatMoney(best.totalCost, currency)}'),
          ),
        ),
      ),
    );
  }
}

/// Skeleton placeholder shown while the detail aggregate loads.
class _LoadingDetail extends StatelessWidget {
  const _LoadingDetail();

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: const [
          MMSkeletonLoader(width: double.infinity, height: 240),
          Padding(
            padding: EdgeInsets.all(md),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                MMSkeletonLoader(width: 160, height: 16),
                SizedBox(height: md),
                MMSkeletonLoader(width: double.infinity, height: 24),
                SizedBox(height: md),
                MMSkeletonLoader(width: double.infinity, height: 110),
                SizedBox(height: lg),
                MMSkeletonLoader(width: double.infinity, height: 140),
                SizedBox(height: lg),
                MMSkeletonLoader(width: double.infinity, height: 180),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
