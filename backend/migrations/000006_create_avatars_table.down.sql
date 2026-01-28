-- Drop avatars table
DROP TABLE IF EXISTS avatars;

-- Add avatar column back to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar VARCHAR(255);
