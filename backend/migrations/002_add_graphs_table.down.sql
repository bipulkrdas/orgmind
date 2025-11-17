-- Remove graph_id column from documents table
DROP INDEX IF EXISTS idx_documents_graph_id;
ALTER TABLE documents DROP COLUMN IF EXISTS graph_id;

-- Drop graph_memberships table
DROP INDEX IF EXISTS idx_graph_memberships_lookup;
DROP INDEX IF EXISTS idx_graph_memberships_user_id;
DROP INDEX IF EXISTS idx_graph_memberships_graph_id;
DROP TABLE IF EXISTS graph_memberships;

-- Drop graphs table
DROP INDEX IF EXISTS idx_graphs_zep_graph_id;
DROP INDEX IF EXISTS idx_graphs_creator_id;
DROP TABLE IF EXISTS graphs;
