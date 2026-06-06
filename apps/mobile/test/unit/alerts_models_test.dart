// Unit tests for the alerts data models (B-06).
//
// Test artefacts: docs/testing/session-09/unit/.

import 'package:flutter_test/flutter_test.dart';
import 'package:mergemarket/models/alert.dart';

import '../mocks/alerts_mock_data.dart';

void main() {
  group('B-06 Alert model', () {
    test('TC-09-U-001: decodes the alerts contract', () {
      final list = AlertList.fromJson(kAlertsJson);
      expect(list.alerts, hasLength(2));
      final first = list.alerts.first;
      expect(first.alertId, 'al-001');
      expect(first.thresholdPrice, 230000);
      expect(first.isActive, isTrue);
      expect(list.alerts[1].isActive, isFalse);
    });

    test('TC-09-U-002: empty alerts decode to no entries', () {
      expect(AlertList.fromJson(kAlertsEmptyJson).alerts, isEmpty);
    });

    test('TC-09-U-003: missing fields fall back to safe defaults', () {
      final a = Alert.fromJson({'alert_id': 'x'});
      expect(a.thresholdPrice, 0);
      expect(a.isActive, isFalse);
      expect(a.currency, '');
    });
  });
}
