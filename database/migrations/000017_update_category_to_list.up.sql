-- Create a temporary column to store the converted data
ALTER TABLE transactions ADD COLUMN category_new JSONB;

-- Convert existing data safely - singular values become arrays with one element
UPDATE transactions 
SET category_new = CASE 
    WHEN category IS NULL OR category = '' THEN '[]'::jsonb
    WHEN category = 'null' THEN 'null'::jsonb
    WHEN category LIKE '["%' AND category LIKE '%"]' THEN category::jsonb
    WHEN category LIKE '"%' AND category LIKE '%"' THEN ('[' || category || ']')::jsonb
    ELSE ('["' || REPLACE(REPLACE(category, '"', '\"'), '\', '\\') || '"]')::jsonb
END;

-- Drop the old column and rename the new one
ALTER TABLE transactions DROP COLUMN category;
ALTER TABLE transactions RENAME COLUMN category_new TO category;

-- Create the GIN index for faster queries
CREATE INDEX idx_transactions_category_gin ON transactions USING GIN (category);