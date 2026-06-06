import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../models/product.dart';
import '../providers/search_providers.dart';
import '../router/app_router.dart';
import '../services/api_exception.dart';
import '../theme/colours.dart';
import '../theme/spacing.dart';
import '../theme/typography.dart';
import '../widgets/mm_error_state.dart';
import '../widgets/mm_product_card.dart';
import '../widgets/mm_skeleton_loader.dart';

/// Results screen (USER_FLOWS Flows 2 & 3) — the concurrent multi-store list.
///
/// Standalone route (no bottom navigation bar) per `COMPONENT_LIBRARY.md
/// §Navigation`. Watches [searchResultsProvider] and renders the four required
/// states: loading (skeletons), success (sorted [MMProductCard] list), empty,
/// and error ([MMErrorState] with retry). Each card sorts by total cost so the
/// cheapest landed-cost offer is first.
class ResultsScreen extends ConsumerWidget {
  const ResultsScreen({super.key, this.query});

  /// The search query that produced these results (from `?q=` on the route).
  final String? query;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final query = this.query?.trim() ?? '';

    if (query.isEmpty) {
      return Scaffold(
        appBar: AppBar(title: const Text('Results')),
        body: const _CentredMessage(
          icon: Icons.search_rounded,
          message: 'Search for a product to see store comparisons.',
        ),
      );
    }

    final results = ref.watch(searchResultsProvider(query));

    return Scaffold(
      backgroundColor: backgroundLight,
      appBar: AppBar(title: Text(query)),
      body: results.when(
        loading: () => const _LoadingList(),
        error: (error, _) => MMErrorState(
          message: error is ApiException
              ? error.message
              : 'Something went wrong. Please try again.',
          icon: error is ApiException && error.kind == ApiErrorKind.timeout
              ? Icons.hourglass_empty_rounded
              : Icons.wifi_off_rounded,
          onRetry: () => ref.invalidate(searchResultsProvider(query)),
        ),
        data: (response) {
          if (response.results.isEmpty) {
            return _CentredMessage(
              icon: Icons.search_off_rounded,
              message: 'No results found for "$query".',
            );
          }
          final sort = ref.watch(resultSortProvider);
          final sorted = sortResults(response.results, sort);
          return Column(
            children: [
              _ResultsHeader(response: response),
              _SortBar(active: sort, onChanged: (m) {
                ref.read(resultSortProvider.notifier).select(m);
              }),
              Expanded(
                child: ListView.builder(
                  padding: const EdgeInsets.fromLTRB(md, sm, md, md),
                  itemCount: sorted.length,
                  itemBuilder: (context, index) {
                    final product = sorted[index];
                    return MMProductCard(
                      productId: product.productId,
                      title: product.title,
                      imageUrl: product.imageUrl,
                      price: product.price,
                      shipping: product.shipping,
                      totalCost: product.totalCost,
                      store: product.store,
                      currency: product.currency,
                      dealScore: product.dealScore,
                      onTap: () =>
                          context.go('${Routes.product}/${product.productId}'),
                    );
                  },
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}

/// Summary line: offer count, distinct store count, latency, and cache state.
class _ResultsHeader extends StatelessWidget {
  const _ResultsHeader({required this.response});

  final SearchResponse response;

  @override
  Widget build(BuildContext context) {
    final seconds = (response.latencyMs / 1000).toStringAsFixed(
        response.latencyMs >= 1000 ? 1 : 2);
    return Padding(
      padding: const EdgeInsets.fromLTRB(md, md, md, 0),
      child: Row(
        children: [
          Expanded(
            child: Text(
              '${response.results.length} offers from '
              '${response.storeCount} stores · ${seconds}s',
              style: bodySmall.copyWith(color: textSecondary),
            ),
          ),
          if (response.cached)
            Text('cached', style: labelBold.copyWith(color: successGreen)),
        ],
      ),
    );
  }
}

/// Segmented sort control mapping to the design's Best Price / Top Rated /
/// Fastest Ship options.
class _SortBar extends StatelessWidget {
  const _SortBar({required this.active, required this.onChanged});

  final ResultSort active;
  final ValueChanged<ResultSort> onChanged;

  static const _labels = {
    ResultSort.bestPrice: 'Best Price',
    ResultSort.topRated: 'Top Rated',
    ResultSort.fastestShip: 'Fastest Ship',
  };

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      padding: const EdgeInsets.symmetric(horizontal: md, vertical: sm),
      child: Row(
        children: [
          for (final entry in _labels.entries)
            Padding(
              padding: const EdgeInsets.only(right: sm),
              child: ChoiceChip(
                label: Text(entry.value),
                selected: active == entry.key,
                onSelected: (_) => onChanged(entry.key),
              ),
            ),
        ],
      ),
    );
  }
}

/// Skeleton placeholder list shown while the search request is in flight.
class _LoadingList extends StatelessWidget {
  const _LoadingList();

  @override
  Widget build(BuildContext context) {
    return ListView.builder(
      padding: const EdgeInsets.all(md),
      itemCount: 5,
      itemBuilder: (context, _) => Padding(
        padding: const EdgeInsets.only(bottom: md),
        child: Row(
          children: const [
            MMSkeletonLoader(width: 72, height: 72),
            SizedBox(width: md),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  MMSkeletonLoader(width: 120, height: 12),
                  SizedBox(height: sm),
                  MMSkeletonLoader(width: double.infinity, height: 16),
                  SizedBox(height: sm),
                  MMSkeletonLoader(width: 90, height: 20),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// Centred icon + message used for the empty and no-query states.
class _CentredMessage extends StatelessWidget {
  const _CentredMessage({required this.icon, required this.message});

  final IconData icon;
  final String message;

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
          ],
        ),
      ),
    );
  }
}
