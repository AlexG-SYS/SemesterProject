CREATE TABLE IF NOT EXISTS categories (
    category_id bigserial PRIMARY KEY,
    name text UNIQUE NOT NULL,
    description text,
    is_active boolean DEFAULT true,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS products (
    product_id bigserial PRIMARY KEY,
    category_id bigint REFERENCES categories(category_id) ON DELETE SET NULL,
    name text NOT NULL,
    description text,
    is_gst_eligible boolean DEFAULT true, -- 12.5% Tax eligible
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS product_variants (
    variant_id bigserial PRIMARY KEY,
    product_id bigint NOT NULL REFERENCES products(product_id) ON DELETE CASCADE,
    sku text UNIQUE NOT NULL,
    size_attr text,
    color_attr text,
    cost_price decimal(10, 2) NOT NULL,
    selling_price decimal(10, 2) NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS locations (
    location_id bigserial PRIMARY KEY,
    name text NOT NULL, -- e.g., Belmopan
    address text,
    is_active boolean DEFAULT true,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS inventory (
    inventory_id bigserial PRIMARY KEY,
    variant_id bigint NOT NULL REFERENCES product_variants(variant_id) ON DELETE CASCADE,
    location_id bigint NOT NULL REFERENCES locations(location_id) ON DELETE CASCADE,
    stock_on_hand integer NOT NULL DEFAULT 0,
    stock_reserved integer NOT NULL DEFAULT 0, -- Atomic lock for orders
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);