-- Create avatars table
CREATE TABLE IF NOT EXISTS avatars (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE,
    public_id VARCHAR(255) NOT NULL,
    public_url TEXT NOT NULL,
    secure_url TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index on user_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_avatars_user_id ON avatars(user_id);

-- Remove avatar column from users table (if exists)
ALTER TABLE users DROP COLUMN IF EXISTS avatar;
