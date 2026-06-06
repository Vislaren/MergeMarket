import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';

import '../models/wishlist.dart';
import '../theme/colours.dart';
import '../theme/spacing.dart';
import '../theme/typography.dart';
import '../utils/money.dart';
import 'mm_skeleton_loader.dart';

/// Visual wishlist board — a two-column grid of tracked products.
///
/// `COMPONENT_LIBRARY.md → MMWishlistBoard`. Each tile shows the image, title,
/// store-tracking count, and best price, with a bell to set a price alert and a
/// swipe-to-remove gesture (USER_FLOWS Flow 4). [onTap] opens Product Detail,
/// [onSetAlert] starts the Set-Alert flow (Flow 5), [onRemove] deletes the entry.
class MMWishlistBoard extends StatelessWidget {
  const MMWishlistBoard({
    super.key,
    required this.items,
    required this.onTap,
    required this.onRemove,
    required this.onSetAlert,
  });

  final List<WishlistItem> items;
  final void Function(String productId) onTap;
  final void Function(String wishlistId) onRemove;
  final void Function(String productId) onSetAlert;

  @override
  Widget build(BuildContext context) {
    return GridView.builder(
      padding: const EdgeInsets.all(md),
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 2,
        crossAxisSpacing: md,
        mainAxisSpacing: md,
        childAspectRatio: 0.66,
      ),
      itemCount: items.length,
      itemBuilder: (context, index) {
        final item = items[index];
        return Dismissible(
          key: ValueKey(item.wishlistId),
          direction: DismissDirection.endToStart,
          onDismissed: (_) => onRemove(item.wishlistId),
          background: Container(
            alignment: Alignment.centerRight,
            padding: const EdgeInsets.only(right: md),
            decoration: BoxDecoration(
              color: accentRed,
              borderRadius: BorderRadius.circular(12),
            ),
            child: const Icon(Icons.delete_outline_rounded,
                color: surfaceWhite),
          ),
          child: _WishlistTile(
            item: item,
            onTap: () => onTap(item.productId),
            onSetAlert: () => onSetAlert(item.productId),
          ),
        );
      },
    );
  }
}

/// A single wishlist product tile.
class _WishlistTile extends StatelessWidget {
  const _WishlistTile({
    required this.item,
    required this.onTap,
    required this.onSetAlert,
  });

  final WishlistItem item;
  final VoidCallback onTap;
  final VoidCallback onSetAlert;

  @override
  Widget build(BuildContext context) {
    final best = item.bestTotalCost ?? 0;
    final currency = item.stores.isEmpty ? '' : 'XAF';
    return Card(
      margin: EdgeInsets.zero,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: const BorderSide(color: borderGrey),
      ),
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: onTap,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            ClipRRect(
              borderRadius:
                  const BorderRadius.vertical(top: Radius.circular(12)),
              child: CachedNetworkImage(
                imageUrl: item.imageUrl,
                height: 110,
                width: double.infinity,
                fit: BoxFit.cover,
                placeholder: (context, url) => const MMSkeletonLoader(
                    width: double.infinity, height: 110),
                errorWidget: (context, url, error) => Container(
                  height: 110,
                  color: backgroundLight,
                  child: const Icon(Icons.image_not_supported_outlined,
                      color: borderGrey),
                ),
              ),
            ),
            Expanded(
              child: Padding(
                padding: const EdgeInsets.all(sm),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      item.title,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: headingSmall.copyWith(color: textPrimary),
                    ),
                    const SizedBox(height: xs),
                    Text(
                      '${item.storeCount} '
                      '${item.storeCount == 1 ? 'store' : 'stores'} tracking',
                      style: bodySmall.copyWith(color: textSecondary),
                    ),
                    const Spacer(),
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text('Best price',
                                  style: bodySmall.copyWith(
                                      color: textSecondary)),
                              Text(
                                formatMoney(best, currency),
                                style: headingSmall.copyWith(color: accentRed),
                              ),
                            ],
                          ),
                        ),
                        IconButton(
                          visualDensity: VisualDensity.compact,
                          padding: EdgeInsets.zero,
                          constraints: const BoxConstraints(),
                          icon: const Icon(Icons.notifications_none_rounded,
                              color: primaryNavy),
                          tooltip: 'Set price alert',
                          onPressed: onSetAlert,
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
