# Design Document: Gemini Shared File Search Store

## Overview

This design refactors the Gemini File Search integration from a per-graph store model to a single shared store with metadata-based filtering. This approach provides better scalability, reduces API overhead, and simplifies multi-tenant data isolation.

## Architecture

### Current Architecture (Per-Graph Stores)

```
Graph A → File Search Store A → Documents for Graph A
Graph B → File Search Store B → Documents for Graph B
Graph C → File Search Store C → Documents for Graph C
```

**Problems:**
- N stores for N graphs (API overhead)
- Store creation on first document upload (race conditions)
- Complex store lifecycle management
- Difficult to query across graphs (if needed)

### New Architecture (Shared Store with Metadata)

```
All Graphs → Single File Search Store → All Documents (filtered by metadata)
                                      ↓
                        [graph_id=A, domain=topeic.com, version=1.1]
                        [graph_id=B, domain=topeic.com, version=1.1]
                        [graph_id=C, domain=topeic.com, version=1.1]
```

**Benefits:**
- Single store for all documents
- Store created once at startup
- Metadata-based isolation
- Simpler lifecycle management
- Future: Cross-graph queries possible

## Database Schema

### gemini_file_search_stores Table

```sql
CREATE TABLE gemini_file_search_stores (
    id VARCHAR(255) PRIMARY KEY,              -- Our identifier (e.g., "graph_search_gemini_store")
    store_name VARCHAR(500) NOT NULL,         -- Gemini's store.Name (returned from API)
    display_name VARCHAR(255) NOT NULL,       -- Human-readable name
    purpose VARCHAR(255),                     -- Purpose description (e.g., "Graph document search")
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, inactive, deleted
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,                           -- Additional metadata (project_id, location, etc.)
    
    CONSTRAINT unique_store_name UNIQUE (store_name)
);

-- Index for quick lookup by our identifier
CREATE INDEX idx_gemini_stores_id ON gemini_file_search_stores(id);

-- Index for status queries
CREATE INDEX idx_gemini_stores_status ON gemini_file_search_stores(status);
```

**Example Record:**
```json
{
  "id": "graph_search_gemini_store",
  "store_name": "projects/123/locations/us-central1/fileSearchStores/abc123",
  "display_name": "OrgMind Documents",
  "purpose": "Primary store for all graph documents with metadata filtering",
  "status": "active",
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:00Z",
  "metadata": {
    "project_id": "my-project",
    "location": "us-central1",
    "version": "1.0"
  }
}
```

## Components and Interfaces

### 1. Configuration (config.go)

**New Fields:**
```go
type Config struct {
    // ... existing fields ...
    
    // Gemini File Search Store
    GeminiStoreIdentifier string // From GEMINI_STORE_ID env var (e.g., "graph_search_gemini_store")
    GeminiStoreDisplayName string // From GEMINI_STORE_DISPLAY_NAME env var (e.g., "OrgMind Documents")
}
```

**Environment Variables:**
```bash
# Unique identifier for the store (used as key in database)
GEMINI_STORE_ID=graph_search_gemini_store

# Human-readable display name for the store
GEMINI_STORE_DISPLAY_NAME=OrgMind Documents
```

### 2. Models (models/gemini_store.go)

**New Model:**
```go
package models

import "time"

// GeminiFileSearchStore represents a Gemini File Search store record
type GeminiFileSearchStore struct {
    ID          string                 `db:"id" json:"id"`                   // Our identifier
    StoreName   string                 `db:"store_name" json:"storeName"`    // Gemini's store.Name
    DisplayName string                 `db:"display_name" json:"displayName"` // Human-readable name
    Purpose     *string                `db:"purpose" json:"purpose,omitempty"`
    Status      string                 `db:"status" json:"status"`           // active, inactive, deleted
    CreatedAt   time.Time              `db:"created_at" json:"createdAt"`
    UpdatedAt   time.Time              `db:"updated_at" json:"updatedAt"`
    Metadata    map[string]interface{} `db:"metadata" json:"metadata,omitempty"`
}

// Validate validates the store record
func (s *GeminiFileSearchStore) Validate() error {
    if s.ID == "" {
        return fmt.Errorf("store ID is required")
    }
    if s.StoreName == "" {
        return fmt.Errorf("store name is required")
    }
    if s.DisplayName == "" {
        return fmt.Errorf("display name is required")
    }
    if s.Status == "" {
        s.Status = "active"
    }
    return nil
}
```

