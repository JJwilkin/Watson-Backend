ALTER TABLE transactions
    ALTER COLUMN teller_transaction_id DROP NOT NULL,
    ALTER COLUMN teller_account_id DROP NOT NULL,
    ALTER COLUMN teller_institution_id DROP NOT NULL;

-- Drop the unique constraint separately
ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS transactions_teller_transaction_id_key;