CREATE TABLE IF NOT EXISTS shipping_methods (
    method_id bigserial PRIMARY KEY,
    provider_name text NOT NULL,
    service_type text NOT NULL,
    base_rate decimal(10, 2) NOT NULL,
    contact_phone text,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS orders (
    order_id bigserial PRIMARY KEY,
    customer_id bigint NOT NULL REFERENCES profiles(profile_id),
    shipping_method_id bigint REFERENCES shipping_methods(method_id),
    order_date timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    status text NOT NULL DEFAULT 'Pending', -- Pending, Paid, Cancelled
    shipping_fee decimal(10, 2) NOT NULL DEFAULT 0.00,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
    order_item_id bigserial PRIMARY KEY,
    order_id bigint NOT NULL REFERENCES orders(order_id) ON DELETE CASCADE,
    variant_id bigint NOT NULL REFERENCES product_variants(variant_id),
    quantity integer NOT NULL,
    price_at_reserve decimal(10, 2) NOT NULL,
    cost_at_reserve decimal(10, 2) NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);