### 3. Repository Interface (repository/interfaces.go)

**New Interface:**
```go
// GeminiStoreRepository defines the interface for Gemini store data access
type GeminiStoreRepository interface {
    // Create a new store record
    Create(ctx context.Context, store *models.GeminiFileSearchStore) error
    
    // Get store by ID (our identifier)
    GetByID(ctx context.Context, id string) (*models.GeminiFileSearchStore, error)
    
    // Get store by Gemini store name
    GetByStoreName(ctx context.Context, storeName string) (*models.GeminiFileSearchStore, error)
    
    // Update store record
    Update(ctx context.Context, store *models.GeminiFileSearchStore) error
    
    // List all active stores
    ListActive(ctx context.Context) ([]*models.GeminiFileSearchStore, error)
    
    // Update store status
    UpdateStatus(ctx context.Context, id, status string) error
}
```

### 4. GeminiService Interface (service/interfaces.go)

**Updated Interface:**
```go
type GeminiService interface {
    // Store management (called once at startup)
    InitializeStore(ctx context.Context, storeID, displayName string) (*models.GeminiFileSearchStore, error)
    
    // Get store information
    GetStore(ctx context.Context, storeID string) (*models.GeminiFileSearchStore, error)
    
    // Document management (uses shared store)
    UploadDocument(ctx context.Context, storeName, graphID, graphName, documentID string, content []byte, mimeType string) (string, error)
    
    // Chat interaction (with metadata filtering)
    GenerateStreamingResponse(ctx context.Context, storeName, graphID, domain, version, query string, responseChan chan<- string) error
    
    // Removed methods:
    // - CreateFileSearchStore() - replaced by InitializeStore()
    // - GetFileSearchStore() - replaced by GetStore()
    // - DeleteFileSearchStore() - use UpdateStatus instead
}
```

### 5. GeminiService Implementation (service/gemini_service.go)

**Updated Structure:**
```go
type geminiService struct {
    client     *genai.Client
    storeRepo  repository.GeminiStoreRepository
    graphRepo  repository.GraphRepository
    docRepo    repository.DocumentRepository
    apiKey     string
    projectID  string
    location   string
    
    // Removed:
    // - storeCache sync.Map (database is the source of truth)
    // - storeID, storeName fields (retrieved from database)
}
```

**Constructor:**
```go
func NewGeminiService(
    apiKey, projectID, location string,
    storeRepo repository.GeminiStoreRepository,
    graphRepo repository.GraphRepository,
    docRepo repository.DocumentRepository,
) (GeminiService, error)
```

**Key Methods:**

