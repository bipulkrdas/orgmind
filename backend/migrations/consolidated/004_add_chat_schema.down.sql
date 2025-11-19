-- Remove gemini_store_id column from graphs table
DROP INDEX IF EXISTS idx_graphs_gemini_store_id;
ALTER TABLE graphs DROP COLUMN IF EXISTS gemini_store_id;

-- Remove gemini_file_id column from documents table
DROP INDEX IF EXISTS idx_documents_gemini_file_id;
ALTER TABLE documents DROP COLUMN IF EXISTS gemini_file_id;

-- Drop chat_messages table
DROP INDEX IF EXISTS idx_chat_messages_created_at;
DROP INDEX IF EXISTS idx_chat_messages_thread_id;
DROP TABLE IF EXISTS chat_messages;

-- Drop chat_threads table
DROP INDEX IF EXISTS idx_chat_threads_created_at;
DROP INDEX IF EXISTS idx_chat_threads_user_id;
DROP INDEX IF EXISTS idx_chat_threads_graph_id;
DROP TABLE IF EXISTS chat_threads;
