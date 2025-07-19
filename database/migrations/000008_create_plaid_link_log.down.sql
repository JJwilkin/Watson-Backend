-- Drop trigger and function
DROP TRIGGER IF EXISTS plaid_tokens_updated_at_trigger ON plaid_tokens;
DROP FUNCTION IF EXISTS update_plaid_tokens_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_plaid_tokens_user_id;
DROP INDEX IF EXISTS idx_plaid_tokens_item_id;
DROP INDEX IF EXISTS idx_plaid_tokens_is_processed;
DROP INDEX IF EXISTS idx_plaid_tokens_created_at;

-- Drop table
DROP TABLE IF EXISTS plaid_tokens;
