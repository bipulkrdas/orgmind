# Requirements Document: Gemini Shared File Search Store

## Introduction

This document outlines the requirements for refactoring the Gemini File Search integration to use a single shared store with metadata-based filtering instead of creating separate stores per graph. This approach improves scalability, reduces API calls, and simplifies store management.

## Glossary

- **File Search Store**: A Gemini API resource that holds uploaded documents for semantic search
- **Custom Metadata**: Key-value pairs attached to documents for filtering and organization
- **Metadata Filter**: Query-time filter expression to retrieve specific documents based on metadata
- **Graph ID**: Unique identifier for a knowledge graph in OrgMind
- **Domain**: Tenant or organization identifier for multi-tenant isolation
- **Store Name**: Human-readable identifier for the File Search store

## Requirements

### Requirement 1: Single Shared Store Architecture

**User Story:** As a system administrator, I want a single File Search store for all documents, so that I can reduce API overhead and simplify management.

#### Acceptance Criteria

1. THE System SHALL create exactly one File Search store during application startup
2. THE System SHALL use the same store for all graph documents across all tenants
3. THE System SHALL NOT create additional stores after initial startup
4. THE System SHALL fail startup if store creation fails and no existing store is configured
5. THE System SHALL log store creation success with store ID and name

### Requirement 2: Store Configuration

**User Story:** As a system administrator, I want to configure the File Search store name via environment variables, so that I can identify stores across environments.

#### Acceptance Criteria

1. THE System SHALL read store name from environment variable `GEMINI_STORE_NAME`
2. THE System SHALL use default name "OrgMind Documents" when `GEMINI_STORE_NAME` is not set
3. THE System SHALL validate store name is not empty before creation
4. THE System SHALL store the created store ID in application configuration
5. THE System SHALL make store ID available to all services that need it

### Requirement 3: Document Metadata Tagging

**User Story:** As a developer, I want documents tagged with graph_id, domain, and version metadata, so that I can filter documents at query time.

#### Acceptance Criteria

1. WHEN uploading a document, THE System SHALL attach custom metadata with key "graph_id" and value equal to the graph's ID
2. WHEN uploading a document, THE System SHALL attach custom metadata with key "domain" and value "topeic.com"
3. WHEN uploading a document, THE System SHALL attach custom metadata with key "version" and value "1.1"
4. WHEN uploading a document, THE System SHALL set display_name to the graph's name
5. THE System SHALL ensure all metadata keys and values are non-empty strings

### Requirement 4: Metadata-Based Query Filtering

**User Story:** As a user, I want chat responses based only on my graph's documents, so that I don't see data from other graphs.

#### Acceptance Criteria

1. WHEN generating AI responses, THE System SHALL apply metadata filter for graph_id matching the current graph
2. WHEN generating AI responses, THE System SHALL apply metadata filter for domain matching "topeic.com"
3. WHEN generating AI responses, THE System SHALL apply metadata filter for version matching "1.1"
4. THE System SHALL combine filters using AND logic
5. THE System SHALL use chunk-level metadata filter syntax: `chunk.custom_metadata.{key}`

### Requirement 5: Store Initialization at Startup

**User Story:** As a system administrator, I want the store created during app startup, so that document uploads don't fail due to missing store.

#### Acceptance Criteria

1. THE System SHALL attempt to create File Search store during main() initialization
2. IF store creation fails, THE System SHALL log detailed error and exit with non-zero status
3. IF Gemini API key is not configured, THE System SHALL skip store creation and log warning
4. THE System SHALL pass store ID to GeminiService constructor
5. THE System SHALL initialize GeminiService after successful store creation

### Requirement 6: Backward Compatibility

**User Story:** As a developer, I want existing graph-specific store references removed, so that the codebase is consistent.

#### Acceptance Criteria

1. THE System SHALL remove `gemini_store_id` column usage from graphs table
2. THE System SHALL remove `GetFileSearchStore()` method from GeminiService
3. THE System SHALL remove `CreateFileSearchStore()` calls from document upload flow
4. THE System SHALL remove store ID caching logic from GeminiService
5. THE System SHALL update all service methods to use shared store ID from config

### Requirement 7: Error Handling and Logging

**User Story:** As a system administrator, I want detailed logging of store operations, so that I can troubleshoot issues.

#### Acceptance Criteria

1. THE System SHALL log store creation attempt with store name
2. THE System SHALL log successful store creation with store ID
3. THE System SHALL log document upload with graph_id, domain, and version metadata
4. THE System SHALL log metadata filter expression used in queries
5. THE System SHALL log detailed errors for store creation failures with retry information

### Requirement 8: Configuration Validation

**User Story:** As a system administrator, I want configuration validated at startup, so that I catch errors early.

#### Acceptance Criteria

1. THE System SHALL validate GEMINI_API_KEY is set before creating store
2. THE System SHALL validate GEMINI_PROJECT_ID is set before creating store
3. THE System SHALL validate GEMINI_LOCATION is set before creating store
4. THE System SHALL validate GEMINI_STORE_NAME format (alphanumeric, spaces, hyphens only)
5. THE System SHALL exit with clear error message if validation fails

### Requirement 9: Domain Configuration (Future)

**User Story:** As a system administrator, I want to configure domain dynamically, so that I can support multiple tenants.

#### Acceptance Criteria

1. THE System SHALL accept domain value as hardcoded "topeic.com" in version 1.1
2. THE System SHALL document that domain will be configurable in future versions
3. THE System SHALL structure code to easily add domain configuration later
4. THE System SHALL use consistent domain value across all documents
5. THE System SHALL include domain in all metadata filters

### Requirement 10: Version Management

**User Story:** As a developer, I want version metadata on documents, so that I can handle schema changes in the future.

#### Acceptance Criteria

1. THE System SHALL use version "1.1" for all documents in this release
2. THE System SHALL document version format and meaning
3. THE System SHALL filter queries by version to ensure compatibility
4. THE System SHALL allow for version upgrades in future releases
5. THE System SHALL log version number with each document upload
