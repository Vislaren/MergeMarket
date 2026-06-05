import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/alerts_providers.dart';
import '../services/api_exception.dart';
import '../theme/colours.dart';
import '../theme/spacing.dart';
import '../theme/typography.dart';
import '../utils/money.dart';
import 'mm_skeleton_loader.dart';

/// Opens the Set Price Alert bottom sheet (USER_FLOWS Flow 5), and — when the
/// user confirms — creates the alert via [alertsActionsProvider].
///
/// Returns the chosen threshold price on success, or `null` if the user
/// cancelled or the create failed (a SnackBar reports any failure). Callers
/// typically navigate to the Alerts screen on a non-null result.
Future<double?> showSetAlertSheet(
  BuildContext context,
  WidgetRef ref, {
  required String productId,
  required String productTitle,
  required double currentPrice,
  double averagePrice = 0,
  required String currency,
  String imageUrl = '',
}) async {
  final threshold = await showModalBottomSheet<double>(
    context: context,
    isScrollControlled: true,
    backgroundColor: surfaceWhite,
    shape: const RoundedRectangleBorder(
      borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
    ),
    builder: (_) => MMSetAlertSheet(
      productTitle: productTitle,
      currentPrice: currentPrice,
      averagePrice: averagePrice,
      currency: currency,
      imageUrl: imageUrl,
    ),
  );

  if (threshold == null) return null;
  if (!context.mounted) return null;

  final messenger = ScaffoldMessenger.of(context);
  try {
    await ref.read(alertsActionsProvider).create(
          productId: productId,
          thresholdPrice: threshold,
          currency: currency,
        );
    messenger
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(
          content: Text('Alert set for ${formatMoney(threshold, currency)}.')));
    return threshold;
  } on ApiException catch (e) {
    messenger
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(content: Text(e.message)));
    return null;
  }
}

/// The Set Price Alert sheet body.
///
/// `COMPONENT_LIBRARY.md` Set-Alert flow / the `set_alert_screen` design. Shows
/// the product + current price, a "drops to" amount synced between a numeric
/// field and a slider (min = half the current price, max = current price), an
/// average-price hint, and Set Alert / Cancel actions. On confirm it pops the
/// chosen threshold; on cancel it pops `null`.
class MMSetAlertSheet extends StatefulWidget {
  const MMSetAlertSheet({
    super.key,
    required this.productTitle,
    required this.currentPrice,
    required this.averagePrice,
    required this.currency,
    this.imageUrl = '',
  });

  final String productTitle;
  final double currentPrice;
  final double averagePrice;
  final String currency;
  final String imageUrl;

  @override
  State<MMSetAlertSheet> createState() => _MMSetAlertSheetState();
}

class _MMSetAlertSheetState extends State<MMSetAlertSheet> {
  late final double _min;
  late final double _max;
  late double _threshold;
  late final TextEditingController _controller;

