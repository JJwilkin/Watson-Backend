-- Drop foreign key constraint if it exists
ALTER TABLE transactions 
    DROP CONSTRAINT IF EXISTS transactions_plaid_account_id_fkey;

-- Drop columns if they exist
ALTER TABLE transactions 
    DROP COLUMN IF EXISTS plaid_transaction_id,
    DROP COLUMN IF EXISTS plaid_account_id,
    DROP COLUMN IF EXISTS currency,
    DROP COLUMN IF EXISTS provider_type;

-- Drop table if it exists
DROP TABLE IF EXISTS plaid_accounts;