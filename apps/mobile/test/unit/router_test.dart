// Unit/widget tests for B-01 navigation setup (go_router + Riverpod).
//
// Source of truth: project_docs/ui/COMPONENT_LIBRARY.md (Navigation Structure)
// and project_docs/flows/USER_FLOWS.md.
// Test artefacts: docs/testing/session-03/unit/.

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';

import 'package:mergemarket/main.dart';
import 'package:mergemarket/router/app_router.dart';

void main() {
  group('B-01 Router Unit Tests', () {
    test('TC-03-U-005: route path constants match COMPONENT_LIBRARY', () {
      expect(Routes.home, '/');
      expect(Routes.results, '/results');
      expect(Routes.product, '/product');
      expect(Routes.wishlist, '/wishlist');
      expect(Routes.alerts, '/alerts');
      expect(Routes.savings, '/savings');
      expect(Routes.login, '/login');
      expect(Routes.register, '/register');
    });

    test('TC-03-U-006: routerProvider builds a GoRouter with all routes', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);
      final router = container.read(routerProvider);
      expect(router, isA<GoRouter>());

      // One StatefulShellRoute (the 4 bottom-nav tabs) + four top-level routes
      // (results, product/:id, login, register).
      expect(router.configuration.routes.length, 5);
    });

    testWidgets('TC-03-U-007: app boots in a ProviderScope and shows Search',
        (tester) async {
      await tester.pumpWidget(const ProviderScope(child: MergeMarketApp()));
      await tester.pumpAndSettle();

      // MaterialApp.router wired and Home/Search placeholder rendered.
      expect(find.byType(MaterialApp), findsOneWidget);
      expect(find.text('Search'), findsWidgets);
      expect(find.text('MergeMarket'), findsOneWidget); // Home AppBar title
    });

    testWidgets('TC-03-U-008: bottom nav shows the four primary destinations',
        (tester) async {
      await tester.pumpWidget(const ProviderScope(child: MergeMarketApp()));
      await tester.pumpAndSettle();

      expect(find.byType(NavigationBar), findsOneWidget);
      final bar = tester.widget<NavigationBar>(find.byType(NavigationBar));
      expect(bar.destinations.length, 4);
    });
  });
}
