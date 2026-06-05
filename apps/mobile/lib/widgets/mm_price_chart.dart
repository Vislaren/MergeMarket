import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../models/price_history.dart';
import '../theme/colours.dart';
import '../theme/typography.dart';
import '../utils/money.dart';

/// Price history line chart for the Product Detail screen.
///
/// `COMPONENT_LIBRARY.md → MMPriceChart`. Navy curved line over a light
/// background, the latest point dotted in [accentRed], and a touch tooltip
/// showing the exact price and date. Renders the most recent [monthsBack]
/// points of [history] (oldest first).
class MMPriceChart extends StatelessWidget {
  const MMPriceChart({
    super.key,
    required this.history,
    required this.currency,
    this.monthsBack = 6,
  });

  final List<PricePoint> history;
  final String currency;
  final int monthsBack;

  @override
  Widget build(BuildContext context) {
    // Keep only the most recent [monthsBack] points, preserving oldest-first.
    final points = history.length > monthsBack
        ? history.sublist(history.length - monthsBack)
        : history;

    if (points.length < 2) {
      return SizedBox(
        height: 180,
        child: Center(
          child: Text(
            'Not enough price history to chart yet.',
            style: bodySmall.copyWith(color: textSecondary),
          ),
        ),
      );
    }

    final spots = <FlSpot>[
      for (var i = 0; i < points.length; i++)
        FlSpot(i.toDouble(), points[i].price),
    ];

    final prices = points.map((p) => p.price).toList();
    final minPrice = prices.reduce((a, b) => a < b ? a : b);
    final maxPrice = prices.reduce((a, b) => a > b ? a : b);
    // Pad the vertical range by ~8% so the line never touches the edges.
    final pad = ((maxPrice - minPrice) * 0.08).clamp(1, double.infinity);

    return SizedBox(
      height: 180,
      child: LineChart(
        LineChartData(
          minX: 0,
          maxX: (points.length - 1).toDouble(),
          minY: minPrice - pad,
          maxY: maxPrice + pad,
          gridData: FlGridData(
            show: true,
            drawVerticalLine: false,
            horizontalInterval: ((maxPrice - minPrice) / 3).clamp(1, double.infinity),
            getDrawingHorizontalLine: (_) =>
                const FlLine(color: borderGrey, strokeWidth: 1),
          ),
          borderData: FlBorderData(show: false),
          titlesData: FlTitlesData(
            topTitles:
                const AxisTitles(sideTitles: SideTitles(showTitles: false)),
            rightTitles:
                const AxisTitles(sideTitles: SideTitles(showTitles: false)),
            leftTitles: AxisTitles(
              sideTitles: SideTitles(
                showTitles: true,
                reservedSize: 44,
                getTitlesWidget: (value, meta) {
                  if (value == meta.min || value == meta.max) {
                    return const SizedBox.shrink();
                  }
                  return Text(
                    NumberFormat.compact().format(value),
                    style: bodySmall.copyWith(color: textSecondary),
                  );
                },
              ),
            ),
            bottomTitles: AxisTitles(
              sideTitles: SideTitles(
                showTitles: true,
                reservedSize: 24,
                interval: 1,
                getTitlesWidget: (value, meta) {
                  final i = value.round();
                  if (i < 0 || i >= points.length) {
                    return const SizedBox.shrink();
                  }
                  return Padding(
                    padding: const EdgeInsets.only(top: 4),
                    child: Text(
                      DateFormat.MMM().format(points[i].recordedAtDate),
                      style: bodySmall.copyWith(color: textSecondary),
                    ),
                  );
                },
              ),
            ),
          ),
          lineTouchData: LineTouchData(
            touchTooltipData: LineTouchTooltipData(
              getTooltipColor: (_) => primaryNavy,
              getTooltipItems: (touched) => [
                for (final spot in touched)
                  LineTooltipItem(
                    '${formatMoney(spot.y, currency)}\n'
                    '${DateFormat.yMMMd().format(points[spot.x.round()].recordedAtDate)}',
                    bodySmall.copyWith(color: surfaceWhite),
                  ),
              ],
            ),
          ),
          lineBarsData: [
            LineChartBarData(
              spots: spots,
              isCurved: true,
              color: primaryNavy,
              barWidth: 3,
              dotData: FlDotData(
                show: true,
                getDotPainter: (spot, percent, bar, index) {
                  final isLast = index == spots.length - 1;
                  return FlDotCirclePainter(
                    radius: isLast ? 5 : 3,
                    color: isLast ? accentRed : primaryNavy,
                    strokeWidth: 0,
                  );
                },
              ),
              belowBarData: BarAreaData(
                show: true,
                color: primaryNavy.withValues(alpha: 0.06),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
