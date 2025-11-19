-- Create gemini_filesearch_stores table
CREATE TABLE IF NOT EXISTS gemini_filesearch_stores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_name VARCHAR(255) NOT NULL UNIQUE,
    store_id TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on store_name for faster lookups
CREATE INDEX idx_gemini_filesearch_stores_store_name ON gemini_filesearch_stores(store_name);

-- Add comment to table
COMMENT ON TABLE gemini_filesearch_stores IS 'Stores Gemini File Search store information for persistence';
COMMENT ON COLUMN gemini_filesearch_stores.store_name IS 'Display name of the File Search store';
COMMENT ON COLUMN gemini_filesearch_stores.store_id IS 'Gemini API store ID';
