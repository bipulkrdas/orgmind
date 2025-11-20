-- Migration to convert all TIMESTAMP columns to TIMESTAMP WITH TIME ZONE
-- This ensures all timestamps are stored in UTC and eliminates timezone ambiguity
-- 
-- IMPORTANT: Existing timestamps are interpreted as America/New_York timezone
-- and will be converted to UTC for storage

-- Users table
ALTER TABLE users 
  ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE USING created_at AT TIME ZONE 'America/New_York',
  ALTER COLUMN updated_at TYPE TIMESTAMP WITH TIME ZONE USING updated_at AT TIME ZONE 'America/New_York';

-- Graphs table
ALTER TABLE graphs 
  ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE USING created_at AT TIME ZONE 'America/New_York',
  ALTER COLUMN updated_at TYPE TIMESTAMP WITH TIME ZONE USING updated_at AT TIME ZONE 'America/New_York';

-- Graph memberships table
ALTER TABLE graph_memberships 
  ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE USING created_at AT TIME ZONE 'America/New_York';

-- Documents table
ALTER TABLE documents 
  ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE USING created_at AT TIME ZONE 'America/New_York',
  ALTER COLUMN updated_at TYPE TIMESTAMP WITH TIME ZONE USING updated_at AT TIME ZONE 'America/New_York';

-- Password reset tokens table
ALTER TABLE password_reset_tokens 
  ALTER COLUMN expires_at TYPE TIMESTAMP WITH TIME ZONE USING expires_at AT TIME ZONE 'America/New_York',
  ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE USING created_at AT TIME ZONE 'America/New_York';

-- Chat threads table
ALTER TABLE chat_threads 
  ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE USING created_at AT TIME ZONE 'America/New_York',
  ALTER COLUMN updated_at TYPE TIMESTAMP WITH TIME ZONE USING updated_at AT TIME ZONE 'America/New_York';

-- Chat messages table
ALTER TABLE chat_messages 
  ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE USING created_at AT TIME ZONE 'America/New_York';
