-- Create filecoin_deals table
CREATE TABLE filecoin_deals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pin_request_id UUID NOT NULL REFERENCES pin_requests(id) ON DELETE CASCADE,
    deal_cid VARCHAR(64),
    miner_id VARCHAR(20) NOT NULL,
    start_epoch BIGINT NOT NULL DEFAULT 0,
    end_epoch BIGINT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    storage_price DECIMAL(18,8) DEFAULT 0,
    retrieval_cost DECIMAL(18,8) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_filecoin_deals_pin_request_id ON filecoin_deals(pin_request_id);
CREATE INDEX idx_filecoin_deals_deal_cid ON filecoin_deals(deal_cid);
CREATE INDEX idx_filecoin_deals_miner_id ON filecoin_deals(miner_id);
CREATE INDEX idx_filecoin_deals_status ON filecoin_deals(status);
CREATE INDEX idx_filecoin_deals_end_epoch ON filecoin_deals(end_epoch);

-- Create updated_at trigger
CREATE TRIGGER update_filecoin_deals_updated_at BEFORE UPDATE
    ON filecoin_deals FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add constraints
ALTER TABLE filecoin_deals ADD CONSTRAINT check_deal_status 
    CHECK (status IN ('pending', 'published', 'active', 'expired', 'slashed', 'failed', 'cancelled'));
ALTER TABLE filecoin_deals ADD CONSTRAINT check_epochs 
    CHECK (end_epoch > start_epoch);

-- Drop trigger
DROP TRIGGER IF EXISTS update_filecoin_deals_updated_at ON filecoin_deals;

-- Drop indexes
DROP INDEX IF EXISTS idx_filecoin_deals_pin_request_id;
DROP INDEX IF EXISTS idx_filecoin_deals_deal_cid;
DROP INDEX IF EXISTS idx_filecoin_deals_miner_id;
DROP INDEX IF EXISTS idx_filecoin_deals_status;
DROP INDEX IF EXISTS idx_filecoin_deals_end_epoch;

-- Drop table
DROP TABLE IF EXISTS filecoin_deals;
