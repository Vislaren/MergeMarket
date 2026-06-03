# MergeMarket — Database Schema

## 1. PostgreSQL — Relational Tables

### users
```sql
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    country_code  CHAR(2) NOT NULL DEFAULT 'CM',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### stores
```sql
CREATE TABLE stores (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT UNIQUE NOT NULL,
    base_url    TEXT NOT NULL,
    config_path TEXT NOT NULL,      -- path to StoreConfig JSON/YAML file
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### products
```sql
CREATE TABLE products (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id      UUID NOT NULL REFERENCES stores(id),
    title         TEXT NOT NULL,
    url           TEXT NOT NULL,
    affiliate_url TEXT,
    image_url     TEXT,
    currency      CHAR(3) NOT NULL DEFAULT 'USD',
    last_price    NUMERIC(12, 2),
    last_shipping NUMERIC(12, 2),
    scraped_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_store_id ON products(store_id);
CREATE INDEX idx_products_scraped_at ON products(scraped_at);
```

### wishlist_items
```sql
CREATE TABLE wishlist_items (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    added_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, product_id)
);

CREATE INDEX idx_wishlist_user_id ON wishlist_items(user_id);
```

### price_alerts
```sql
CREATE TABLE price_alerts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id      UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    threshold_price NUMERIC(12, 2) NOT NULL,
    currency        CHAR(3) NOT NULL DEFAULT 'USD',
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alerts_user_id ON price_alerts(user_id);
CREATE INDEX idx_alerts_product_id ON price_alerts(product_id);
```

### return_policies
```sql
CREATE TABLE return_policies (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id    UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    policy_text TEXT NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (store_id)
);
```

### scrape_jobs
```sql
CREATE TYPE job_status AS ENUM ('pending', 'running', 'done', 'failed');

CREATE TABLE scrape_jobs (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id     UUID NOT NULL REFERENCES stores(id),
    query        TEXT NOT NULL,
    status       job_status NOT NULL DEFAULT 'pending',
    error        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_jobs_status ON scrape_jobs(status);
CREATE INDEX idx_jobs_created_at ON scrape_jobs(created_at);
```

---

## 2. TimescaleDB — Price History Hypertable

```sql
CREATE TABLE price_history (
    product_id  UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    price       NUMERIC(12, 2) NOT NULL,
    shipping    NUMERIC(12, 2),
    currency    CHAR(3) NOT NULL DEFAULT 'USD',
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Convert to TimescaleDB hypertable partitioned by time
SELECT create_hypertable('price_history', 'recorded_at');

CREATE INDEX idx_price_history_product ON price_history(product_id, recorded_at DESC);
```

---

## 3. Redis — Key Schema

| Key Pattern | Value Type | TTL | Description |
|---|---|---|---|
| `store_meta:{store_id}` | JSON string | 24 hrs | CSS selectors and API paths for the store |
| `proxy_pool` | Redis Set | 5 mins | Working proxy IP:PORT strings |
| `session:{user_id}` | JSON string | 1 hr | User session data and preferences |
| `product_score:{product_id}` | JSON string | 24–48 hrs | AI sentiment score result |
| `search:{query_hash}` | JSON string | 5 min – 24 hrs | Aggregated search results |
| `circuit:{store_id}` | String (open/closed) | Until manual reset | Circuit breaker state per store |

---

## 4. Environment Variables Reference

```env
# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_NAME=mergemarket
DB_USER=postgres
DB_PASSWORD=changeme

# TimescaleDB (same connection as Postgres if using TimescaleDB extension)
TIMESCALE_HOST=localhost
TIMESCALE_PORT=5432

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Auth
JWT_SECRET=changeme
JWT_EXPIRY_HOURS=1
REFRESH_TOKEN_EXPIRY_DAYS=30

# Services
SCRAPER_PORT=8083
NORMALIZATION_PORT=8084
HISTORY_PORT=8085
PROXY_VALIDATOR_PORT=8086
AUTH_PORT=8081
BFF_PORT=8082

# SonarQube
SONAR_HOST_URL=http://your_vps_ip:9000
SONAR_TOKEN=changeme

# Firebase (Push Notifications)
FIREBASE_SERVER_KEY=changeme
```
