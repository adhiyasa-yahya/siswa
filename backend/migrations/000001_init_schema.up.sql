-- =============================================================================
-- SISWA - Sistem Item Stok Warehouse & Aset
-- 000001_init_schema : master data, stock and request log tables + indexes
-- =============================================================================

-- Trigram support so the "search by name/SKU" filter (ILIKE %term%) can use an index.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ---------------------------------------------------------------------------
-- categories : master data kategori item
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS categories (
    id     BIGSERIAL PRIMARY KEY,
    name   VARCHAR(100) NOT NULL,
    status VARCHAR(16)  NOT NULL DEFAULT 'active'
           CONSTRAINT categories_status_check CHECK (status IN ('active', 'inactive'))
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_categories_name ON categories (name);
-- categories.id is already indexed by its primary key; the explicit index below
-- keeps the requested index list visible/greppable in one place.
CREATE INDEX IF NOT EXISTS idx_categories_status ON categories (status);

-- ---------------------------------------------------------------------------
-- locations : lokasi rak / zona di gudang
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS locations (
    id   BIGSERIAL PRIMARY KEY,
    code VARCHAR(32) NOT NULL,
    zone VARCHAR(32) NOT NULL,
    type VARCHAR(32) NOT NULL
         CONSTRAINT locations_type_check CHECK (type IN ('rack', 'bin', 'pallet', 'cold_storage', 'staging'))
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_locations_code ON locations (code);
CREATE INDEX IF NOT EXISTS idx_locations_zone ON locations (zone);

-- ---------------------------------------------------------------------------
-- items : master data barang
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS items (
    id          BIGSERIAL PRIMARY KEY,
    sku         VARCHAR(64)  NOT NULL,
    name        VARCHAR(160) NOT NULL,
    category_id BIGINT       NOT NULL REFERENCES categories (id) ON DELETE RESTRICT,
    unit        VARCHAR(24)  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ  NULL
);

-- SKU uniqueness is enforced on live rows only, so a SKU becomes reusable once
-- the item has been soft deleted. Duplicates are reported as HTTP 409 (code 1003).
CREATE UNIQUE INDEX IF NOT EXISTS ux_items_sku_active ON items (sku) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_items_sku            ON items (sku);
CREATE INDEX IF NOT EXISTS idx_items_name           ON items (name);
CREATE INDEX IF NOT EXISTS idx_items_category_id    ON items (category_id);
CREATE INDEX IF NOT EXISTS idx_items_deleted_at     ON items (deleted_at);
CREATE INDEX IF NOT EXISTS idx_items_created_at     ON items (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_items_sku_trgm       ON items USING GIN (sku gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_items_name_trgm      ON items USING GIN (name gin_trgm_ops);

-- ---------------------------------------------------------------------------
-- stocks : kuantitas item per lokasi
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS stocks (
    id          BIGSERIAL PRIMARY KEY,
    item_id     BIGINT      NOT NULL REFERENCES items (id) ON DELETE CASCADE,
    location_id BIGINT      NOT NULL REFERENCES locations (id) ON DELETE RESTRICT,
    qty         BIGINT      NOT NULL DEFAULT 0
                CONSTRAINT stocks_qty_non_negative CHECK (qty >= 0),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- One row per item/location pair: receiving stock upserts on this key.
CREATE UNIQUE INDEX IF NOT EXISTS ux_stocks_item_location ON stocks (item_id, location_id);
CREATE INDEX IF NOT EXISTS idx_stocks_item_id     ON stocks (item_id);
CREATE INDEX IF NOT EXISTS idx_stocks_location_id ON stocks (location_id);

-- ---------------------------------------------------------------------------
-- request_logs : realtime logging of every request lifecycle
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS request_logs (
    id               VARCHAR(36) PRIMARY KEY,
    method           VARCHAR(10) NOT NULL,
    api_url          TEXT        NOT NULL,
    payload          TEXT        NOT NULL DEFAULT '',
    executed_time_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
    status           INTEGER     NOT NULL,
    response_code    INTEGER     NOT NULL DEFAULT 0,
    client_ip        VARCHAR(64) NOT NULL DEFAULT '',
    user_agent       TEXT        NOT NULL DEFAULT '',
    response_bytes   INTEGER     NOT NULL DEFAULT 0,
    datetime         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_request_logs_datetime ON request_logs (datetime DESC);
CREATE INDEX IF NOT EXISTS idx_request_logs_method   ON request_logs (method);
CREATE INDEX IF NOT EXISTS idx_request_logs_status   ON request_logs (status);
