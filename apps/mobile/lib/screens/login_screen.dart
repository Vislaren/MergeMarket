import 'package:flutter/material.dart';

import '../widgets/placeholder_screen.dart';

/// Login screen (USER_FLOWS Flow 1). Real UI lands in B-08.
///
/// Standalone route (no bottom navigation bar).
class LoginScreen extends StatelessWidget {
  const LoginScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Log In')),
      body: const PlaceholderScreen(
        title: 'Log In',
        implementedIn: 'B-08',
        icon: Icons.login_rounded,
      ),
    );
  }
}
