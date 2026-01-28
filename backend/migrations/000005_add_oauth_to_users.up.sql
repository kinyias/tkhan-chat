-- Add OAuth fields to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_provider VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_id VARCHAR(255);
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;

-- Create index on oauth_provider and oauth_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_oauth ON users(oauth_provider, oauth_id);

-- Add unique constraint for oauth_provider and oauth_id combination
ALTER TABLE users ADD CONSTRAINT unique_oauth_provider_id UNIQUE (oauth_provider, oauth_id);
