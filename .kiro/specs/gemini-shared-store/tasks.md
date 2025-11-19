# Implementation Plan: Gemini Shared File Search Store

## Overview

This implementation plan refactors the Gemini File Search integration to use a single shared store with metadata-based filtering. Tasks are ordered to maintain backward compatibility during the transition.

## Tasks

- [x] 1. Update configuration for shared store
- [x] 1.1 Add GEMINI_STORE_NAME to environment configuration
  - Add to `backend/.env.example` with default value "OrgMind Documents"
  - Add to `backend/.env` for local development
  - Update `backend/internal/config/config.go` to read GEMINI_STORE_NAME
  - Add GeminiStoreID field to Config struct (runtime value)
  - _Requirements: 2.1, 2.2, 2.3_

- [x] 1.2 Update environment documentation
  - Update `backend/GEMINI_SETUP.md` with shared store architecture
  - Update `ENVIRONMENT_SETUP.md` with GEMINI_STORE_NAME variable
  - Document metadata filtering approach
  - _Requirements: 2.1, 7.1, 7.2_

- [x] 2. Refactor GeminiService interface
- [x] 2.1 Update service interface definition
  - Add `InitializeStore(ctx, storeName) (storeID, error)` method
  - Update `UploadDocument` signature to include graphID, graphName
  - Update `GenerateStreamingResponse` signature to include graphID, domain, version
  - Remove `CreateFileSearchStore`, `GetFileSearchStore`, `DeleteFileSearchStore` methods
  - Update `backend/internal/service/interfaces.go`
  - _Requirements: 1.1, 1.2, 3.1, 4.1, 6.2_

- [x] 2.2 Update GeminiService struct
  - Add `storeID` field (shared store ID)
  - Add `storeName` field (store display name)
  - Remove `storeCache sync.Map` field
  - Update constructor to accept storeID and storeName
  - Update `backend/internal/service/gemini_service.go`
  - _Requirements: 1.1, 6.4_

- [x] 3. Implement store initialization
- [x] 3.1 Implement InitializeStore method
  - Create File Search store with provided name
  - Implement retry logic (3 attempts with exponential backoff)
  - Return store ID on success
  - Log detailed errors on failure
  - Update `backend/internal/service/gemini_service.go`
  - _Requirements: 1.1, 5.1, 5.2, 7.1, 7.2_

- [x] 3.2 Add store initialization to main.go
  - Call `InitializeStore` after GeminiService creation
  - Store returned store ID in config
  - Exit with error if initialization fails (when Gemini is configured)
  - Skip initialization if GEMINI_API_KEY not set
  - Update `backend/cmd/server/main.go`
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

- [x] 4. Implement metadata-based document upload
- [x] 4.1 Update UploadDocument method signature
  - Add graphID parameter
  - Add graphName parameter
  - Keep documentID, content, mimeType parameters
  - Update `backend/internal/service/gemini_service.go`
  - _Requirements: 3.1, 3.4_

- [x] 4.2 Implement metadata attachment in UploadDocument
  - Set display_name to graphName
  - Add custom_metadata with graph_id = graphID
  - Add custom_metadata with domain = "topeic.com" (hardcoded)
  - Add custom_metadata with version = "1.1"
  - Use shared storeID from service struct
  - Update `backend/internal/service/gemini_service.go`
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 4.3 Update document service to pass graph information
  - Get graph name from graphRepo in uploadToFileSearch
  - Pass graphID and graphName to UploadDocument
  - Remove store lookup logic
  - Update `backend/internal/service/document_service.go`
  - _Requirements: 3.1, 3.4, 6.3_

- [x] 5. Implement metadata-based query filtering
- [x] 5.1 Update GenerateStreamingResponse signature
  - Add graphID parameter
  - Add domain parameter
  - Add version parameter
  - Keep query and responseChan parameters
  - Update `backend/internal/service/gemini_service.go`
  - _Requirements: 4.1, 4.2, 4.3_

- [x] 5.2 Implement metadata filter construction
  - Build filter expression: `(chunk.custom_metadata.graph_id = "{graphID}" AND chunk.custom_metadata.domain = "{domain}" AND chunk.custom_metadata.version = "{version}")`
  - Escape special characters in values
  - Validate filter syntax
  - Log filter expression
  - Update `backend/internal/service/gemini_service.go`
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 7.4_

- [x] 5.3 Apply metadata filter in File Search tool
  - Add MetadataFilter field to FileSearch config
  - Use shared storeID from service struct
  - Update `backend/internal/service/gemini_service.go`
  - _Requirements: 4.1, 4.2, 4.3, 4.5_

