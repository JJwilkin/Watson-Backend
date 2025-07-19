-- 1. Drop new foreign key constraint
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_teller_account_id_fkey;

-- 2. Add teller_account_id column back
ALTER TABLE teller_accounts ADD COLUMN teller_account_id UUID;

-- 3. (Optional) Set id column back to UUID if needed
-- ALTER TABLE teller_accounts ALTER COLUMN id TYPE UUID USING id::uuid;

-- 4. Add old foreign key constraint back
ALTER TABLE transactions
    ALTER COLUMN teller_account_id TYPE UUID USING teller_account_id::uuid,
    ADD CONSTRAINT transactions_teller_account_id_fkey
        FOREIGN KEY (teller_account_id) REFERENCES teller_accounts(id) ON DELETE CASCADE;
