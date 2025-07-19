-- Drop indexes first
DROP INDEX IF EXISTS idx_teller_accounts_user_id;
DROP INDEX IF EXISTS idx_teller_accounts_enrollment_id;
DROP INDEX IF EXISTS idx_teller_accounts_teller_account_id;
DROP INDEX IF EXISTS idx_teller_accounts_status;

-- Drop the table
DROP TABLE IF EXISTS teller_accounts;
