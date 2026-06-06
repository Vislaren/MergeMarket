import 'package:flutter/material.dart';

import '../theme/colours.dart';
import '../theme/spacing.dart';
import '../theme/typography.dart';
import '../utils/money.dart';

/// AI Deal Meter — a horizontal gauge rating how good the current price is.
///
/// `COMPONENT_LIBRARY.md → MMDealMeter`. The bar runs red → amber → green →
/// gold; a needle marks [score] and the label/colour follow the documented
/// ranges (0–30 poor, 31–60 average, 61–80 good, 81–100 exceptional). The
/// comparison line states how [currentPrice] compares to [averagePrice].
class MMDealMeter extends StatelessWidget {
  const MMDealMeter({
    super.key,
    required this.score,
    required this.currentPrice,
    required this.averagePrice,
    required this.currency,
  });

  final int score;
  final double currentPrice;
  final double averagePrice;
  final String currency;

  /// Clamps [score] into the valid 0–100 gauge range.
  int get _clamped => score.clamp(0, 100);

  /// (label, colour) for the current score band.
  (String, Color) get _band => switch (_clamped) {
        >= 81 => ('Exceptional', dealGold),
        >= 61 => ('Hot Deal', successGreen),
        >= 31 => ('Average', warningAmber),
        _ => ('Poor Deal', accentRed),
      };

  /// Human comparison vs. the 6-month average, e.g. "12% below the 30-day
  /// average. Recommended time to buy.".
  String get _comparison {
    if (averagePrice <= 0 || currentPrice <= 0) {
      return 'Not enough price history yet to compare.';
    }
    final delta = (averagePrice - currentPrice) / averagePrice * 100;
    final pct = delta.abs().round();
    if (delta > 1) {
      return 'Price is $pct% below the 6-month average. Recommended time to buy.';
    }
    if (delta < -1) {
      return 'Price is $pct% above the 6-month average. Consider waiting.';
    }
    return 'Price is in line with the 6-month average.';
  }

  @override
  Widget build(BuildContext context) {
    final (label, colour) = _band;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          crossAxisAlignment: CrossAxisAlignment.baseline,
          textBaseline: TextBaseline.alphabetic,
          children: [
            Text('$_clamped/100',
                style: headingMedium.copyWith(color: colour)),
            const SizedBox(width: sm),
            Text(label, style: headingSmall.copyWith(color: textPrimary)),
          ],
        ),
        const SizedBox(height: sm),
        _Gauge(score: _clamped),
        const SizedBox(height: sm),
        Text(
          _comparison,
          style: bodySmall.copyWith(color: textSecondary),
        ),
        if (averagePrice > 0) ...[
          const SizedBox(height: xs),
          Text(
            '6-month average ${formatMoney(averagePrice, currency)}',
            style: bodySmall.copyWith(color: textSecondary),
          ),
        ],
      ],
    );
  }
}

/// The gradient gauge bar with a needle marker at the score position.
class _Gauge extends StatelessWidget {
  const _Gauge({required this.score});

  final int score;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        const height = 12.0;
        const needle = 16.0;
        final width = constraints.maxWidth;
        // Centre the needle over the score fraction of the bar width.
        final x = (width - needle) * (score / 100);
        return SizedBox(
          height: needle + sm,
          child: Stack(
            clipBehavior: Clip.none,
            children: [
              Align(
                alignment: Alignment.bottomCenter,
                child: Container(
                  height: height,
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(100),
                    gradient: const LinearGradient(
                      colors: [accentRed, warningAmber, successGreen, dealGold],
                    ),
                  ),
                ),
              ),
              Positioned(
                left: x,
                top: 0,
                child: Container(
                  width: needle,
                  height: needle,
                  decoration: BoxDecoration(
                    color: surfaceWhite,
                    shape: BoxShape.circle,
                    border: Border.all(color: primaryNavy, width: 3),
                    boxShadow: const [
                      BoxShadow(
                        color: Color(0x1A1A2B4A),
                        blurRadius: 8,
                        offset: Offset(0, 2),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}
