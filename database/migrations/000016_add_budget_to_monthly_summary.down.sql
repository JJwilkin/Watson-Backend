-- Drop the trigger first
DROP TRIGGER IF EXISTS update_monthly_budget_spend_category_updated_at ON monthly_budget_spend_category;

-- Drop the function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop the table
DROP TABLE IF EXISTS monthly_budget_spend_category;

-- Add back the budget column to monthly_summary
ALTER TABLE monthly_summary
    ADD COLUMN budget JSONB;
