-- ─────────────────────────────────────────────────────────────────────────
-- MergeMarket — Local Dev Database Bootstrap
-- ─────────────────────────────────────────────────────────────────────────
-- This script runs ONCE, automatically, the first time the Postgres/Timescale
-- container initialises an empty data volume (Docker entrypoint convention:
-- everything in /docker-entrypoint-initdb.d/ is executed in lexical order).
--
-- It is the canonical local-dev mirror of project_docs/database/DATABASE_SCHEMA.md.
-- Keep the two in sync. This is a convenience for local development only; the
-- production schema lifecycle (migrations) is owned by the services that need
-- it. Re-running requires destroying the named volume (`docker compose down -v`).
-- ─────────────────────────────────────────────────────────────────────────

-- TimescaleDB ships with the timescale/timescaledb image; enable it so the
-- price_history hypertable below can be created on this single connection.
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- gen_random_uuid() lives in pgcrypto on older PGs; it is built-in on PG13+,
-- but enabling the extension keeps this portable.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ── Relational tables (PostgreSQL) ─────────────────────────────────────────

CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    country_code  CHAR(2) NOT NULL DEFAULT 'CM',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS stores (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT UNIQUE NOT NULL,
    base_url    TEXT NOT NULL,
    config_path TEXT NOT NULL,      -- path to StoreConfig JSON/YAML file
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS products (
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

CREATE INDEX IF NOT EXISTS idx_products_store_id   ON products(store_id);
CREATE INDEX IF NOT EXISTS idx_products_scraped_at ON products(scraped_at);

CREATE TABLE IF NOT EXISTS wishlist_items (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    added_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_wishlist_user_id ON wishlist_items(user_id);

CREATE TABLE IF NOT EXISTS price_alerts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id      UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    threshold_price NUMERIC(12, 2) NOT NULL,
    currency        CHAR(3) NOT NULL DEFAULT 'USD',
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alerts_user_id    ON price_alerts(user_id);
CREATE INDEX IF NOT EXISTS idx_alerts_product_id ON price_alerts(product_id);

CREATE TABLE IF NOT EXISTS return_policies (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id    UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    policy_text TEXT NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (store_id)
);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'job_status') THEN
        CREATE TYPE job_status AS ENUM ('pending', 'running', 'done', 'failed');
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS scrape_jobs (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id     UUID NOT NULL REFERENCES stores(id),
    query        TEXT NOT NULL,
    status       job_status NOT NULL DEFAULT 'pending',
    error        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_jobs_status     ON scrape_jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON scrape_jobs(created_at);

-- ── Time-series hypertable (TimescaleDB) ───────────────────────────────────

CREATE TABLE IF NOT EXISTS price_history (
    product_id  UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    price       NUMERIC(12, 2) NOT NULL,
    shipping    NUMERIC(12, 2),
    currency    CHAR(3) NOT NULL DEFAULT 'USD',
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Idempotent: if_not_exists avoids errors if the script is ever re-applied.
SELECT create_hypertable('price_history', 'recorded_at', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_price_history_product
    ON price_history(product_id, recorded_at DESC);