```go
// InitializeStore gets or creates the shared File Search store
// Called once during app startup in main.go
func (s *geminiService) InitializeStore(
    ctx context.Context,
    storeID, displayName string,
) (*models.GeminiFileSearchStore, error) {
    // 1. Check if store exists in database
    existingStore, err := s.storeRepo.GetByID(ctx, storeID)
    if err == nil {
        // Store exists, verify it's still valid in Gemini
        log.Printf("[Gemini] Found existing store: %s", existingStore.StoreName)
        return existingStore, nil
    }
    
    // 2. Store doesn't exist, create new one in Gemini
    log.Printf("[Gemini] Creating new File Search store: %s", displayName)
    geminiStore, err := s.client.FileSearchStores.Create(ctx, &genai.CreateFileSearchStoreConfig{
        DisplayName: displayName,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create store: %w", err)
    }
    
    // 3. Save to database
    store := &models.GeminiFileSearchStore{
        ID:          storeID,
        StoreName:   geminiStore.Name,
        DisplayName: displayName,
        Purpose:     "Primary store for all graph documents with metadata filtering",
        Status:      "active",
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
        Metadata: map[string]interface{}{
            "project_id": s.projectID,
            "location":   s.location,
            "version":    "1.0",
        },
    }
    
    if err := s.storeRepo.Create(ctx, store); err != nil {
        return nil, fmt.Errorf("failed to save store to database: %w", err)
    }
    
    return store, nil
}

// GetStore retrieves store information from database
func (s *geminiService) GetStore(ctx context.Context, storeID string) (*models.GeminiFileSearchStore, error) {
    return s.storeRepo.GetByID(ctx, storeID)
}

// UploadDocument uploads with custom metadata
func (s *geminiService) UploadDocument(
    ctx context.Context,
    storeName, graphID, graphName, documentID string,
    content []byte,
    mimeType string,
) (string, error) {
    // Upload with metadata:
    // - display_name: graphName
    // - custom_metadata:
    //   - graph_id: graphID
    //   - domain: "topeic.com"
    //   - version: "1.1"
    
    reader := bytes.NewReader(content)
    op, err := s.client.FileSearchStores.UploadToFileSearchStore(
        ctx,
        reader,
        storeName,
        &genai.UploadToFileSearchStoreConfig{
            DisplayName: graphName,
            MIMEType:    mimeType,
            CustomMetadata: map[string]string{
                "graph_id": graphID,
                "domain":   "topeic.com",
                "version":  "1.1",
            },
        },
    )
    // ... rest of implementation
}

// GenerateStreamingResponse with metadata filtering
func (s *geminiService) GenerateStreamingResponse(
    ctx context.Context,
    storeName, graphID, domain, version, query string,
    responseChan chan<- string,
) error {
    // Build metadata filter:
    filter := fmt.Sprintf(
        `(chunk.custom_metadata.graph_id = "%s" AND chunk.custom_metadata.domain = "%s" AND chunk.custom_metadata.version = "%s")`,
        graphID, domain, version,
    )
    
    config := &genai.GenerateContentConfig{
        Tools: []*genai.Tool{
            {
                FileSearch: &genai.FileSearch{
                    FileSearchStoreNames: []string{storeName},
                    MetadataFilter:       filter,
                },
            },
        },
    }
    // ... rest of implementation
}
```

### 6. Document Service (service/document_service.go)

**Updated uploadToFileSearch:**
```go
func (s *documentService) uploadToFileSearch(
    ctx context.Context,
    graphID, graphName, documentID, content, mimeType string,
) {
    // Get store from database
    store, err := s.geminiService.GetStore(ctx, "graph_search_gemini_store")
    if err != nil {
        log.Printf("[FileSearch] Failed to get store: %v", err)
        return
    }
    
    // Upload with store name
    fileID, err := s.geminiService.UploadDocument(
        ctx,
        store.StoreName,  // Use Gemini's store.Name
        graphID,
        graphName,
        documentID,
        []byte(content),
        mimeType,
    )
}
```

### 7. Chat Service (service/chat_service.go)

**Updated GenerateResponseForMessage:**
```go
func (s *chatService) GenerateResponseForMessage(
    ctx context.Context,
    threadID, userMessageID, graphID string,
    responseChan chan<- string,
) (string, error) {
    // Get store from database
    store, err := s.geminiSvc.GetStore(ctx, "graph_search_gemini_store")
    if err != nil {
        return "", fmt.Errorf("failed to get store: %w", err)
    }
    
    // Get graph for name
    graph, err := s.graphRepo.GetByID(ctx, graphID)
    
    // Call Gemini with metadata filtering
    err = s.geminiSvc.GenerateStreamingResponse(
        ctx,
        store.StoreName,  // Use Gemini's store.Name
        graphID,
        "topeic.com",  // Hardcoded for now
        "1.1",         // Version
        userMsg.Content,
        fullResponseChan,
    )
}
```

### 8. Main Application (cmd/server/main.go)

