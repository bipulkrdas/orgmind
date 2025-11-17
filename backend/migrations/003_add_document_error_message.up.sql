-- Add error_message column to documents table for storing extraction error details
ALTER TABLE documents ADD COLUMN error_message TEXT;

-- Add index on status for efficient querying of failed documents
CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status);
