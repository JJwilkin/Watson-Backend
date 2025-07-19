ALTER TABLE plaid_accounts 
    ADD COLUMN plaid_token_id UUID;

ALTER TABLE plaid_accounts 
    ADD CONSTRAINT plaid_token_fk 
    FOREIGN KEY (plaid_token_id) 
    REFERENCES plaid_tokens(id);

ALTER TABLE plaid_accounts 
    DROP COLUMN access_token;

CREATE INDEX IF NOT EXISTS idx_plaid_accounts_plaid_token_id ON plaid_accounts(plaid_token_id);

