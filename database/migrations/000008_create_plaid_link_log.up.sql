CREATE TABLE IF NOT EXISTS plaid_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL,
    access_token VARCHAR NOT NULL,
    item_id VARCHAR NOT NULL UNIQUE,
    is_processed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraint
    CONSTRAINT plaid_tokens_user_id_fkey 
        FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_plaid_tokens_user_id ON plaid_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_plaid_tokens_item_id ON plaid_tokens(item_id);
CREATE INDEX IF NOT EXISTS idx_plaid_tokens_is_processed ON plaid_tokens(is_processed);
CREATE INDEX IF NOT EXISTS idx_plaid_tokens_created_at ON plaid_tokens(created_at);

-- Create trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_plaid_tokens_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER plaid_tokens_updated_at_trigger
    BEFORE UPDATE ON plaid_tokens
    FOR EACH ROW
    EXECUTE FUNCTION update_plaid_tokens_updated_at();
