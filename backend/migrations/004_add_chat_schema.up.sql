-- Create chat_threads table
CREATE TABLE chat_threads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    graph_id UUID NOT NULL REFERENCES graphs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    summary VARCHAR(200),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_chat_threads_graph_id ON chat_threads(graph_id);
CREATE INDEX idx_chat_threads_user_id ON chat_threads(user_id);
CREATE INDEX idx_chat_threads_created_at ON chat_threads(created_at DESC);

-- Create chat_messages table
CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thread_id UUID NOT NULL REFERENCES chat_threads(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant')),
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_chat_messages_thread_id ON chat_messages(thread_id);
CREATE INDEX idx_chat_messages_created_at ON chat_messages(created_at ASC);

-- Add gemini_file_id column to documents table for Gemini File Search tracking
ALTER TABLE documents 
ADD COLUMN gemini_file_id VARCHAR(255);

CREATE INDEX idx_documents_gemini_file_id ON documents(gemini_file_id);

-- Add gemini_store_id column to graphs table for File Search store tracking
ALTER TABLE graphs 
ADD COLUMN gemini_store_id VARCHAR(255);

CREATE INDEX idx_graphs_gemini_store_id ON graphs(gemini_store_id);
