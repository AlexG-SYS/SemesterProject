CREATE TABLE IF NOT EXISTS users (
    user_id bigserial PRIMARY KEY,
    email text UNIQUE NOT NULL,
    password_hash text NOT NULL,
    role text NOT NULL DEFAULT 'Customer', -- Options: Admin, Staff, Customer
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS profiles (
    profile_id bigserial PRIMARY KEY,
    user_id bigint UNIQUE NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    full_name text NOT NULL,
    phone text,
    address text,
    district text,
    town_village text,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);