-- Drop the GIN index first
DROP INDEX IF EXISTS idx_transactions_category_gin;

-- Convert JSONB back to text, extracting the string value
ALTER TABLE transactions
    ALTER COLUMN category TYPE TEXT USING category::text;
