# Normalization Service (Go)

Consumes raw scrape output and converts it into the standard product schema
(`price, currency, title, image_url, shipping, affiliate_url`). Includes the
Affiliate Link Injection module that wraps outbound URLs with retailer-
specific parameters.

- **Port:** 8084
- **Endpoints:** `GET /health`
- **Task:** A-06

_Implementation pending._
