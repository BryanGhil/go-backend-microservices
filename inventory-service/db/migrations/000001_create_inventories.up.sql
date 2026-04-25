CREATE TABLE IF NOT EXISTS inventories (
    product_id BIGINT PRIMARY KEY, -- 1:1 relationship with product
    stock_quantity INT DEFAULT 0,
    reserved_quantity INT DEFAULT 0, 
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);