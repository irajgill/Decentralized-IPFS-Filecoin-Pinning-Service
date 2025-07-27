-- Create pin_requests table
CREATE TABLE pin_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    cid VARCHAR(64) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    size_bytes BIGINT DEFAULT 0,
    price_fil DECIMAL(18,8) DEFAULT 0,
    duration_days INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_pin_requests_user_id ON pin_requests(user_id);
CREATE INDEX idx_pin_requests_cid ON pin_requests(cid);
CREATE INDEX idx_pin_requests_status ON pin_requests(status);
CREATE INDEX idx_pin_requests_created_at ON pin_requests(created_at);

-- Create updated_at trigger
CREATE TRIGGER update_pin_requests_updated_at BEFORE UPDATE
    ON pin_requests FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add constraints
ALTER TABLE pin_requests ADD CONSTRAINT check_status 
    CHECK (status IN ('pending', 'pinned', 'failed', 'cancelled'));
ALTER TABLE pin_requests ADD CONSTRAINT check_duration_days 
    CHECK (duration_days > 0);

-- Drop trigger
DROP TRIGGER IF EXISTS update_pin_requests_updated_at ON pin_requests;

-- Drop indexes
DROP INDEX IF EXISTS idx_pin_requests_user_id;
DROP INDEX IF EXISTS idx_pin_requests_cid;
DROP INDEX IF EXISTS idx_pin_requests_status;
DROP INDEX IF EXISTS idx_pin_requests_created_at;

-- Drop table
DROP TABLE IF EXISTS pin_requests;