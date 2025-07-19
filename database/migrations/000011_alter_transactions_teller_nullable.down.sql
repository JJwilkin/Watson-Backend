-- Note: This migration may fail if there are NULL values in these columns
-- You may need to clean up NULL values before running this migration

ALTER TABLE transactions
    ALTER COLUMN teller_transaction_id SET NOT NULL,
    ADD CONSTRAINT transactions_teller_transaction_id_key UNIQUE (teller_transaction_id),
    ALTER COLUMN teller_account_id SET NOT NULL,
    ALTER COLUMN teller_institution_id SET NOT NULL;