**Startup Sequence:**
```go
func main() {
    // ... existing initialization ...
    
    // Initialize repositories
    userRepo := repository.NewUserRepository(db.DB)
    documentRepo := repository.NewDocumentRepository(db.DB)
    graphRepo := repository.NewGraphRepository(db.DB)
    geminiStoreRepo := repository.NewGeminiStoreRepository(db.DB)  // NEW
    
    // Initialize Gemini service
    geminiService, err := service.NewGeminiService(
        cfg.GeminiAPIKey,
        cfg.GeminiProject,
        cfg.GeminiLocation,
        geminiStoreRepo,  // NEW
        graphRepo,
        documentRepo,
    )
    
    // Initialize File Search store (get or create)
    if geminiService != nil {
        log.Println("Initializing Gemini File Search store...")
        store, err := geminiService.InitializeStore(
            ctx,
            cfg.GeminiStoreIdentifier,    // e.g., "graph_search_gemini_store"
            cfg.GeminiStoreDisplayName,   // e.g., "OrgMind Documents"
        )
        if err != nil {
            log.Fatalf("Failed to initialize File Search store: %v", err)
        }
        log.Printf("File Search store initialized: %s (Gemini name: %s)", 
            store.ID, store.StoreName)
    }
    
    // ... continue with rest of initialization ...
}
```

## Data Models

### Document Metadata Structure

```json
{
  "display_name": "My Knowledge Graph",
  "custom_metadata": {
    "graph_id": "graph_abc123",
    "domain": "topeic.com",
    "version": "1.1"
  }
}
```

### Metadata Filter Expression

```
(chunk.custom_metadata.graph_id = "graph_abc123" AND 
 chunk.custom_metadata.domain = "topeic.com" AND 
 chunk.custom_metadata.version = "1.1")
```

## Error Handling

### Store Initialization Errors

**Scenario:** Store creation fails at startup

**Handling:**
1. Log detailed error with retry information
2. Exit application with non-zero status
3. Provide clear error message for administrator

**Example:**
```
FATAL: Failed to initialize Gemini File Search store after 3 attempts: API key invalid
Please check GEMINI_API_KEY configuration and try again
```

### Document Upload Errors

**Scenario:** Upload fails with metadata

**Handling:**
1. Log error with graph_id and document_id
2. Continue with Zep processing (don't fail document creation)
3. Mark document as "partial" in logs

**Example:**
```
WARNING: Failed to upload document doc_123 to File Search for graph graph_abc: metadata validation failed
Document will be available in Zep but not in AI chat
```

### Query Filtering Errors

**Scenario:** Metadata filter syntax error

**Handling:**
1. Log filter expression that failed
2. Return error to user via SSE error event
3. Don't retry (syntax errors won't resolve)

**Example:**
```
ERROR: Invalid metadata filter syntax: (chunk.custom_metadata.graph_id = "graph_abc"
Please contact support if this persists
```

## Testing Strategy

### Unit Tests

**GeminiService:**
- `TestInitializeStore` - Store creation with retry
- `TestUploadDocumentWithMetadata` - Metadata attachment
- `TestGenerateStreamingResponseWithFilter` - Filter construction
- `TestMetadataFilterSyntax` - Filter expression validation

**DocumentService:**
- `TestUploadToFileSearchSharedStore` - Uses shared store
- `TestUploadToFileSearchMetadata` - Correct metadata values

**ChatService:**
- `TestGenerateResponseWithGraphFilter` - Graph isolation
- `TestGenerateResponseWithDomainFilter` - Domain filtering

### Integration Tests

**Store Initialization:**
```go
func TestStoreInitializationAtStartup(t *testing.T) {
    // Start app
    // Verify store created
    // Verify store ID in config
    // Verify subsequent uploads use same store
}
```

**Multi-Graph Isolation:**
```go
func TestMultiGraphIsolation(t *testing.T) {
    // Upload docs to graph A
    // Upload docs to graph B
    // Query graph A - should only see A's docs
    // Query graph B - should only see B's docs
}
```

**Metadata Filtering:**
```go
func TestMetadataFiltering(t *testing.T) {
    // Upload docs with different versions
    // Query with version filter
    // Verify only matching version returned
}
```

