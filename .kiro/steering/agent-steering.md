---
inclusion: always
---

# OrgMind Project Guidelines

## Project Overview

OrgMind is an enterprise document processing platform that:
- Accepts documents from multiple sources (file uploads, text editor, email servers, API clients)
- Parses and chunks documents (max 10,000 characters per chunk due to Zep limitations)
- Creates knowledge graphs using Zep SDK in Zep Cloud
- Visualizes knowledge graphs using sigma.js

Use the Zep documentation MCP server (configured in `.kiro/settings/mcp.json`) for Zep-related implementation details.

## Tech Stack

### Frontend
- Next.js with React, TypeScript, Tailwind CSS
- Source: `src/` directory
- Routes: `src/app/` directory

### Backend
- Go with gin-gonic/gin framework
- PostgreSQL with sqlx and Masterminds/squirrel
- Database connection via `DATABASE_URL` environment variable
- NEVER use GORM - use sqlx only

### File Storage
- Interface-based design with multiple implementations:
  - AWS S3
  - Google Cloud Storage
  - MinIO

### Document Processing
- Use efficient Go libraries for document chunking
- Integrate Zep SDK for knowledge graph creation
- Consider langchain-go for document processing pipelines

## Architecture Patterns

### Frontend Routing (Next.js App Router)
- `(folderName)` = route groups (not in URL path)
- `folderName` = route segments (in URL path)
- `(public)` = unauthenticated routes
- `(auth)` = authenticated routes

### Backend Layered Architecture
Follow this strict layering:
1. Router - defines API endpoints
2. Handler - handles HTTP requests/responses
3. Service - business logic (use interfaces for multiple implementations)
4. Repository - data access layer

API routes should mirror frontend structure with public/auth separation.

### Authentication
Support multiple authentication methods:
- Email/password signup
- Google OAuth
- OpenID Connect (Okta and other enterprise SSO providers)

### Design Principles
- Use interfaces everywhere for swappable implementations (storage, services, message queues)
- Modular, expandable codebase for maintainability
- Frictionless UX: allow browsing without registration, require auth for actions

## Code Style Guidelines

### Go Backend
- Always use interfaces for dependencies (enables testing and swapping implementations)
- JSON struct tags: always use camelCase to match frontend expectations
- SQL queries: explicitly list all columns, NEVER use `SELECT *`
- For `pq: bind message` errors: use `sqlx.QueryxContext` with manual `rows.Scan()` instead of `SelectContext`/`GetContext`
- Use popular, well-maintained libraries from GitHub

### TypeScript Frontend
- JSON/API types: use camelCase to match backend
- Always check for `undefined`/`null` before accessing properties (especially `.length`)
- Use optional chaining (`?.`) and nullish coalescing (`??`) operators
- Client components: use `useParams()` hook for route parameters, NOT `params` prop
- Knowledge graph visualization: use sigma.js library
- Use popular, well-maintained libraries from npm

## Defensive Programming & Error Handling

### API Response Handling (Critical)
**Problem**: Backend may return inconsistent response formats (wrapped objects vs direct arrays), null values, or unexpected structures. This causes runtime crashes like `"graphs.map is not a function"`.

**Solution - Defense in Depth**:

1. **Frontend API Layer** (`frontend/lib/api/*.ts`):
   - ALWAYS handle multiple response formats defensively
   - Check if response is an array before returning
   - Check for wrapped responses (e.g., `{graphs: [...]}` vs `[...]`)
   - Return empty arrays as fallback for list operations
   - Example pattern:
   ```typescript
   export async function listItems(): Promise<Item[]> {
     const response = await apiCall<{ items: Item[] } | Item[]>('/api/items');
     
     // Handle direct array response
     if (Array.isArray(response)) {
       return response;
     }
     
     // Handle wrapped response
     if (response && typeof response === 'object' && 'items' in response) {
       return Array.isArray(response.items) ? response.items : [];
     }
     
     // Fallback to empty array
     return [];
   }
   ```

2. **React Components**:
   - ALWAYS check `Array.isArray()` before calling `.map()`, `.filter()`, etc.
   - Add defensive checks in both empty state conditions AND render logic
   - Example:
   ```typescript
   if (!items || !Array.isArray(items) || items.length === 0) {
     return <EmptyState />;
   }
   
   return (
     <div>
       {Array.isArray(items) && items.map(item => <Item key={item.id} {...item} />)}
     </div>
   );
   ```