  @override
  void initState() {
    super.initState();
    _max = widget.currentPrice <= 0 ? 1 : widget.currentPrice;
    _min = (_max * 0.5).roundToDouble();
    // Default to just below the current price (or the average, when lower).
    final suggested = widget.averagePrice > 0 && widget.averagePrice < _max
        ? widget.averagePrice
        : _max * 0.9;
    _threshold = suggested.clamp(_min, _max).roundToDouble();
    _controller = TextEditingController(text: _threshold.round().toString());
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  void _setThreshold(double value) {
    final clamped = value.clamp(_min, _max);
    setState(() {
      _threshold = clamped;
      _controller.text = clamped.round().toString();
      _controller.selection = TextSelection.fromPosition(
          TextPosition(offset: _controller.text.length));
    });
  }

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: EdgeInsets.only(
        left: lg,
        right: lg,
        top: md,
        bottom: MediaQuery.of(context).viewInsets.bottom + lg,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Center(
            child: Container(
              width: 40,
              height: 4,
              decoration: BoxDecoration(
                color: borderGrey,
                borderRadius: BorderRadius.circular(100),
              ),
            ),
          ),
          const SizedBox(height: md),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text('Set Price Alert',
                  style: headingLarge.copyWith(color: textPrimary)),
              IconButton(
                icon: const Icon(Icons.close_rounded),
                onPressed: () => Navigator.of(context).pop(),
              ),
            ],
          ),
          const SizedBox(height: sm),
          _ProductRow(
            title: widget.productTitle,
            currentPrice: widget.currentPrice,
            currency: widget.currency,
            imageUrl: widget.imageUrl,
          ),
          const SizedBox(height: lg),
          Center(
            child: Text('ALERT ME WHEN PRICE DROPS TO',
                style: labelBold.copyWith(color: textSecondary)),
          ),
          const SizedBox(height: sm),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            crossAxisAlignment: CrossAxisAlignment.center,
            children: [
              Text(widget.currency,
                  style: headingMedium.copyWith(color: primaryNavy)),
              const SizedBox(width: sm),
              SizedBox(
                width: 140,
                child: TextField(
                  controller: _controller,
                  textAlign: TextAlign.center,
                  keyboardType: TextInputType.number,
                  inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                  style: headingLarge.copyWith(color: textPrimary),
                  decoration: const InputDecoration(
                    isDense: true,
                    border: UnderlineInputBorder(),
                  ),
                  onChanged: (text) {
                    final value = double.tryParse(text);
                    if (value != null) {
                      setState(() => _threshold = value.clamp(_min, _max));
                    }
                  },
                ),
              ),
            ],
          ),
          const SizedBox(height: md),
          Slider(
            value: _threshold.clamp(_min, _max),
            min: _min,
            max: _max,
            activeColor: primaryNavy,
            onChanged: _setThreshold,
          ),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text('${formatMoney(_min, widget.currency)} (Min)',
                  style: bodySmall.copyWith(color: textSecondary)),
              Text('${formatMoney(_max, widget.currency)} (Current)',
                  style: bodySmall.copyWith(color: textSecondary)),
            ],
          ),
          if (widget.averagePrice > 0) ...[
            const SizedBox(height: md),
            _InfoBox(
              text: 'This item averages '
                  '${formatMoney(widget.averagePrice, widget.currency)}. '
                  'An alert set at ${formatMoney(_threshold, widget.currency)} '
                  'gives you a strong chance of catching a deal within the next '
                  '30 days based on market trends.',
            ),
          ],
          const SizedBox(height: lg),
          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: primaryNavy,
                foregroundColor: surfaceWhite,
                padding: const EdgeInsets.symmetric(vertical: md),
              ),
              onPressed: () =>
                  Navigator.of(context).pop(_threshold.roundToDouble()),
              child: const Text('Set Alert'),
            ),
          ),
          const SizedBox(height: sm),
          Center(
            child: TextButton(
              onPressed: () => Navigator.of(context).pop(),
              child: Text('Cancel',
                  style: headingSmall.copyWith(color: primaryNavy)),
            ),
          ),
        ],
      ),
    );
  }
}

/// Product summary row at the top of the sheet.
class _ProductRow extends StatelessWidget {
  const _ProductRow({
    required this.title,
    required this.currentPrice,
    required this.currency,
    required this.imageUrl,
  });

  final String title;
  final double currentPrice;
  final String currency;
  final String imageUrl;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(md),
      decoration: BoxDecoration(
        color: backgroundLight,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: borderGrey),
      ),
      child: Row(
        children: [
          ClipRRect(
            borderRadius: BorderRadius.circular(sm),
            child: CachedNetworkImage(
              imageUrl: imageUrl,
              width: 56,
              height: 56,
              fit: BoxFit.cover,
              placeholder: (context, url) =>
                  const MMSkeletonLoader(width: 56, height: 56, borderRadius: sm),
              errorWidget: (context, url, error) => Container(
                width: 56,
                height: 56,
                color: surfaceWhite,
                child: const Icon(Icons.image_not_supported_outlined,
                    color: borderGrey),
              ),
            ),
          ),
          const SizedBox(width: md),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(title,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: headingSmall.copyWith(color: textPrimary)),
                const SizedBox(height: xs),
                Text('Current: ${formatMoney(currentPrice, currency)}',
                    style: bodySmall.copyWith(color: textSecondary)),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

/// Informational hint box with a leading info icon.
class _InfoBox extends StatelessWidget {
  const _InfoBox({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(md),
      decoration: BoxDecoration(
        color: backgroundLight,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: borderGrey),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Icon(Icons.info_outline_rounded, color: primaryNavy, size: 20),
          const SizedBox(width: sm),
          Expanded(
            child: Text(text,
                style: bodySmall.copyWith(color: textSecondary)),
          ),
        ],
      ),
    );
  }
}
