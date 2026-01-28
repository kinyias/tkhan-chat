-- Remove OAuth fields from users table
ALTER TABLE users DROP CONSTRAINT IF EXISTS unique_oauth_provider_id;
DROP INDEX IF EXISTS idx_users_oauth;
ALTER TABLE users DROP COLUMN IF EXISTS oauth_id;
ALTER TABLE users DROP COLUMN IF EXISTS oauth_provider;
ALTER TABLE users ALTER COLUMN password SET NOT NULL;