3. **Backend Handlers** (Go):
   - Return consistent response structures
   - For list operations, ALWAYS return an array (even if empty)
   - Prefer direct array responses over wrapped objects for simplicity
   - If wrapping is needed, be consistent across all endpoints
   - Example:
   ```go
   // Good: Direct array (preferred)
   c.JSON(http.StatusOK, response)
   
   // Also acceptable: Wrapped (but be consistent)
   c.JSON(http.StatusOK, gin.H{"items": response})
   ```

**Apply this pattern to ALL list operations**:
- Graph lists
- Document lists
- Member lists
- Any other collection endpoints

**Why Defense in Depth?**
- Backend may change response format during refactoring
- Network issues may cause partial responses
- Third-party APIs may return unexpected formats
- Better to handle gracefully than crash the application

## Critical External Dependencies (Zep Integration)

### Graph Creation Order (Critical)
When creating resources that depend on external services (like Zep Cloud), follow this strict order:

**1. Create in External Service FIRST** (Zep is the source of truth for knowledge graphs)
```go
// Step 1: Create in Zep Cloud (CRITICAL - fail fast if this fails)
zepGraphID, err := s.zepSvc.CreateGraph(ctx, graphID, name, description)
if err != nil {
    return nil, fmt.Errorf("%w: %v", ErrZepGraphCreation, err)
}

// Step 2: Create in database (with rollback on failure)
if err := s.graphRepo.Create(ctx, graph); err != nil {
    _ = s.zepSvc.DeleteGraph(ctx, zepGraphID) // Rollback Zep
    return nil, fmt.Errorf("failed to create graph in database: %w", err)
}

// Step 3: Create related records (with full rollback on failure)
if err := s.graphRepo.CreateMembership(ctx, membership); err != nil {
    _ = s.graphRepo.Delete(ctx, graphID)      // Rollback database
    _ = s.zepSvc.DeleteGraph(ctx, zepGraphID) // Rollback Zep
    return nil, fmt.Errorf("failed to create membership: %w", err)
}
```

