-- Drop the transactions table and all its dependencies
DROP TABLE IF EXISTS transactions CASCADE;

-- Drop the trigger function if it's no longer used by other tables
DROP FUNCTION IF EXISTS update_transactions_updated_at() CASCADE;
