import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../screens/alerts_screen.dart';
import '../screens/home_screen.dart';
import '../screens/login_screen.dart';
import '../screens/product_detail_screen.dart';
import '../screens/register_screen.dart';
import '../screens/results_screen.dart';
import '../screens/savings_dashboard_screen.dart';
import '../screens/wishlist_screen.dart';
import '../widgets/app_shell.dart';

/// Route path constants — reference these instead of raw strings so
/// navigation call sites stay consistent and refactor-safe.
abstract final class Routes {
  static const home = '/';
  static const results = '/results';
  static const product = '/product'; // full path: /product/:id
  static const wishlist = '/wishlist';
  static const alerts = '/alerts';
  static const savings = '/savings';
  static const login = '/login';
  static const register = '/register';
}

/// Exposes the app's [GoRouter] so it can be created within the Riverpod
/// container and later read redirect/auth state from other providers (B-08).
///
/// Layout mirrors `COMPONENT_LIBRARY.md §Navigation`: the four primary
/// destinations live inside a [StatefulShellRoute] (bottom nav visible);
/// Results, Product Detail, Login, and Register are top-level routes that
/// render without the bottom bar.
final routerProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    initialLocation: Routes.home,
    routes: [
      StatefulShellRoute.indexedStack(
        builder: (context, state, navigationShell) =>
            AppShell(navigationShell: navigationShell),
        branches: [
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: Routes.home,
                builder: (context, state) => const HomeScreen(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: Routes.wishlist,
                builder: (context, state) => const WishlistScreen(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: Routes.alerts,
                builder: (context, state) => const AlertsScreen(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: Routes.savings,
                builder: (context, state) => const SavingsDashboardScreen(),
              ),
            ],
          ),
        ],
      ),
      GoRoute(
        path: Routes.results,
        builder: (context, state) =>
            ResultsScreen(query: state.uri.queryParameters['q']),
      ),
      GoRoute(
        path: '${Routes.product}/:id',
        builder: (context, state) =>
            ProductDetailScreen(productId: state.pathParameters['id']!),
      ),
      GoRoute(
        path: Routes.login,
        builder: (context, state) => const LoginScreen(),
      ),
      GoRoute(
        path: Routes.register,
        builder: (context, state) => const RegisterScreen(),
      ),
    ],
  );
});
