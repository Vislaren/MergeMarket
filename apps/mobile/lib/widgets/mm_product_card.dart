import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';

import '../theme/colours.dart';
import '../theme/spacing.dart';
import '../theme/typography.dart';
import '../utils/money.dart';
import 'mm_skeleton_loader.dart';

/// A single store offer card used on the Results and Wishlist screens.
///
/// `COMPONENT_LIBRARY.md → MMProductCard`. Image left, details right, deal
/// badge derived from [dealScore], and the total cost (price + shipping) bold
/// in [accentRed] at the bottom-right — the figure the user actually compares.
class MMProductCard extends StatelessWidget {
  const MMProductCard({
    super.key,
    required this.productId,
    required this.title,
    required this.imageUrl,
    required this.price,
    required this.shipping,
    required this.totalCost,
    required this.store,
    required this.currency,
    required this.dealScore,
    required this.onTap,
    this.showWishlistButton = true,
  });

  final String productId;
  final String title;
  final String imageUrl;
  final double price;
  final double shipping;
  final double totalCost;
  final String store;
  final String currency;
  final int dealScore;
  final VoidCallback onTap;
  final bool showWishlistButton;

  @override
  Widget build(BuildContext context) {
    final shippingLabel =
        shipping <= 0 ? 'Free shipping' : '+ ${formatMoney(shipping, '')} shipping';

    return Card(
      margin: const EdgeInsets.only(bottom: md),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: const BorderSide(color: borderGrey),
      ),
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.all(md),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _Thumbnail(imageUrl: imageUrl),
              const SizedBox(width: md),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Expanded(
                          child: Text(
                            store,
                            style: bodySmall.copyWith(color: textSecondary),
                          ),
                        ),
                        if (showWishlistButton)
                          _WishlistButton(title: title),
                      ],
                    ),
                    const SizedBox(height: xs),
                    Text(
                      title,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: headingSmall.copyWith(color: textPrimary),
                    ),
                    const SizedBox(height: sm),
                    _DealBadge(score: dealScore),
                    const Divider(height: lg, color: borderGrey),
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Expanded(
                          child: Text(
                            'Base: ${formatMoney(price, currency)}\n$shippingLabel',
                            style: bodySmall.copyWith(color: textSecondary),
                          ),
                        ),
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.end,
                          children: [
                            Text('Total',
                                style: bodySmall.copyWith(color: textSecondary)),
                            Text(
                              formatMoney(totalCost, currency),
                              style: headingMedium.copyWith(color: accentRed),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

/// Product thumbnail with a shimmer placeholder while loading and a neutral
/// fallback when the image can't be fetched (the mock URLs don't resolve).
class _Thumbnail extends StatelessWidget {
  const _Thumbnail({required this.imageUrl});

  final String imageUrl;

  @override
  Widget build(BuildContext context) {
    const size = 72.0;
    return ClipRRect(
      borderRadius: BorderRadius.circular(sm),
      child: CachedNetworkImage(
        imageUrl: imageUrl,
        width: size,
        height: size,
        fit: BoxFit.cover,
        placeholder: (context, url) =>
            const MMSkeletonLoader(width: size, height: size, borderRadius: sm),
        errorWidget: (context, url, error) => Container(
          width: size,
          height: size,
          color: backgroundLight,
          child: const Icon(Icons.image_not_supported_outlined,
              color: borderGrey),
        ),
      ),
    );
  }
}

/// Deal-quality pill, coloured by the AI Deal Meter ranges in
/// `COMPONENT_LIBRARY.md → MMDealMeter`.
class _DealBadge extends StatelessWidget {
  const _DealBadge({required this.score});

  final int score;

  @override
  Widget build(BuildContext context) {
    final (label, background, foreground) = switch (score) {
      >= 81 => ('Hot Deal', dealGold, textPrimary),
      >= 61 => ('Good Value', successGreen, surfaceWhite),
      >= 31 => ('Average', warningAmber, surfaceWhite),
      _ => ('Low Value', accentRed, surfaceWhite),
    };
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: sm, vertical: xs),
      decoration: BoxDecoration(
        color: background,
        borderRadius: BorderRadius.circular(100),
      ),
      child: Text(label, style: labelBold.copyWith(color: foreground)),
    );
  }
}

/// Wishlist heart. Wishlist mutation lands in B-05, so for now it confirms the
/// intent with a SnackBar rather than silently doing nothing.
class _WishlistButton extends StatelessWidget {
  const _WishlistButton({required this.title});

  final String title;

  @override
  Widget build(BuildContext context) {
    return IconButton(
      visualDensity: VisualDensity.compact,
      padding: EdgeInsets.zero,
      constraints: const BoxConstraints(),
      icon: const Icon(Icons.favorite_border_rounded, color: textSecondary),
      tooltip: 'Add to wishlist',
      onPressed: () {
        ScaffoldMessenger.of(context)
          ..hideCurrentSnackBar()
          ..showSnackBar(
            SnackBar(content: Text('Wishlist coming soon — "$title"')),
          );
      },
    );
  }
}
