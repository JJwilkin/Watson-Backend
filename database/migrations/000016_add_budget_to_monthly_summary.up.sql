ALTER TABLE monthly_summary
    DROP COLUMN budget;

CREATE TABLE monthly_budget_spend_category (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id INT NOT NULL,
    monthly_summary_id INT NOT NULL,
    month_year INT NOT NULL,
    category VARCHAR(255) NOT NULL,
    budget DECIMAL(10, 2) NOT NULL,
    total_spent DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (monthly_summary_id) REFERENCES monthly_summary(id)
);

-- Create a function to automatically update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create a trigger to automatically update updated_at on row updates
CREATE TRIGGER update_monthly_budget_spend_category_updated_at 
    BEFORE UPDATE ON monthly_budget_spend_category 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();