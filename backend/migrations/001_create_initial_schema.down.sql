-- Drop tables in reverse order to handle foreign key constraints
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS users;

-- Drop UUID extension
DROP EXTENSION IF EXISTS "pgcrypto";
