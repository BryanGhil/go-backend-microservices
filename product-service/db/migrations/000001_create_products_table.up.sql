CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    seller_id BIGINT NOT NULL,          -- Important for your "Seller" role access
    name VARCHAR(255) NOT NULL,         -- 50 might be too short for product names
    description TEXT,                   -- Essential for Elasticsearch indexing
    category VARCHAR(100),              -- Needed for the Recommendation System logic
    price NUMERIC(20,2) NOT NULL,
    image_url TEXT,
    is_active BOOLEAN DEFAULT TRUE,     -- Soft delete / hide product
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);