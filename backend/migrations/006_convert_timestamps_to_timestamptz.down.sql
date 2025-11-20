-- Rollback migration: convert TIMESTAMP WITH TIME ZONE back to TIMESTAMP
-- Note: This will lose timezone information

-- Users table
ALTER TABLE users 
  ALTER COLUMN created_at TYPE TIMESTAMP,
  ALTER COLUMN updated_at TYPE TIMESTAMP;

-- Graphs table
ALTER TABLE graphs 
  ALTER COLUMN created_at TYPE TIMESTAMP,
  ALTER COLUMN updated_at TYPE TIMESTAMP;

-- Graph memberships table
ALTER TABLE graph_memberships 
  ALTER COLUMN created_at TYPE TIMESTAMP;

-- Documents table
ALTER TABLE documents 
  ALTER COLUMN created_at TYPE TIMESTAMP,
  ALTER COLUMN updated_at TYPE TIMESTAMP;

-- Password reset tokens table
ALTER TABLE password_reset_tokens 
  ALTER COLUMN expires_at TYPE TIMESTAMP,
  ALTER COLUMN created_at TYPE TIMESTAMP;

-- Chat threads table
ALTER TABLE chat_threads 
  ALTER COLUMN created_at TYPE TIMESTAMP,
  ALTER COLUMN updated_at TYPE TIMESTAMP;

-- Chat messages table
ALTER TABLE chat_messages 
  ALTER COLUMN created_at TYPE TIMESTAMP;
