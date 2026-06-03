# MergeMarket — API Contracts

> These are the canonical request/response schemas for all services.
> Agent A implements these exactly.
> Agent B builds mocks that return these exact shapes.
> Neither agent changes a contract without logging it in their DONE.md.

All endpoints are prefixed with `/api/v1/`.
All responses are JSON. All timestamps are ISO 8601.
Authentication: Bearer JWT in `Authorization` header (all routes except auth).

---

## Auth Service

### POST /api/v1/auth/register
```json
Request:
{ "email": "string", "password": "string" }

Response 201:
{ "token": "string", "refresh_token": "string", "expires_at": "ISO8601" }

Response 400:
{ "error": "invalid_input", "message": "string" }

Response 409:
{ "error": "email_exists", "message": "string" }
```

### POST /api/v1/auth/login
```json
Request:
{ "email": "string", "password": "string" }

Response 200:
{ "token": "string", "refresh_token": "string", "expires_at": "ISO8601" }

Response 401:
{ "error": "invalid_credentials", "message": "string" }
```

### POST /api/v1/auth/refresh
```json
Request:
{ "refresh_token": "string" }

Response 200:
{ "token": "string", "refresh_token": "string", "expires_at": "ISO8601" }

Response 401:
{ "error": "token_expired", "message": "string" }
```

---

## Search

### GET /api/v1/search
```
Query params:
  q          string  required  — product search query
  location   string  required  — ISO 3166-1 alpha-2 country code (e.g. "CM")

Response 200:
{
  "query": "string",
  "results": [
    {
      "product_id":    "string",
      "title":         "string",
      "price":         number,
      "currency":      "string",
      "shipping":      number,
      "total_cost":    number,
      "image_url":     "string",
      "store":         "string",
      "affiliate_url": "string",
      "deal_score":    number,       // 0–100, AI Deal Meter
      "scraped_at":    "ISO8601"
    }
  ],
  "cached":     boolean,
  "latency_ms": number
}

Response 400:
{ "error": "missing_query", "message": "string" }

Response 504:
{ "error": "timeout", "message": "Results took too long. Try again." }
```

---

## Products & Price History

### GET /api/v1/products/{product_id}/history
```json
Response 200:
{
  "product_id": "string",
  "title": "string",
  "history": [
    { "price": number, "currency": "string", "recorded_at": "ISO8601" }
  ],
  "average_6m": number,
  "lowest_30d": number
}

Response 404:
{ "error": "not_found", "message": "string" }
```

### GET /api/v1/products/{product_id}/truth-score
```json
Response 200:
{
  "product_id": "string",
  "score": number,          // 0–100
  "sentiment": "positive" | "mixed" | "negative",
  "fake_review_risk": "low" | "medium" | "high",
  "summary": "string"
}
```

---

## Wishlist

### GET /api/v1/wishlist
```json
Response 200:
{
  "items": [
    {
      "wishlist_id": "string",
      "product_id":  "string",
      "title":       "string",
      "image_url":   "string",
      "stores": [
        { "store": "string", "price": number, "total_cost": number }
      ],
      "added_at": "ISO8601"
    }
  ]
}
```

### POST /api/v1/wishlist
```json
Request:
{ "product_id": "string" }

Response 201:
{ "wishlist_id": "string", "added_at": "ISO8601" }

Response 409:
{ "error": "already_in_wishlist", "message": "string" }
```

### DELETE /api/v1/wishlist/{wishlist_id}
```json
Response 204: (no body)
Response 404: { "error": "not_found", "message": "string" }
```

---

## Alerts

### GET /api/v1/alerts
```json
Response 200:
{
  "alerts": [
    {
      "alert_id":        "string",
      "product_id":      "string",
      "title":           "string",
      "threshold_price": number,
      "currency":        "string",
      "is_active":       boolean,
      "created_at":      "ISO8601"
    }
  ]
}
```

### POST /api/v1/alerts
```json
Request:
{ "product_id": "string", "threshold_price": number, "currency": "string" }

Response 201:
{ "alert_id": "string", "created_at": "ISO8601" }
```

### DELETE /api/v1/alerts/{alert_id}
```json
Response 204: (no body)
Response 404: { "error": "not_found", "message": "string" }
```

---

## Savings Dashboard

### GET /api/v1/savings
```json
Response 200:
{
  "total_saved":   number,
  "currency":      "string",
  "transactions": [
    {
      "product_id":  "string",
      "title":       "string",
      "saved":       number,
      "bought_at":   "ISO8601"
    }
  ]
}
```

---

## Health (all services)

### GET /health
```json
Response 200:
{ "status": "ok", "service": "string", "version": "string" }
```

---

## Error Shape (all endpoints)

All errors follow this shape:
```json
{
  "error":   "snake_case_code",
  "message": "Human readable description"
}
```

Standard HTTP status codes apply:
`400` Bad request · `401` Unauthorized · `403` Forbidden ·
`404` Not found · `409` Conflict · `429` Rate limited · `500` Server error
