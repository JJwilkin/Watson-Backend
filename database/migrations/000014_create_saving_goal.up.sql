CREATE TABLE IF NOT EXISTS saving_goal (
    id serial PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    total DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    currently_saved DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    redeemed BOOLEAN NOT NULL DEFAULT FALSE,
    transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on user_id for efficient lookups
CREATE INDEX idx_saving_goal_user_id ON saving_goal(user_id);

-- Create index on redeemed for filtering completed goals
CREATE INDEX idx_saving_goal_redeemed ON saving_goal(redeemed);

-- Create index on transaction_id for transaction lookups
CREATE INDEX idx_saving_goal_transaction_id ON saving_goal(transaction_id);
