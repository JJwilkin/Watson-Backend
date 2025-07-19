CREATE TABLE IF NOT EXISTS monthly_balance (
    id serial PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    current_balance DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    total_owing DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    net_cash DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    available_balance DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    monthyear INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, monthyear)
);

-- Create index on user_id and monthyear for efficient lookups
CREATE INDEX idx_monthly_balance_user_month ON monthly_balance(user_id, monthyear);

-- Create index on monthyear for date-based queries
CREATE INDEX idx_monthly_balance_monthyear ON monthly_balance(monthyear);
