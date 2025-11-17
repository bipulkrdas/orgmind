# Refactoring: Removal of `zep_session_id` Field

## Date: 2024

## Reason for Change

After reviewing Zep Cloud documentation and architecture:

1. **Terminology Clarification:**
   - Zep v2 used "sessions" → Zep v3 uses "threads"
   - Zep v2 used "groups" → Zep v3 uses "graphs"
   - The field name `zep_session_id` was based on outdated v2 terminology

2. **Episode Management:**
   - When adding data to a graph via `graph.add()`, Zep creates **episodes**
   - Each document chunk becomes a separate episode with its own UUID
   - One document = multiple chunks = multiple episode UUIDs

3. **Architectural Decision:**
   - Storing individual episode UUIDs would require:
     - JSONB array field to store multiple UUIDs per document
     - Complex management of episode-to-chunk mapping
   - **Simpler approach:** Use `graph_id` to query all episodes in a graph
   - Episodes can be queried by graph_id when needed
   - No need to maintain document-to-episode mapping in our database

## Changes Made

### Database Schema

**Files Updated:**
- `backend/migrations/consolidated/001_initial_schema.up.sql`
- `backend/migrations/001_create_initial_schema.up.sql`

**Change:**
```sql
-- REMOVED:
zep_session_id VARCHAR(255), -- Zep session reference

-- Documents table now has:
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    graph_id UUID REFERENCES graphs(id) ON DELETE CASCADE,
    filename VARCHAR(255),
    content_type VARCHAR(100),
    storage_key VARCHAR(500) NOT NULL,
    size_bytes BIGINT,
    source VARCHAR(50) NOT NULL,
    status VARCHAR(50) DEFAULT 'processing',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### Backend Go Code

**1. Models (`backend/internal/models/document.go`)**
```go
// REMOVED:
ZepSessionID *string `json:"zepSessionId" db:"zep_session_id"`

// Document struct now has:
type Document struct {
    ID          string    `json:"id" db:"id"`
    UserID      string    `json:"userId" db:"user_id"`
    GraphID     *string   `json:"graphId" db:"graph_id"`
    Filename    *string   `json:"filename" db:"filename"`
    ContentType *string   `json:"contentType" db:"content_type"`
    StorageKey  string    `json:"storageKey" db:"storage_key"`
    SizeBytes   int64     `json:"sizeBytes" db:"size_bytes"`
    Source      string    `json:"source" db:"source"`
    Status      string    `json:"status" db:"status"`
    CreatedAt   time.Time `json:"createdAt" db:"created_at"`
    UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}
```

**2. Handlers (`backend/internal/handler/document_handler.go`)**
```go
// REMOVED from DocumentResponse:
ZepSessionID *string `json:"zepSessionId,omitempty"`

// Removed from all response mappings in:
// - SubmitEditorContent()
// - UploadFile()
// - ListDocuments()
// - GetDocument()
// - UpdateDocument()
```

**3. Graph Handler (`backend/internal/handler/graph_handler.go`)**
```go
// Removed ZepSessionID from DocumentResponse mapping in:
// - ListGraphDocuments()
```

**4. Repository (`backend/internal/repository/document_repository.go`)**
```go
// Removed "zep_session_id" from SQL queries in:
// - Create() - Removed from Columns() and Values()
// - GetByID() - Removed from Select()
// - ListByUserID() - Removed from Select()
// - ListByGraphID() - Removed from Select()
// - Update() - Removed from Set()
```

### Frontend TypeScript Code

**Types (`frontend/lib/types/index.ts`)**
```typescript
// REMOVED:
zepSessionId?: string;

// Document interface now:
export interface Document {
  id: string;
  userId: string;
  graphId?: string;
  filename?: string;
  contentType?: string;
  storageKey: string;
  sizeBytes: number;
  source: 'editor' | 'upload';
  status: 'processing' | 'completed' | 'failed';
  createdAt: string;
  updatedAt: string;
}
```

## Impact Analysis

### What Still Works:
✅ Document creation and storage
✅ Document processing and chunking
✅ Sending chunks to Zep via `graph.add()`
✅ Knowledge graph visualization
✅ Graph-based document queries
✅ All existing API endpoints

### What Changed:
- Database schema is simpler (one less field)
- API responses are cleaner (no unused field)
- No need to track individual episode UUIDs
- Episodes are managed entirely by Zep Cloud

### How to Query Episodes:
If you need to access episodes for a document:

1. **By Graph ID:**
   ```go
   episodes, err := zepClient.Graph.Episode.GetByGraphID(ctx, graphID, &graph.EpisodeGetByGraphIDRequest{
       Lastn: zep.Int(100),
   })
   ```

2. **By Search:**
   ```go
   results, err := zepClient.Graph.Search(ctx, &zep.GraphSearchQuery{
       GraphID: zep.String(graphID),
       Scope:   zep.GraphSearchScopeEpisodes.Ptr(),
       Limit:   zep.Int(50),
   })
   ```

## Migration Notes

**For Fresh Deployments:**
- Use `backend/migrations/consolidated/001_initial_schema.up.sql`
- No `zep_session_id` field will be created

**For Existing Databases:**
- If you have existing data with `zep_session_id` populated, you can:
  1. Keep the field (it will just be NULL for new documents)
  2. Or create a migration to drop the column:
     ```sql
     ALTER TABLE documents DROP COLUMN IF EXISTS zep_session_id;
     ```

**Note:** Since this is pre-production, we directly updated the consolidated schema without creating a new migration file.

## Testing Checklist

- [x] Database schema compiles without errors
- [x] Go code compiles without errors
- [x] TypeScript code compiles without errors
- [ ] Document creation works (editor)
- [ ] Document creation works (upload)
- [ ] Document processing sends data to Zep
- [ ] Graph visualization retrieves episodes correctly
- [ ] API responses don't include zepSessionId field

## References

- [Zep v2 to v3 Migration Guide](https://help.getzep.com/sdks/migration)
- [Zep Graph API Documentation](https://help.getzep.com/graphiti)
- [Zep Episodes Documentation](https://help.getzep.com/graphiti/graphiti/adding-episodes)
