-- Write your migrate up statements here

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    seller_id UUID NOT NULL REFERENCES users (id),

    product_name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,

    base_price FLOAT NOT NULL,
    auction_start TIMESTAMPTZ NOT NULL, -- This will be deleted
    auction_end TIMESTAMPTZ NOT NULL,

    is_sold BOOLEAN NOT NULL DEFAULT false,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

---- create above / drop below ----

DROP TABLE IF EXISTS products;

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
