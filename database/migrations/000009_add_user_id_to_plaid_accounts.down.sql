-- Drop foreign key constraint first
ALTER TABLE plaid_accounts
    DROP CONSTRAINT IF EXISTS plaid_accounts_user_id_fkey;

-- Drop user_id column
ALTER TABLE plaid_accounts
    DROP COLUMN IF EXISTS user_id;
