-- Drop the index first
DROP INDEX IF EXISTS idx_plaid_accounts_plaid_token_id;

-- Add back the access_token column
ALTER TABLE plaid_accounts 
    ADD COLUMN access_token VARCHAR;

-- Drop the foreign key constraint
ALTER TABLE plaid_accounts 
    DROP CONSTRAINT IF EXISTS plaid_token_fk;

-- Drop the plaid_token_id column
ALTER TABLE plaid_accounts 
    DROP COLUMN IF EXISTS plaid_token_id;
    