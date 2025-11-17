# Requirements Document

## Introduction

OrgMind is an enterprise document processing platform that enables users to upload, edit, and process documents while automatically generating knowledge graphs. The system supports multiple authentication methods, stores documents in cloud storage (AWS S3), processes content through Zep Cloud for knowledge graph generation, and visualizes the resulting graphs using sigma.js.

## Glossary

- **OrgMind_System**: The complete document processing platform including frontend, backend, and integrations
- **Authentication_Service**: The service handling user authentication via email/password, Google OAuth, OpenID Connect (Okta, Office365)
- **Document_Editor**: The Lexical-based rich text editor component in the frontend
- **Storage_Service**: The cloud storage service (AWS S3 implementation) for document persistence
- **Processing_Service**: The backend service that chunks, cleans, and prepares documents for AI processing
- **Zep_Service**: The service integrating with Zep Cloud SDK for knowledge graph creation and retrieval
- **Graph_Visualizer**: The sigma.js-based component for rendering knowledge graphs
- **JWT_Token**: JSON Web Token used for authenticated API requests
- **Document_Chunk**: A segment of document content (max 10,000 characters) prepared for Zep processing
- **Knowledge_Graph**: The graph structure representing relationships between document entities, stored in Zep Grafiti memory

## Requirements

### Requirement 1

**User Story:** As a new user, I want to sign up using multiple authentication methods, so that I can access the platform with my preferred identity provider

#### Acceptance Criteria

1. WHEN a user submits valid email and password credentials, THE Authentication_Service SHALL create a new user account and return a JWT_Token
2. WHEN a user initiates Google OAuth authentication, THE Authentication_Service SHALL redirect to Google's authorization endpoint and process the callback to create or authenticate the user
3. WHEN a user initiates OpenID Connect authentication with Okta or Office365, THE Authentication_Service SHALL complete the OIDC flow and return a JWT_Token upon successful authentication
4. WHERE a user account already exists with the provided credentials, THE Authentication_Service SHALL return an appropriate error message indicating the account exists
5. WHEN authentication succeeds through any method, THE OrgMind_System SHALL store the JWT_Token in the frontend and redirect the user to the home page

### Requirement 2

**User Story:** As an authenticated user, I want to create document content using a rich text editor, so that I can compose documents directly in the platform

#### Acceptance Criteria

1. WHEN an authenticated user navigates to the home page, THE OrgMind_System SHALL display the Document_Editor with full editing capabilities
2. WHILE the user is editing content, THE Document_Editor SHALL provide rich text formatting options including headings, lists, and text styling
3. WHEN the user submits content from the Document_Editor, THE OrgMind_System SHALL send the content to the authenticated backend API with the JWT_Token
4. IF the JWT_Token is invalid or expired, THEN THE OrgMind_System SHALL reject the request and return an authentication error
5. WHEN the backend receives valid editor content, THE Processing_Service SHALL extract the text content for further processing

### Requirement 3

**User Story:** As an authenticated user, I want to upload document files, so that I can process existing documents without manual retyping

#### Acceptance Criteria

1. WHEN an authenticated user selects a file for upload, THE OrgMind_System SHALL validate the file type and size before initiating upload
2. WHEN the user submits a file upload, THE OrgMind_System SHALL send the file to the authenticated backend API with the JWT_Token
3. IF the JWT_Token is invalid or expired, THEN THE OrgMind_System SHALL reject the upload request and return an authentication error
4. WHEN the backend receives a valid file upload, THE Processing_Service SHALL extract the text content from the file
5. WHERE the file format is unsupported, THE OrgMind_System SHALL return an error message indicating supported file types

### Requirement 4

**User Story:** As the system, I want to store documents in cloud storage, so that user content is persisted and retrievable

#### Acceptance Criteria

1. WHEN the Processing_Service receives document content, THE Storage_Service SHALL generate a unique identifier for the document
2. WHEN storing a document, THE Storage_Service SHALL associate the document with the authenticated user's userID extracted from the JWT_Token
3. WHEN the document is ready for storage, THE Storage_Service SHALL upload the document to AWS S3 using credentials from environment variables
4. IF the S3 upload fails, THEN THE Storage_Service SHALL return an error and prevent further processing of that document
5. WHEN the S3 upload succeeds, THE Storage_Service SHALL return the document identifier and storage location

### Requirement 5

**User Story:** As the system, I want to chunk and process documents for AI analysis, so that content can be sent to Zep Cloud within size constraints

