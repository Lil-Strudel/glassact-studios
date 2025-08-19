DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_accounts_updated_at ON accounts;

DROP FUNCTION IF EXISTS update_updated_at_column();
