import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'providers/notification_providers.dart';
import 'router/app_router.dart';
import 'services/notification_service.dart';
import 'theme/app_theme.dart';

void main() {
  runApp(const ProviderScope(child: MergeMarketApp()));
}

/// Root widget for the MergeMarket client.
///
/// Wraps the app in Riverpod's [ProviderScope] (in [main]) and wires
/// `go_router` + the Material 3 design-token theme. Business logic and screen
/// content live in their own providers/screens — this widget only composes the
/// router and theme.
class MergeMarketApp extends ConsumerWidget {
  const MergeMarketApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);

    // Deep-link a tapped price-drop / restock notification straight to the
    // relevant Product Detail screen (USER_FLOWS Flow 6). With the default
    // NoopPushBackend this never fires; a real backend (configured at release)
    // drives it.
    ref.listen(notificationTapsProvider, (_, next) {
      final notification = next.value;
      if (notification == null) return;
      final route = NotificationService.routeFor(notification);
      if (route != null) router.go(route);
    });

    return MaterialApp.router(
      title: 'MergeMarket',
      debugShowCheckedModeBanner: false,
      theme: buildAppTheme(),
      routerConfig: router,
    );
  }
}
