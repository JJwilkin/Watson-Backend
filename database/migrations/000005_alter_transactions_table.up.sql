-- Drop existing table and recreate with Teller API schema
DROP TABLE IF EXISTS transactions CASCADE;

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id integer NOT NULL,
    teller_institution_id UUID NOT NULL,
    teller_account_id UUID NOT NULL,
    teller_transaction_id VARCHAR(255) UNIQUE NOT NULL,
    
    -- Core transaction fields from Teller API
    amount VARCHAR(50) NOT NULL, -- Amount as string from Teller API
    description VARCHAR(500) NOT NULL,
    date DATE NOT NULL,
    type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL, -- 'posted' or 'pending'
    running_balance VARCHAR(50), -- Can be null, string from Teller API
    
    -- Transaction details from Teller API
    processing_status VARCHAR(50), -- 'pending' or 'complete'
    category VARCHAR(100), -- Teller categories like 'groceries', 'dining', etc.
    counterparty_name VARCHAR(255),
    counterparty_type VARCHAR(50), -- 'organization' or 'person'
    
    -- Links from Teller API
    self_link TEXT,
    account_link TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraints
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (teller_institution_id) REFERENCES teller_institutions(id) ON DELETE CASCADE,
    FOREIGN KEY (teller_account_id) REFERENCES teller_accounts(id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_teller_account_id ON transactions(teller_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(date);
CREATE INDEX IF NOT EXISTS idx_transactions_user_date ON transactions(user_id, date);
CREATE INDEX IF NOT EXISTS idx_transactions_teller_transaction_id ON transactions(teller_transaction_id);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_category ON transactions(category);

-- Create trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_transactions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_transactions_updated_at 
    BEFORE UPDATE ON transactions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_transactions_updated_at();
