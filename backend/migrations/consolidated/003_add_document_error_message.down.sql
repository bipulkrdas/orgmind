-- Remove error_message column from documents table
DROP INDEX IF EXISTS idx_documents_status;
ALTER TABLE documents DROP COLUMN error_message;
