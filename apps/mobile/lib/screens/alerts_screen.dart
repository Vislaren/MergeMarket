import 'package:flutter/material.dart';

import '../widgets/placeholder_screen.dart';

/// Alerts screen (USER_FLOWS Flow 5). Real UI lands in B-06.
class AlertsScreen extends StatelessWidget {
  const AlertsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Alerts')),
      body: const PlaceholderScreen(
        title: 'Alerts',
        implementedIn: 'B-06',
        icon: Icons.notifications_none_rounded,
      ),
    );
  }
}
