CREATE TABLE IF NOT EXISTS shipments (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id),
    seller_id BIGINT NOT NULL,          -- Only this seller can update this shipment
    tracking_number VARCHAR(255),
    carrier VARCHAR(100),               -- e.g., 'FedEx', 'DHL', 'JNE'
    status VARCHAR(50) NOT NULL,        -- Valid values: PENDING, PICKED_UP, IN_TRANSIT, OUT_OF_DELIVERY, DELIVERED, FAILED_DELIVERY
    
    shipped_at TIMESTAMP WITH TIME ZONE,
    estimated_arrival TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);