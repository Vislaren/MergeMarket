import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'router/app_router.dart';
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
    return MaterialApp.router(
      title: 'MergeMarket',
      debugShowCheckedModeBanner: false,
      theme: buildAppTheme(),
      routerConfig: router,
    );
  }
}
