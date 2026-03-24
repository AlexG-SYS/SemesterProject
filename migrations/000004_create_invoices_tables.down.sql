CREATE TABLE IF NOT EXISTS invoices (
    invoice_id bigserial PRIMARY KEY,
    order_id bigint UNIQUE NOT NULL REFERENCES orders(order_id) ON DELETE CASCADE,
    merchant_tin text NOT NULL,
    total_gst decimal(10, 2) NOT NULL, -- Calculated at 12.5%
    grand_total decimal(10, 2) NOT NULL,
    finalized_at timestamp(0) with time zone,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payments (
    payment_id bigserial PRIMARY KEY,
    order_id bigint NOT NULL REFERENCES orders(order_id) ON DELETE CASCADE,
    amount decimal(10, 2) NOT NULL,
    payment_method text NOT NULL, -- e.g., Bank Transfer, DigiWallet
    reference_number text,
    status text NOT NULL DEFAULT 'Pending', -- Verified, Pending, Failed
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);