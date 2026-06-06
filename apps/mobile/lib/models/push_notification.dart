/// Typed model for an incoming push notification (B-10).
///
/// Both FCM (Android) and APNs (iOS) deliver a flat string→string `data` map.
/// This model is the single place that parses that map, so the notification
/// service, providers, and UI consume a typed value — never raw map keys.
///
/// Payload contract (sent by the History service on a price drop / restock):
/// ```json
/// {
///   "type": "price_drop" | "restock",
///   "product_id": "prod-001",
///   "title": "iPhone 15",
///   "body": "iPhone 15 dropped to XAF 799,000 on Amazon — below your alert",
///   "price": "799000",          // optional, string (FCM data is all strings)
///   "store": "Amazon",          // optional
///   "threshold": "850000"       // optional
/// }
/// ```
library;

/// The kind of event a push notification represents (USER_FLOWS Flow 6).
enum PushType {
  /// A tracked product's price fell below the user's alert threshold.
  priceDrop,

  /// A tracked product is back in stock.
  restock,

  /// An unrecognised payload — shown generically, never routed blindly.
  unknown;

  /// Maps the wire `type` string to a [PushType].
  static PushType fromWire(String? value) {
    switch (value) {
      case 'price_drop':
        return PushType.priceDrop;
      case 'restock':
        return PushType.restock;
      default:
        return PushType.unknown;
    }
  }
}

/// A parsed push notification.
class PushNotification {
  const PushNotification({
    required this.type,
    required this.productId,
    required this.title,
    required this.body,
    this.price,
    this.store,
    this.threshold,
  });

  final PushType type;
  final String productId;
  final String title;
  final String body;

  /// The new price that triggered the notification, if provided.
  final double? price;

  /// The store offering [price], if provided.
  final String? store;

  /// The user's alert threshold the price dropped below, if provided.
  final double? threshold;

  /// Whether this notification can deep-link to a product (Flow 6 requires a
  /// product id and a known type).
  bool get isRoutable => productId.isNotEmpty && type != PushType.unknown;

  /// Parses a flat FCM/APNs `data` map. Tolerant: unknown types and missing
  /// fields decode to safe values rather than throwing.
  factory PushNotification.fromData(Map<String, dynamic> data) {
    return PushNotification(
      type: PushType.fromWire(data['type'] as String?),
      productId: (data['product_id'] as String?)?.trim() ?? '',
      title: data['title'] as String? ?? '',
      body: data['body'] as String? ?? '',
      price: _toDouble(data['price']),
      store: data['store'] as String?,
      threshold: _toDouble(data['threshold']),
    );
  }

  /// Coerces a value that may arrive as a number or a numeric string (FCM data
  /// payloads are always strings) into a double, or null if absent/invalid.
  static double? _toDouble(Object? value) {
    if (value == null) return null;
    if (value is num) return value.toDouble();
    if (value is String) return double.tryParse(value);
    return null;
  }
}
