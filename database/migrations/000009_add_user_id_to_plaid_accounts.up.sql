-- Add user_id column to plaid_accounts table
ALTER TABLE plaid_accounts
    ADD COLUMN user_id INTEGER;

-- Add foreign key constraint
ALTER TABLE plaid_accounts
    ADD CONSTRAINT plaid_accounts_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;

ALTER TABLE plaid_tokens
    ADD CONSTRAINT plaid_tokens_access_token_unique UNIQUE (access_token);