#### Acceptance Criteria

1. WHEN the Processing_Service receives document content, THE Processing_Service SHALL split the content into Document_Chunks with a maximum size of 10,000 characters each
2. WHEN creating Document_Chunks, THE Processing_Service SHALL clean and prepare the text by removing unnecessary whitespace and formatting artifacts
3. WHEN Document_Chunks are created, THE Processing_Service SHALL maintain the logical structure and context of the original document
4. WHEN chunking is complete, THE Processing_Service SHALL pass the Document_Chunks to the Zep_Service for knowledge graph creation
5. IF chunking fails due to content issues, THEN THE Processing_Service SHALL log the error and return a failure status

### Requirement 6

**User Story:** As the system, I want to create knowledge graphs in Zep Cloud, so that document relationships and entities are extracted and stored

#### Acceptance Criteria

1. WHEN the Zep_Service receives Document_Chunks, THE Zep_Service SHALL authenticate with Zep Cloud using API keys from environment variables
2. WHEN creating a knowledge graph, THE Zep_Service SHALL use the Zep Go SDK to add memory entries to Zep Grafiti memory
3. WHEN adding memory to Zep Cloud, THE Zep_Service SHALL associate the memory with the user's session or identifier
4. IF the Zep Cloud API returns an error, THEN THE Zep_Service SHALL retry the operation up to three times before failing
5. WHEN the knowledge graph creation succeeds, THE Zep_Service SHALL return a success status to the Processing_Service

### Requirement 7

**User Story:** As an authenticated user, I want to view knowledge graphs generated from my documents, so that I can understand the relationships and entities extracted

#### Acceptance Criteria

1. WHEN an authenticated user requests to view a knowledge graph, THE OrgMind_System SHALL validate the JWT_Token and extract the userID
2. WHEN the backend receives a graph visualization request, THE Zep_Service SHALL query Zep Cloud for the knowledge graph data associated with the user
3. WHEN the Zep_Service retrieves graph data, THE OrgMind_System SHALL transform the data into the format required by sigma.js
4. WHEN the frontend receives graph data, THE Graph_Visualizer SHALL render the knowledge graph with nodes and edges using sigma.js
5. WHERE no knowledge graph exists for the user, THE OrgMind_System SHALL display a message indicating no graphs are available

### Requirement 8

**User Story:** As a returning user, I want to sign in using my existing credentials, so that I can access my documents and knowledge graphs

#### Acceptance Criteria

1. WHEN a user submits valid sign-in credentials, THE Authentication_Service SHALL verify the credentials and return a JWT_Token
2. WHEN sign-in succeeds, THE OrgMind_System SHALL store the JWT_Token and redirect the user to the home page
3. IF the credentials are invalid, THEN THE Authentication_Service SHALL return an error message indicating authentication failure
4. WHEN a user requests password reset, THE Authentication_Service SHALL send a password reset email to the registered email address
5. WHEN the user completes the password reset flow, THE Authentication_Service SHALL update the password and allow sign-in with new credentials

### Requirement 9

**User Story:** As an unauthenticated user, I want to browse the platform landing page, so that I can learn about the platform before registering

#### Acceptance Criteria

1. WHEN an unauthenticated user visits the platform, THE OrgMind_System SHALL display public routes including landing page, sign-up, and sign-in pages
2. WHILE browsing public routes, THE OrgMind_System SHALL not require authentication or JWT_Token validation
3. WHEN an unauthenticated user attempts to access authenticated routes, THE OrgMind_System SHALL redirect to the sign-in page
4. WHERE the user is on a public route, THE OrgMind_System SHALL display navigation options to sign up or sign in
5. WHEN the user navigates between public routes, THE OrgMind_System SHALL maintain a frictionless browsing experience

### Requirement 10

**User Story:** As a system administrator, I want all sensitive credentials stored in environment variables, so that security best practices are maintained

#### Acceptance Criteria

1. WHEN the backend initializes, THE OrgMind_System SHALL load all API keys and credentials from environment variables
2. THE OrgMind_System SHALL require environment variables for DATABASE_URL, AWS S3 credentials, Zep API keys, and OAuth client secrets
3. IF required environment variables are missing, THEN THE OrgMind_System SHALL fail to start and log descriptive error messages
4. THE OrgMind_System SHALL never hardcode sensitive credentials in source code or configuration files
5. WHEN deploying to different environments, THE OrgMind_System SHALL use environment-specific variable values without code changes
