CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT,
    full_name VARCHAR(100) NOT NULL,
    phone_number VARCHAR(30),
    address TEXT,
    role VARCHAR(20) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,        -- For account moderation
    last_login_at TIMESTAMP WITH TIME ZONE, -- Signal for Recommendation Engine
    provider VARCHAR(50) DEFAULT 'local',
    provider_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);