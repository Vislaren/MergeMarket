import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/alerts_providers.dart';
import '../providers/notification_providers.dart';
import '../router/app_router.dart';
import '../services/api_exception.dart';
import '../services/notification_service.dart';
import '../theme/colours.dart';
import '../theme/spacing.dart';
import '../theme/typography.dart';
import '../widgets/mm_alert_card.dart';
import '../widgets/mm_error_state.dart';
import '../widgets/mm_skeleton_loader.dart';

/// Alerts screen (USER_FLOWS Flow 5) — the list of active price alerts.
///
/// A primary bottom-nav destination. Watches [alertsProvider] and renders the
/// four states: loading (skeletons), success (a list of [MMAlertCard]), empty
/// (prompt to set alerts from the wishlist), and error ([MMErrorState] with
/// retry). Swiping a card deletes the alert.
class AlertsScreen extends ConsumerWidget {
  const AlertsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final alerts = ref.watch(alertsProvider);

    // Surface a foreground price-drop / restock push as an in-app banner while
    // the user is on the Alerts screen (USER_FLOWS Flow 6). Tapping it opens the
    // product. With the default NoopPushBackend this never fires.
    ref.listen(incomingNotificationsProvider, (_, next) {
      final notification = next.value;
      if (notification == null) return;
      final messenger = ScaffoldMessenger.of(context);
      final route = NotificationService.routeFor(notification);
      messenger
        ..hideCurrentSnackBar()
        ..showSnackBar(SnackBar(
          content: Text(notification.body.isNotEmpty
              ? notification.body
              : notification.title),
          action: route == null
              ? null
              : SnackBarAction(
                  label: 'View',
                  onPressed: () => context.go(route),
                ),
        ));
    });

    return Scaffold(
      backgroundColor: backgroundLight,
      appBar: AppBar(title: const Text('Price Alerts')),
      body: alerts.when(
        loading: () => const _LoadingList(),
        error: (error, _) => MMErrorState(
          message: error is ApiException
              ? error.message
              : 'Something went wrong. Please try again.',
          onRetry: () => ref.invalidate(alertsProvider),
        ),
        data: (data) {
          if (data.alerts.isEmpty) return const _EmptyAlerts();
          return Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Padding(
                padding: const EdgeInsets.fromLTRB(md, md, md, 0),
                child: Text(
                  'Tracking ${data.alerts.length} '
                  '${data.alerts.length == 1 ? 'item' : 'items'} for price drops.',
                  style: bodyRegular.copyWith(color: textSecondary),
                ),
              ),
              Expanded(
                child: ListView.builder(
                  padding: const EdgeInsets.all(md),
                  itemCount: data.alerts.length,
                  itemBuilder: (context, index) {
                    final alert = data.alerts[index];
                    return MMAlertCard(
                      alertId: alert.alertId,
                      productTitle: alert.title,
                      thresholdPrice: alert.thresholdPrice,
                      currency: alert.currency,
                      isActive: alert.isActive,
                      onDelete: () => _delete(context, ref, alert.alertId),
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

  /// Deletes an alert and reports the outcome via a SnackBar.
  Future<void> _delete(
      BuildContext context, WidgetRef ref, String alertId) async {
    final messenger = ScaffoldMessenger.of(context);
    try {
      await ref.read(alertsActionsProvider).remove(alertId);
      messenger
        ..hideCurrentSnackBar()
        ..showSnackBar(const SnackBar(content: Text('Alert removed.')));
    } on ApiException catch (e) {
      messenger
        ..hideCurrentSnackBar()
        ..showSnackBar(SnackBar(content: Text(e.message)));
    }
  }
}

/// Skeleton list shown while alerts load.
class _LoadingList extends StatelessWidget {
  const _LoadingList();

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: const EdgeInsets.all(md),
      children: const [
        MMSkeletonLoader(width: double.infinity, height: 72, borderRadius: 12),
        SizedBox(height: md),
        MMSkeletonLoader(width: double.infinity, height: 72, borderRadius: 12),
        SizedBox(height: md),
        MMSkeletonLoader(width: double.infinity, height: 72, borderRadius: 12),
      ],
    );
  }
}

/// Empty-state panel pointing the user to the wishlist to create alerts.
class _EmptyAlerts extends StatelessWidget {
  const _EmptyAlerts();

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(lg),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.notifications_off_outlined,
                size: 48, color: textSecondary),
            const SizedBox(height: md),
            Text('No price alerts yet',
                style: headingMedium.copyWith(color: textPrimary)),
            const SizedBox(height: sm),
            Text(
              'Open a product in your wishlist and tap the bell to get notified '
              'when its price drops.',
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
              onPressed: () => context.go(Routes.wishlist),
              icon: const Icon(Icons.favorite_border_rounded),
              label: const Text('Go to Wishlist'),
            ),
          ],
        ),
      ),
    );
  }
}
