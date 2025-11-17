-- OrgMind Initial Schema
-- This is a consolidated schema that includes all tables for fresh deployments
-- For existing databases, use the incremental migrations in the parent directory

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
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
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
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_graphs_creator_id ON graphs(creator_id);
CREATE INDEX idx_graphs_zep_graph_id ON graphs(zep_graph_id);

-- ============================================================================
-- GRAPH MEMBERSHIPS TABLE (Many-to-Many)
-- ============================================================================
CREATE TABLE graph_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    graph_id UUID NOT NULL REFERENCES graphs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'member', -- 'owner', 'editor', 'viewer', 'member'
    created_at TIMESTAMP DEFAULT NOW(),
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
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_documents_user ON documents(user_id);
CREATE INDEX idx_documents_graph_id ON documents(graph_id);
CREATE INDEX idx_documents_status ON documents(status);

-- ============================================================================
-- PASSWORD RESET TOKENS TABLE
-- ============================================================================
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_reset_tokens_token ON password_reset_tokens(token);
