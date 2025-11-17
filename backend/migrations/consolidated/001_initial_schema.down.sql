-- OrgMind Schema Teardown
-- This drops all tables in the correct order to handle foreign key constraints

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS graph_memberships;
DROP TABLE IF EXISTS graphs;
DROP TABLE IF EXISTS users;

-- Drop UUID extension
DROP EXTENSION IF EXISTS "pgcrypto";
