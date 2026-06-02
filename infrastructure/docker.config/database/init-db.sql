-- ============================================================
-- PostgreSQL + pgvector
-- ============================================================
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pg_search;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Function hỗ trợ tìm kiếm không dấu
CREATE OR REPLACE FUNCTION immutable_unaccent(text)
RETURNS text LANGUAGE sql IMMUTABLE STRICT PARALLEL SAFE AS
$$ SELECT public.unaccent($1) $$;

-- ============================================================
-- NHÓM 1: CẤU TRÚC CHO AGENTS
-- ============================================================


CREATE TABLE knowledge_base (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title      VARCHAR(200) NOT NULL,
    content    TEXT NOT NULL,
    embedding  vector(1024),                   -- OpenAI / Anthropic embedding dim
    metadata   JSONB,
    category   VARCHAR(50),
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE IF NOT EXISTS memory_entries (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    app_name VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    event_id VARCHAR(255) NOT NULL,
    author VARCHAR(255),
    content JSONB NOT NULL,
    content_text TEXT NOT NULL,
    embedding  vector(1024),
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(app_name, user_id, session_id, event_id)
);


-- ============================================================
-- NHÓM 2: E-COMMERCE CORE TABLES
-- ============================================================

-- 1. Users & Address
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    gender VARCHAR(20),
    dob DATE,
    role VARCHAR(50) NOT NULL DEFAULT 'customer',
    avatar TEXT,
    is_lock BOOLEAN NOT NULL DEFAULT FALSE,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE address (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    full_name VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    district VARCHAR(100) NOT NULL,
    province VARCHAR(100) NOT NULL,
    ward VARCHAR(100) NOT NULL,
    detail_address TEXT NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

-- 2. Store
CREATE TABLE store (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    hotline VARCHAR(50),
    district VARCHAR(100) NOT NULL,
    province VARCHAR(100) NOT NULL,
    ward VARCHAR(100) NOT NULL,
    road VARCHAR(255),
    email VARCHAR(255),
    lat NUMERIC(10, 8),  -- tọa độ lat
    lng NUMERIC(11, 8),  -- tọa độ lng (khuyên dùng PostGIS geography(Point, 4326))
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Brand & Catalog
CREATE TABLE brand (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    logo_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE category (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    parent_id INTEGER REFERENCES category(id) ON DELETE SET NULL,
    icon VARCHAR(255),
    slug VARCHAR(255) UNIQUE NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

-- 4. Product & Specs
CREATE TABLE product (
    id VARCHAR(50) PRIMARY KEY,
    category_id INTEGER NOT NULL REFERENCES category(id),
    brand_id INTEGER NOT NULL REFERENCES brand(id),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    meta_title VARCHAR(255),
    meta_description TEXT,
    img_thumb TEXT,
    weight NUMERIC(10, 2),
    low_stock_threshold INTEGER NOT NULL DEFAULT 5,
    specs_jsonb JSONB,  -- snapshot specs for filter
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE product_spec (
    id SERIAL PRIMARY KEY,
    product_id VARCHAR(50) NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    "group" VARCHAR(100) NOT NULL,
    key VARCHAR(100) NOT NULL,
    value TEXT NOT NULL,
    unit VARCHAR(50),
    sort_order INTEGER NOT NULL DEFAULT 0
);

-- 5. Variant & Option
CREATE TABLE product_option_type (
    id SERIAL PRIMARY KEY,
    product_id VARCHAR(50) NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE product_option_value (
    id SERIAL PRIMARY KEY,
    option_type_id INTEGER NOT NULL REFERENCES product_option_type(id) ON DELETE CASCADE,
    value VARCHAR(255) NOT NULL,
    color_code VARCHAR(20),
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE product_variant (
    id SERIAL PRIMARY KEY,
    product_id VARCHAR(50) NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    sku VARCHAR(100) UNIQUE NOT NULL,
    price NUMERIC(15, 2) NOT NULL,
    price_base NUMERIC(15, 2),
    weight NUMERIC(10, 2),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE product_image (
    id SERIAL PRIMARY KEY,
    product_id VARCHAR(50) NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    variant_id INTEGER REFERENCES product_variant(id) ON DELETE SET NULL,
    url TEXT NOT NULL,
    alt_text VARCHAR(255),
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_thumbnail BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE product_variant_option (
    id SERIAL PRIMARY KEY,
    variant_id INTEGER NOT NULL REFERENCES product_variant(id) ON DELETE CASCADE,
    option_value_id INTEGER NOT NULL REFERENCES product_option_value(id) ON DELETE CASCADE,
    UNIQUE (variant_id, option_value_id)
);

-- 6. Inventory & Suppliers
CREATE TABLE product_inventory (
    variant_id INTEGER NOT NULL REFERENCES product_variant(id) ON DELETE CASCADE,
    store_id INTEGER NOT NULL REFERENCES store(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL DEFAULT 0,
    reserved INTEGER NOT NULL DEFAULT 0,
    last_updated TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (variant_id, store_id)
);

CREATE TABLE inventory_reservations (
    id VARCHAR(100) PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    store_id INTEGER NOT NULL REFERENCES store(id) ON DELETE CASCADE,
    items JSONB NOT NULL,
    status VARCHAR(50) NOT NULL,
    payment_code VARCHAR(100),
    payos_order_code VARCHAR(100),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE suppliers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    contact_name VARCHAR(255),
    contact_phone VARCHAR(50),
    contact_email VARCHAR(255),
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE import_invoices (
    id SERIAL PRIMARY KEY,
    supplier_id INTEGER NOT NULL REFERENCES suppliers(id),
    store_id INTEGER NOT NULL REFERENCES store(id),
    created_by INTEGER NOT NULL REFERENCES users(id),
    total_items INTEGER NOT NULL DEFAULT 0,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE import_invoice_details (
    id SERIAL PRIMARY KEY,
    invoice_id INTEGER NOT NULL REFERENCES import_invoices(id) ON DELETE CASCADE,
    variant_id INTEGER NOT NULL REFERENCES product_variant(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    stock_before INTEGER NOT NULL,
    price_import NUMERIC(15, 2) NOT NULL
);

CREATE TABLE inventory_log (
    id SERIAL PRIMARY KEY,
    variant_id INTEGER NOT NULL REFERENCES product_variant(id) ON DELETE CASCADE,
    store_id INTEGER NOT NULL REFERENCES store(id) ON DELETE CASCADE,
    change_qty INTEGER NOT NULL,      -- dương = nhập, âm = xuất
    qty_after INTEGER NOT NULL,       -- snapshot sau thay đổi
    reason VARCHAR(100) NOT NULL,     -- "order_confirmed", "order_cancelled", "import", "manual_adjust"
    ref_id VARCHAR(100),              -- order_id hoặc invoice_id tương ứng
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- 7. Cart
CREATE TABLE cart_items (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    session_id VARCHAR(100),
    variant_id INTEGER NOT NULL REFERENCES product_variant(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 8. Promotion & Voucher
CREATE TABLE promotions (
    id SERIAL PRIMARY KEY,
    product_id VARCHAR(50) NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    variant_id INTEGER REFERENCES product_variant(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    discount_type VARCHAR(50) NOT NULL,
    discount_value NUMERIC(15, 2) NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE vouchers (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    discount_type VARCHAR(50) NOT NULL,
    discount_value NUMERIC(15, 2) NOT NULL,
    discount_target VARCHAR(50) NOT NULL DEFAULT 'order',
    min_order_value NUMERIC(15, 2) NOT NULL DEFAULT 0,
    max_discount_amount NUMERIC(15, 2),
    max_usage_total INTEGER,
    max_usage_per_user INTEGER NOT NULL DEFAULT 1,
    used_count INTEGER NOT NULL DEFAULT 0,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

-- 9. Order & Status
CREATE TABLE order_status (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    label VARCHAR(100) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE payment_status (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    label VARCHAR(100) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE shipping_status (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    label VARCHAR(100) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    order_code VARCHAR(100) UNIQUE NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id),
    store_id INTEGER NOT NULL REFERENCES store(id),
    voucher_id INTEGER REFERENCES vouchers(id),
    order_status_id INTEGER NOT NULL REFERENCES order_status(id),
    payment_status_id INTEGER NOT NULL REFERENCES payment_status(id),
    shipping_status_id INTEGER NOT NULL REFERENCES shipping_status(id),
    total_amount NUMERIC(15, 2) NOT NULL,
    voucher_discount NUMERIC(15, 2) NOT NULL DEFAULT 0,
    shipping_price NUMERIC(15, 2) NOT NULL DEFAULT 0,
    payment_method VARCHAR(50),
    payment_code VARCHAR(100),
    payos_order_code VARCHAR(100),
    note TEXT,
    receiver_name VARCHAR(255) NOT NULL,
    receiver_address TEXT NOT NULL,
    receiver_phone VARCHAR(50) NOT NULL,
    sender_name VARCHAR(255),
    sender_address TEXT,
    sender_phone VARCHAR(50),
    shipping_provider  VARCHAR(50),   -- "ghn", "ghtk"
shipping_code      VARCHAR(100),  -- mã vận đơn
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE voucher_usages (
    id SERIAL PRIMARY KEY,
    voucher_id INTEGER NOT NULL REFERENCES vouchers(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (voucher_id, user_id, order_id)
);

CREATE TABLE order_details (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    variant_id INTEGER NOT NULL REFERENCES product_variant(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(15, 2) NOT NULL,
    total_cost NUMERIC(15, 2) NOT NULL
);

CREATE TABLE order_status_history (
    id SERIAL PRIMARY KEY,
    order_id      INTEGER NOT NULL REFERENCES orders(id),
    status_type   VARCHAR(20) NOT NULL,  -- "order", "payment", "shipping"
    from_status   VARCHAR(50),
    to_status     VARCHAR(50) NOT NULL,
    changed_by    INTEGER REFERENCES users(id),
    note          TEXT,
    changed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 10. Reviews
CREATE TABLE reviews (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id VARCHAR(50) NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT,
    images JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 11. Banner on homepage website
CREATE TABLE banners (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    subtitle VARCHAR(250),
    description TEXT,
    image_url TEXT NOT NULL,
    tag VARCHAR(100),
    link_url VARCHAR(255),
    sort_order INT DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- INDEXES
-- ============================================================

-- E-Commerce Foreign Key & Query Indexes
CREATE INDEX IF NOT EXISTS idx_address_user ON address(user_id);
CREATE INDEX IF NOT EXISTS idx_product_category ON product(category_id);
CREATE INDEX IF NOT EXISTS idx_product_brand ON product(brand_id);
CREATE INDEX IF NOT EXISTS idx_product_specs_jsonb ON product USING GIN (specs_jsonb);
CREATE INDEX IF NOT EXISTS idx_product_spec_product ON product_spec(product_id);
CREATE INDEX IF NOT EXISTS idx_product_option_type_product ON product_option_type(product_id);
CREATE INDEX IF NOT EXISTS idx_product_option_value_type ON product_option_value(option_type_id);
CREATE INDEX IF NOT EXISTS idx_product_variant_product ON product_variant(product_id);
CREATE INDEX IF NOT EXISTS idx_product_image_product_variant ON product_image(product_id, variant_id);
CREATE INDEX IF NOT EXISTS idx_product_inventory_store ON product_inventory(store_id);
CREATE INDEX IF NOT EXISTS idx_inventory_res_user ON inventory_reservations(user_id);
CREATE INDEX IF NOT EXISTS idx_inventory_res_expiry ON inventory_reservations(expires_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_import_invoices_supplier ON import_invoices(supplier_id);
CREATE INDEX IF NOT EXISTS idx_import_invoices_store ON import_invoices(store_id);
CREATE INDEX IF NOT EXISTS idx_import_invoice_details_invoice ON import_invoice_details(invoice_id);
CREATE INDEX IF NOT EXISTS idx_inventory_log_store_var ON inventory_log(store_id, variant_id);
CREATE INDEX IF NOT EXISTS idx_inventory_log_created ON inventory_log(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_cart_items_user ON cart_items(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_cart_items_session ON cart_items(session_id) WHERE session_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_promotions_product ON promotions(product_id, variant_id);
CREATE INDEX IF NOT EXISTS idx_promotions_date ON promotions(start_date, end_date) WHERE is_active = TRUE AND is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_vouchers_code ON vouchers(code);
CREATE INDEX IF NOT EXISTS idx_vouchers_active ON vouchers(code, start_date, end_date) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_orders_user ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_code ON orders(order_code);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(order_status_id, payment_status_id, shipping_status_id);
CREATE INDEX IF NOT EXISTS idx_order_details_order ON order_details(order_id);
CREATE INDEX IF NOT EXISTS idx_order_status_history_order ON order_status_history(order_id);
CREATE INDEX IF NOT EXISTS idx_reviews_product ON reviews(product_id);
CREATE INDEX IF NOT EXISTS idx_reviews_order ON reviews(order_id);

-- ============================================================
-- HYBRID SEARCH (BM25 + Vector)
-- ============================================================

-- Vector Index (HNSW)
CREATE INDEX idx_kb_embedding 
    ON knowledge_base 
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 128); -- Tăng lên 128 để RAG chính xác hơn

-- Filter Index
CREATE INDEX idx_kb_filter ON knowledge_base (category) WHERE is_active = TRUE;


-- BM25 Index (ParadeDB)
CREATE INDEX knowledge_base_search_bm25_index ON knowledge_base
USING bm25 (id, title, content)
WITH (
    key_field = 'id',
    text_fields = '{
        "title":   {"tokenizer": {"type": "icu"}},
        "content": {"tokenizer": {"type": "icu"}}
    }'
);


CREATE INDEX memory_entries_search_bm25_index ON memory_entries
USING bm25 (id, content_text)
WITH (
    key_field = 'id',
    text_fields = '{
        "content_text": {"tokenizer": {"type": "icu"}}
    }'
);