**Why This Order?**
- Zep is the critical dependency - without it, the graph is useless
- Fail fast if Zep is unavailable (don't create orphaned database records)
- Database records are easier to clean up than Zep graphs
- Prevents inconsistent state between systems

**Deletion Order:**
```go
// Step 1: Delete from Zep first (log errors but continue)
if err := s.zepSvc.DeleteGraph(ctx, zepGraphID); err != nil {
    // Log but don't fail - Zep graph might already be deleted
    fmt.Printf("Warning: failed to delete from Zep: %v\n", err)
}

// Step 2: Delete from database (cascade deletes related records)
if err := s.graphRepo.Delete(ctx, graphID); err != nil {
    return fmt.Errorf("failed to delete from database: %w", err)
}
```

**Apply this pattern to:**
- Graph creation/deletion
- Document processing (Zep memory operations)
- Any operation involving external critical services

## Neon Database Concurrency Issues (Known Limitation)

### Problem: Prepared Statement Conflicts
Neon's serverless PostgreSQL has issues with concurrent queries using prepared statements (sqlx). When multiple requests hit the backend simultaneously, you may see errors like:

```
"pq: bind message supplies 2 parameters, but prepared statement \"\" requires 1"
```

This happens inconsistently - sometimes one request fails, sometimes another, making it hard to debug.

### Root Cause
- Neon uses connection pooling with statement caching
- Concurrent requests can interfere with each other's prepared statements
- The `sqlx` library in Go uses prepared statements by default
- Statement cache conflicts occur when multiple queries execute simultaneously

### Workaround (Frontend)
**DO NOT use `Promise.all()` for concurrent API calls to the same backend**

```typescript
// ❌ BAD: Causes Neon prepared statement conflicts
const [graphDetails, graphDocs] = await Promise.all([
  getGraph(graphId),
  listGraphDocuments(graphId),
]);

// ✅ GOOD: Sequential requests avoid conflicts
const graphDetails = await getGraph(graphId);
const graphDocs = await listGraphDocuments(graphId);
```

**Always add a comment explaining why:**
```typescript
// WORKAROUND: Neon DB has issues with concurrent prepared statements
// causing "bind message supplies X parameters, but prepared statement requires Y" errors
// Using sequential requests instead of Promise.all until backend connection pooling is fixed
const graphDetails = await getGraph(graphId);
const graphDocs = await listGraphDocuments(graphId);
```

**Keep the original `Promise.all` commented out** for future reference when the issue is resolved.

### Potential Backend Solutions (Not Yet Implemented)
1. **Disable Prepared Statements in sqlx:**
   ```go
   // Use Unsafe() to bypass prepared statements
   db.Unsafe().QueryxContext(ctx, query, args...)
   ```

2. **Use Connection Pooling with Statement Cache Disabled:**
   ```go
   db.SetMaxOpenConns(1) // Force single connection (not recommended for production)
   ```

3. **Switch to `database/sql` with manual scanning** instead of sqlx

4. **Configure Neon Connection String:**
   ```
   ?statement_cache_mode=describe
   ```

### When to Apply This Workaround
- Any page that fetches multiple resources on load
- Any component that makes concurrent API calls
- Especially when queries involve JOINs or complex WHERE clauses
- More likely to occur with graph/document operations (they use membership checks)

### Future Resolution
Once Neon fixes their prepared statement caching or we implement a backend solution:
1. Uncomment the `Promise.all` code
2. Remove the sequential workaround
3. Update this documentation
4. Test thoroughly with concurrent requests

## Google Gemini File Search Integration

### Overview
OrgMind integrates Google Gemini's File Search capability for AI-powered document Q&A. Documents are uploaded to File Search stores (one per graph) and used to provide context-aware responses.

### Official Documentation & Resources

**Primary Documentation:**
- **File Search Overview**: https://ai.google.dev/gemini-api/docs/file-search
- **File Search Stores Guide**: https://ai.google.dev/gemini-api/docs/file-search#file-search-stores
- **REST API Reference**: https://ai.google.dev/gemini-api/docs/file-search#rest

**Go SDK Resources:**
- **GitHub Repository**: https://github.com/googleapis/go-genai
- **Example Implementation**: https://github.com/googleapis/go-genai/blob/main/examples/filesearchstores/create_upload_and_call_file_search.go
- **SDK Implementation**: https://github.com/googleapis/go-genai/blob/main/filesearchstores.go

### Key Implementation Notes

1. **File Search Store Management**:
   - Create one File Search store per graph (use graph_id as identifier)
   - Store the `gemini_store_id` in the graphs table
   - Cache store IDs to avoid repeated lookups

2. **Document Upload**:
   - Upload documents asynchronously in background goroutines
   - Store `gemini_file_id` in documents table for tracking
   - Handle upload failures gracefully (log but don't fail main flow)
   - Continue with Zep processing even if Gemini upload fails

3. **Query and Response**:
   - Query File Search store for relevant document content
   - Use retrieved context with Gemini API to generate responses
   - Stream responses using Server-Sent Events (SSE)
   - Save both user and AI messages to database

4. **Error Handling**:
   - Implement retry logic with exponential backoff (3 attempts)
   - Log all Gemini API interactions for debugging
   - Provide user-friendly error messages
   - Fall back gracefully if File Search is unavailable

5. **SDK Usage Pattern**:
   ```go
   // Example pattern from go-genai SDK
   import "github.com/googleapis/go-genai"
   
   // Create client
   client, err := genai.NewClient(ctx, &genai.ClientConfig{
       APIKey: os.Getenv("GEMINI_API_KEY"),
       Backend: genai.BackendGoogleAI,
   })
   
   // Create File Search store
   store, err := client.CreateFileSearchStore(ctx, &genai.CreateFileSearchStoreRequest{
       DisplayName: graphID,
   })
   
   // Upload document
   file, err := client.UploadFile(ctx, &genai.UploadFileRequest{
       File: bytes.NewReader(content),
       MIMEType: mimeType,
   })
   
   // Query with File Search
   response := client.GenerateContent(ctx, &genai.GenerateContentRequest{
       Model: "gemini-1.5-flash",
       Contents: []*genai.Content{{
           Parts: []*genai.Part{{Text: query}},
       }},
       Tools: []*genai.Tool{{
           FileSearchTool: &genai.FileSearchTool{
               FileSearchStoreIDs: []string{storeID},
           },
       }},
   })
   ```

6. **Environment Variables**:
   ```bash
   GEMINI_API_KEY=your-api-key
   GEMINI_PROJECT_ID=your-project-id
   GEMINI_LOCATION=us-central1
   ```

### Integration with Existing Services

Follow the same pattern as Zep integration:
- Create external resource FIRST (File Search store)
- Store reference ID in database
- Implement rollback on failure
- Process asynchronously to avoid blocking

### Cost Considerations
- Monitor API usage and costs
- Implement rate limiting
- Consider caching responses for common queries
- Set appropriate quotas and alerts

## Required Libraries
- Backend: Zep SDK, Google Gemini SDK (go-genai), langchain-go, sqlx, Masterminds/squirrel, gin-gonic/gin
- Frontend: sigma.js for knowledge graph visualization, Zustand for chat state management
