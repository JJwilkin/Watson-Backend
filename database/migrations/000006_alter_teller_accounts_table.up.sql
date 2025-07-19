-- 1. Drop old foreign key constraint in transactions
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_teller_account_id_fkey;

-- 2. Drop the teller_account_id column from teller_accounts
ALTER TABLE teller_accounts DROP COLUMN IF EXISTS teller_account_id;

-- 3. Change the id column to VARCHAR and set as PRIMARY KEY (if not already)
ALTER TABLE teller_accounts
    ALTER COLUMN id TYPE VARCHAR,
    ALTER COLUMN id SET NOT NULL;

-- 4. Add new foreign key constraint in transactions to reference teller_accounts(id)
ALTER TABLE transactions
    ALTER COLUMN teller_account_id TYPE VARCHAR,
    ADD CONSTRAINT transactions_teller_account_id_fkey
        FOREIGN KEY (teller_account_id) REFERENCES teller_accounts(id) ON DELETE CASCADE;