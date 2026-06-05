import 'package:flutter/material.dart';

import '../widgets/placeholder_screen.dart';

/// Register screen (USER_FLOWS Flow 1). Real UI lands in B-08.
///
/// Standalone route (no bottom navigation bar).
class RegisterScreen extends StatelessWidget {
  const RegisterScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Register')),
      body: const PlaceholderScreen(
        title: 'Register',
        implementedIn: 'B-08',
        icon: Icons.person_add_alt_rounded,
      ),
    );
  }
}
