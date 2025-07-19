CREATE TABLE IF NOT EXISTS teller_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL,
    teller_institution_id UUID NOT NULL,
    
    -- Teller API account properties
    teller_account_id VARCHAR(255) NOT NULL UNIQUE,
    enrollment_id VARCHAR(255) NOT NULL,
    account_name VARCHAR(255) NOT NULL,
    account_type VARCHAR(50) NOT NULL CHECK (account_type IN ('depository', 'credit')),
    account_subtype VARCHAR(50) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    last_four VARCHAR(4) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'closed')),
    
    -- Institution information
    institution_id VARCHAR(255) NOT NULL,
    institution_name VARCHAR(255) NOT NULL,
    
    -- Links for API access
    self_link TEXT,
    details_link TEXT,
    balances_link TEXT,
    transactions_link TEXT,
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraints
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (teller_institution_id) REFERENCES teller_institutions(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX idx_teller_accounts_user_id ON teller_accounts(user_id);
CREATE INDEX idx_teller_accounts_enrollment_id ON teller_accounts(enrollment_id);
CREATE INDEX idx_teller_accounts_teller_account_id ON teller_accounts(teller_account_id);
CREATE INDEX idx_teller_accounts_status ON teller_accounts(status);