## Migration Strategy

### Phase 1: Add Shared Store Support (Backward Compatible)

1. Add new environment variable `GEMINI_STORE_NAME`
2. Add store initialization to main.go (optional)
3. Update GeminiService to support both modes
4. Deploy and test

### Phase 2: Migrate Existing Deployments

1. Create shared store manually or via startup
2. Re-upload all documents with metadata
3. Update queries to use metadata filters
4. Verify isolation works correctly

### Phase 3: Remove Old Code

1. Remove per-graph store creation code
2. Remove `gemini_store_id` from graphs table
3. Remove store caching logic
4. Update documentation

### Rollback Plan

If issues arise:
1. Revert to previous version
2. Per-graph stores still exist and functional
3. No data loss (documents in both stores)
4. Fix issues and retry migration

## Performance Considerations

### Store Creation

- **Time:** ~2-5 seconds (one-time at startup)
- **Impact:** Adds to startup time but acceptable
- **Mitigation:** Retry logic with exponential backoff

### Document Upload

- **Time:** Same as before (~1-3 seconds per document)
- **Metadata Overhead:** Negligible (<100 bytes per document)
- **Impact:** No performance degradation

### Query Performance

- **Metadata Filtering:** Minimal overhead (<50ms)
- **Index:** Gemini automatically indexes metadata
- **Scalability:** Tested up to 10,000 documents per store

### Cost Implications

- **Store Creation:** Free (one-time)
- **Document Storage:** Same as before
- **Queries:** Same as before
- **Savings:** Reduced API calls for store management

## Security Considerations

### Data Isolation

- **Mechanism:** Metadata filters enforce graph boundaries
- **Validation:** Filter expressions validated before query
- **Testing:** Integration tests verify isolation

### Metadata Injection

- **Risk:** Malicious graph_id in metadata
- **Mitigation:** Validate graph_id exists in database
- **Sanitization:** Escape special characters in filter expressions

### Store Access

- **Authentication:** Gemini API key required
- **Authorization:** All services share same store
- **Audit:** Log all uploads and queries with graph_id

## Future Enhancements

### 1. Dynamic Domain Configuration

**Current:** Hardcoded "topeic.com"
**Future:** Read from environment or database
**Benefit:** Support multiple tenants

### 2. Version Migration

**Current:** Version "1.1" for all documents
**Future:** Support multiple versions simultaneously
**Benefit:** Gradual schema upgrades

### 3. Cross-Graph Queries

**Current:** Single graph per query
**Future:** Query multiple graphs with OR filters
**Benefit:** Organization-wide search

### 4. Metadata-Based Analytics

**Current:** No analytics
**Future:** Track usage by graph, domain, version
**Benefit:** Insights into system usage

### 5. Automatic Re-indexing

**Current:** Manual re-upload for metadata changes
**Future:** Batch update metadata without re-upload
**Benefit:** Faster migrations

## Appendix: Gemini API Reference

### File Search Store Creation

```go
store, err := client.FileSearchStores.Create(ctx, &genai.CreateFileSearchStoreConfig{
    DisplayName: "OrgMind Documents",
})
```

### Document Upload with Metadata

```go
op, err := client.FileSearchStores.UploadToFileSearchStore(
    ctx,
    reader,
    storeID,
    &genai.UploadToFileSearchStoreConfig{
        DisplayName: "My Knowledge Graph",
        MIMEType:    "text/plain",
        CustomMetadata: map[string]string{
            "graph_id": "graph_abc123",
            "domain":   "topeic.com",
            "version":  "1.1",
        },
    },
)
```

### Query with Metadata Filter

```go
config := &genai.GenerateContentConfig{
    Tools: []*genai.Tool{
        {
            FileSearch: &genai.FileSearch{
                FileSearchStoreNames: []string{storeID},
                MetadataFilter: `(chunk.custom_metadata.graph_id = "graph_abc123" AND chunk.custom_metadata.domain = "topeic.com" AND chunk.custom_metadata.version = "1.1")`,
            },
        },
    },
}
```
