CREATE TABLE IF NOT EXISTS monthly_summary (
    id serial PRIMARY KEY,
    total_spent DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    budget JSONB,
    starting_balance DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    monthyear INTEGER NOT NULL,
    income DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    saved_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    invested DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, monthyear)
);

-- Create index on user_id and monthyear for efficient lookups
CREATE INDEX idx_monthly_summary_user_month ON monthly_summary(user_id, monthyear);

-- Create index on monthyear for date-based queries
CREATE INDEX idx_monthly_summary_monthyear ON monthly_summary(monthyear);
