import 'package:flutter/material.dart';

import '../widgets/placeholder_screen.dart';

/// Wishlist screen (USER_FLOWS Flow 4). Real UI lands in B-05.
class WishlistScreen extends StatelessWidget {
  const WishlistScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Wishlist')),
      body: const PlaceholderScreen(
        title: 'Wishlist',
        implementedIn: 'B-05',
        icon: Icons.favorite_border_rounded,
      ),
    );
  }
}