- [x] 5.4 Update chat service to pass filtering parameters
  - Get graph from graphRepo in GenerateResponseForMessage
  - Pass graphID, "topeic.com", "1.1" to GenerateStreamingResponse
  - Update `backend/internal/service/chat_service.go`
  - _Requirements: 4.1, 4.2, 4.3, 9.1_

- [x] 6. Remove deprecated code
- [x] 6.1 Remove per-graph store methods
  - Remove CreateFileSearchStore method (keep only InitializeStore)
  - Remove GetFileSearchStore method
  - Remove DeleteFileSearchStore method
  - Remove storeCache field and related code
  - Update `backend/internal/service/gemini_service.go`
  - _Requirements: 6.2, 6.3, 6.4_

- [x] 6.2 Update graph repository interface
  - Mark UpdateGeminiStoreID as deprecated (keep for backward compatibility)
  - Document that gemini_store_id column is no longer used
  - Update `backend/internal/repository/interfaces.go`
  - _Requirements: 6.1_

- [x] 6.3 Remove store creation from document upload flow
  - Remove GetFileSearchStore calls
  - Remove CreateFileSearchStore calls
  - Remove store ID caching logic
  - Update `backend/internal/service/document_service.go`
  - _Requirements: 6.3_

- [x] 7. Add configuration validation
- [x] 7.1 Implement startup validation
  - Validate GEMINI_API_KEY is set (if Gemini features enabled)
  - Validate GEMINI_PROJECT_ID is set
  - Validate GEMINI_LOCATION is set
  - Validate GEMINI_STORE_NAME format (alphanumeric, spaces, hyphens)
  - Exit with clear error if validation fails
  - Update `backend/cmd/server/main.go`
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

- [x] 8. Update logging
- [x] 8.1 Add store initialization logging
  - Log store creation attempt with name
  - Log successful creation with store ID
  - Log failures with retry information
  - Update `backend/internal/service/gemini_service.go`
  - _Requirements: 1.5, 7.1, 7.2_

- [x] 8.2 Add document upload logging
  - Log upload with graph_id, domain, version metadata
  - Log successful upload with file ID
  - Log failures with detailed error
  - Update `backend/internal/service/gemini_service.go`
  - _Requirements: 7.3, 10.5_

- [x] 8.3 Add query filtering logging
  - Log metadata filter expression used
  - Log query execution with graph_id
  - Log streaming completion
  - Update `backend/internal/service/gemini_service.go`
  - _Requirements: 7.4_

- [ ] 9. Update documentation
- [ ] 9.1 Update GEMINI_SETUP.md
  - Document shared store architecture
  - Document metadata-based filtering
  - Update setup instructions for GEMINI_STORE_NAME
  - Document domain and version metadata
  - Add troubleshooting for store initialization
  - Update `backend/GEMINI_SETUP.md`
  - _Requirements: 1.5, 3.1, 3.2, 3.3, 4.1, 9.2_

- [ ] 9.2 Update CHAT_FLOW.md
  - Document metadata filtering in query flow
  - Update sequence diagram with shared store
  - Document graph isolation mechanism
  - Update `backend/CHAT_FLOW.md`
  - _Requirements: 4.1, 4.2, 4.3_

- [ ] 9.3 Create migration guide
  - Document migration from per-graph to shared store
  - Provide rollback instructions
  - Document data re-upload process (if needed)
  - Create `backend/GEMINI_MIGRATION.md`
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [ ] 10. Testing
- [ ] 10.1 Test store initialization
  - Test successful store creation at startup
  - Test retry logic on transient failures
  - Test error handling on permanent failures
  - Test skip when GEMINI_API_KEY not set
  - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [ ] 10.2 Test document upload with metadata
  - Test metadata attachment (graph_id, domain, version)
  - Test display_name set to graph name
  - Test upload to shared store
  - Test error handling
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [ ] 10.3 Test metadata filtering in queries
  - Test graph isolation (query graph A, only see A's docs)
  - Test domain filtering
  - Test version filtering
  - Test combined AND filters
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [ ] 10.4 Test multi-graph isolation
  - Upload docs to multiple graphs
  - Query each graph independently
  - Verify no cross-graph data leakage
  - _Requirements: 4.1, 4.2_

- [ ]* 10.5 Test configuration validation
  - Test missing GEMINI_API_KEY
  - Test missing GEMINI_PROJECT_ID
  - Test invalid GEMINI_STORE_NAME format
  - Test error messages are clear
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

## Notes

- Tasks marked with `*` are optional (testing tasks)
- Store initialization happens once at startup, not per-request
- Metadata filtering ensures graph isolation without separate stores
- Domain is hardcoded to "topeic.com" for now, will be configurable later
- Version "1.1" allows for future schema migrations
- Backward compatibility maintained during transition
