-- OrgMind Complete Schema (November 2025)
-- This is the final consolidated schema including all features
-- Use this for fresh database deployments

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- USERS TABLE
-- ============================================================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255), -- NULL for OAuth-only users
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    oauth_provider VARCHAR(50), -- 'google', 'okta', 'office365', NULL
    oauth_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_oauth ON users(oauth_provider, oauth_id);

-- ============================================================================
-- GRAPHS TABLE
-- ============================================================================
CREATE TABLE graphs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    zep_graph_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    document_count INTEGER DEFAULT 0,
    gemini_store_id VARCHAR(255), -- Gemini File Search store ID
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_graphs_creator_id ON graphs(creator_id);
CREATE INDEX idx_graphs_zep_graph_id ON graphs(zep_graph_id);
CREATE INDEX idx_graphs_gemini_store_id ON graphs(gemini_store_id);

-- ============================================================================
-- GRAPH MEMBERSHIPS TABLE (Many-to-Many)
-- ============================================================================
CREATE TABLE graph_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    graph_id UUID NOT NULL REFERENCES graphs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'member', -- 'owner', 'editor', 'viewer', 'member'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(graph_id, user_id)
);

CREATE INDEX idx_graph_memberships_graph_id ON graph_memberships(graph_id);
CREATE INDEX idx_graph_memberships_user_id ON graph_memberships(user_id);
CREATE INDEX idx_graph_memberships_lookup ON graph_memberships(user_id, graph_id);

-- ============================================================================
-- DOCUMENTS TABLE
-- ============================================================================
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    graph_id UUID REFERENCES graphs(id) ON DELETE CASCADE,
    filename VARCHAR(255),
    content_type VARCHAR(100),
    storage_key VARCHAR(500) NOT NULL, -- S3/GCS/MinIO key
    size_bytes BIGINT,
    source VARCHAR(50) NOT NULL, -- 'editor' or 'upload'
    status VARCHAR(50) DEFAULT 'processing', -- 'processing', 'completed', 'failed'
    error_message TEXT, -- Error details for failed documents
    gemini_file_id VARCHAR(255), -- Gemini File Search file ID
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_documents_user ON documents(user_id);
CREATE INDEX idx_documents_graph_id ON documents(graph_id);
CREATE INDEX idx_documents_status ON documents(status);
CREATE INDEX idx_documents_gemini_file_id ON documents(gemini_file_id);

-- ============================================================================
-- PASSWORD RESET TOKENS TABLE
-- ============================================================================
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_reset_tokens_token ON password_reset_tokens(token);

-- ============================================================================
-- CHAT THREADS TABLE
-- ============================================================================
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

-- ============================================================================
-- CHAT MESSAGES TABLE
-- ============================================================================
CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thread_id UUID NOT NULL REFERENCES chat_threads(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant')),
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_chat_messages_thread_id ON chat_messages(thread_id);
CREATE INDEX idx_chat_messages_created_at ON chat_messages(created_at ASC);

-- ============================================================================
-- GEMINI FILE SEARCH STORES TABLE
-- ============================================================================
CREATE TABLE gemini_filesearch_stores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_name VARCHAR(255) NOT NULL UNIQUE,
    store_id TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_gemini_filesearch_stores_store_name ON gemini_filesearch_stores(store_name);

-- Table comments
COMMENT ON TABLE gemini_filesearch_stores IS 'Stores Gemini File Search store information for persistence';
COMMENT ON COLUMN gemini_filesearch_stores.store_name IS 'Display name of the File Search store';
COMMENT ON COLUMN gemini_filesearch_stores.store_id IS 'Gemini API store ID';
