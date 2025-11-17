-- Create graphs table
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

-- Create graph_memberships table (many-to-many relationship)
CREATE TABLE graph_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    graph_id UUID NOT NULL REFERENCES graphs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'member',
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(graph_id, user_id)
);

CREATE INDEX idx_graph_memberships_graph_id ON graph_memberships(graph_id);
CREATE INDEX idx_graph_memberships_user_id ON graph_memberships(user_id);
CREATE INDEX idx_graph_memberships_lookup ON graph_memberships(user_id, graph_id);

-- Add graph_id column to documents table
ALTER TABLE documents 
ADD COLUMN graph_id UUID REFERENCES graphs(id) ON DELETE CASCADE;

CREATE INDEX idx_documents_graph_id ON documents(graph_id);

-- Note: graph_id will be made NOT NULL after data migration
-- ALTER TABLE documents ALTER COLUMN graph_id SET NOT NULL;
