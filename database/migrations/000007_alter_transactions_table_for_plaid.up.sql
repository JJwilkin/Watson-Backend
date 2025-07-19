CREATE TABLE IF NOT EXISTS plaid_accounts (
    id VARCHAR PRIMARY KEY,
    available_balance DECIMAL(10, 2),
    current_balance DECIMAL(10, 2),
    account_limit DECIMAL(10, 2),
    currency VARCHAR,
    account_name VARCHAR,
    official_name VARCHAR,
    account_type VARCHAR,
    account_subtype VARCHAR,
    account_holder_category VARCHAR,
    access_token VARCHAR,
    plaid_item_id VARCHAR,
    plaid_account_type VARCHAR
);

-- Add columns if they don't exist
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'transactions' AND column_name = 'plaid_transaction_id') THEN
        ALTER TABLE transactions ADD COLUMN plaid_transaction_id VARCHAR;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'transactions' AND column_name = 'plaid_account_id') THEN
        ALTER TABLE transactions ADD COLUMN plaid_account_id VARCHAR;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'transactions' AND column_name = 'currency') THEN
        ALTER TABLE transactions ADD COLUMN currency VARCHAR DEFAULT 'USD';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'transactions' AND column_name = 'provider_type') THEN
        ALTER TABLE transactions ADD COLUMN provider_type VARCHAR DEFAULT 'teller';
    END IF;
END $$;

-- Backfill existing records with USD and teller provider
UPDATE transactions SET currency = 'USD' WHERE currency IS NULL;
UPDATE transactions SET provider_type = 'teller' WHERE provider_type IS NULL;

-- Add foreign key constraint if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'transactions_plaid_account_id_fkey') THEN
        ALTER TABLE transactions 
        ADD CONSTRAINT transactions_plaid_account_id_fkey 
        FOREIGN KEY (plaid_account_id) REFERENCES plaid_accounts(id);
    END IF;
END $$;
