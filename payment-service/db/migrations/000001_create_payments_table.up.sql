CREATE TABLE IF NOT EXISTS payments (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,        -- Link to user service
    amount NUMERIC(20, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'IDR',
    status VARCHAR(50) NOT NULL,
    payment_method VARCHAR(50),     -- 'credit_card', 'bank_transfer'
    external_id VARCHAR(255),       -- The ID from the payment gateway
    correlation_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);