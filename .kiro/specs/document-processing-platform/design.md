# Design Document

## Overview

OrgMind is a full-stack enterprise document processing platform built with Next.js (frontend) and Go (backend). The system enables users to authenticate via multiple methods, create or upload documents, process them into knowledge graphs using Zep Cloud, and visualize the results. The architecture follows a layered backend design with interface-based services for flexibility and testability.

### Key Technologies
- **Frontend**: Next.js 14+ (App Router), React, TypeScript, Tailwind CSS, Lexical Editor, sigma.js
- **Backend**: Go 1.21+, gin-gonic/gin, sqlx, Masterminds/squirrel, Zep Go SDK
- **Storage**: AWS S3 (interface-based for future providers)
- **Database**: PostgreSQL
- **Authentication**: JWT with multiple OAuth providers

## Architecture

### High-Level Architecture

```mermaid
graph TB
    subgraph Frontend["Frontend (Next.js)"]
        Public[Public Routes<br/>Landing, Sign-up, Sign-in]
        Auth[Authenticated Routes<br/>Home, Editor, Graph Viewer]
        Editor[Lexical Editor]
        Upload[File Upload]
        GraphViz[Sigma.js Graph Visualizer]
    end
    
    subgraph Backend["Backend (Go + Gin)"]
        PublicAPI[Public API Routes<br/>Auth endpoints]
        AuthAPI[Authenticated API Routes<br/>JWT validation]
        AuthSvc[Authentication Service]
        DocSvc[Document Service]
        ProcSvc[Processing Service]
        ZepSvc[Zep Service]
        StorageSvc[Storage Service]
    end
    
    subgraph External["External Services"]
        S3[AWS S3]
        Zep[Zep Cloud]
        OAuth[OAuth Providers<br/>Google, Okta, Office365]
        DB[(PostgreSQL)]
    end
    
    Public --> PublicAPI
    Auth --> AuthAPI
    Editor --> AuthAPI
    Upload --> AuthAPI
    GraphViz --> AuthAPI
    
    PublicAPI --> AuthSvc
    AuthAPI --> DocSvc
    DocSvc --> ProcSvc
    DocSvc --> StorageSvc
    ProcSvc --> ZepSvc
    
    AuthSvc --> OAuth
    AuthSvc --> DB
    DocSvc --> DB
    StorageSvc --> S3
    ZepSvc --> Zep
```

### Frontend Architecture (Next.js App Router)

```
src/
├── app/
│   ├── (public)/              # Route group - unauthenticated
│   │   ├── page.tsx           # Landing page
│   │   ├── signup/
│   │   │   └── page.tsx       # Sign-up page
│   │   ├── signin/
│   │   │   └── page.tsx       # Sign-in page
│   │   ├── reset-password/
│   │   │   └── page.tsx       # Password reset
│   │   └── auth/
│   │       └── callback/
│   │           └── page.tsx   # OAuth callback handler
│   ├── (auth)/                # Route group - authenticated
│   │   ├── layout.tsx         # Auth layout with JWT check
│   │   ├── home/
│   │   │   └── page.tsx       # Home with editor & upload
│   │   └── graphs/
│   │       └── [graphId]/
│   │           └── page.tsx   # Graph visualization
│   └── layout.tsx             # Root layout
├── components/
│   ├── editor/
│   │   └── LexicalEditor.tsx  # Lexical rich text editor
│   ├── upload/
│   │   └── FileUpload.tsx     # File upload component
│   ├── graph/
│   │   └── GraphVisualizer.tsx # Sigma.js graph renderer
│   └── auth/
│       ├── SignUpForm.tsx
│       ├── SignInForm.tsx
│       └── OAuthButtons.tsx
├── lib/
│   ├── api/
│   │   ├── auth.ts            # Auth API calls
│   │   ├── documents.ts       # Document API calls
│   │   └── graphs.ts          # Graph API calls
│   ├── auth/
│   │   └── jwt.ts             # JWT token management
│   └── types/
│       └── index.ts           # TypeScript types
└── middleware.ts              # Route protection middleware
```

### Backend Architecture (Go Layered)

```
backend/
├── cmd/
│   └── server/
│       └── main.go            # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go          # Environment config loader
│   ├── router/
│   │   ├── router.go          # Main router setup
│   │   ├── public.go          # Public routes
│   │   └── auth.go            # Authenticated routes
│   ├── handler/
│   │   ├── auth_handler.go    # Auth HTTP handlers
│   │   ├── document_handler.go
│   │   └── graph_handler.go
│   ├── service/
│   │   ├── auth_service.go    # Auth business logic
│   │   ├── document_service.go
│   │   ├── processing_service.go
│   │   ├── zep_service.go
│   │   └── interfaces.go      # Service interfaces
│   ├── repository/
│   │   ├── user_repository.go
│   │   ├── document_repository.go
│   │   └── interfaces.go      # Repository interfaces
│   ├── storage/
│   │   ├── storage.go         # Storage interface
│   │   └── s3_storage.go      # S3 implementation
│   ├── middleware/
│   │   └── auth_middleware.go # JWT validation
│   └── models/
│       ├── user.go
│       ├── document.go
│       └── graph.go
├── pkg/
│   └── utils/
│       ├── jwt.go             # JWT utilities
│       └── chunker.go         # Document chunking
├── migrations/
│   └── *.sql                  # Database migrations
├── .env.example
└── go.mod
```

## Components and Interfaces

### Frontend Components

#### 1. LexicalEditor Component
```typescript
interface LexicalEditorProps {
  onSubmit: (content: string) => Promise<void>;
  initialContent?: string;
}

// Features:
// - Rich text editing (headings, lists, bold, italic)
// - Auto-save to local storage
// - Submit button to send content to backend
// - Loading state during submission
```

#### 2. FileUpload Component
```typescript
interface FileUploadProps {
  onUpload: (file: File) => Promise<void>;
  acceptedTypes: string[];
  maxSizeBytes: number;
}

// Features:
// - Drag-and-drop support
// - File type validation
// - Size validation
// - Progress indicator
// - Error handling
```

#### 3. GraphVisualizer Component
```typescript
interface GraphVisualizerProps {
  graphData: GraphData;
  width: number;
  height: number;
}

interface GraphData {
  nodes: Array<{
    id: string;
    label: string;
    size: number;
    color: string;
  }>;
  edges: Array<{
    id: string;
    source: string;
    target: string;
    label?: string;
  }>;
}

// Features:
// - Interactive node/edge visualization
// - Zoom and pan controls
// - Node selection and highlighting
// - Layout algorithms (force-directed)
```

#### 4. Authentication Components
```typescript
// SignUpForm, SignInForm
interface AuthFormProps {
  onSubmit: (credentials: Credentials) => Promise<void>;
  onOAuthClick: (provider: OAuthProvider) => void;
}

// OAuthButtons
interface OAuthButtonsProps {
  providers: OAuthProvider[];
  onProviderClick: (provider: OAuthProvider) => void;
}

type OAuthProvider = 'google' | 'okta' | 'office365';
```

### Backend Interfaces

#### 1. Storage Interface
```go
type StorageService interface {
    Upload(ctx context.Context, userID string, filename string, content io.Reader) (string, error)
    Download(ctx context.Context, documentID string) (io.ReadCloser, error)
    Delete(ctx context.Context, documentID string) error
    GetURL(ctx context.Context, documentID string) (string, error)
}

// S3StorageService implements StorageService
type S3StorageService struct {
    client    *s3.Client
    bucket    string
    region    string
}
```

#### 2. Authentication Service Interface
```go
type AuthService interface {
    SignUp(ctx context.Context, email, password string) (*User, string, error)
    SignIn(ctx context.Context, email, password string) (string, error)
    ValidateToken(ctx context.Context, token string) (*Claims, error)
    InitiateOAuth(ctx context.Context, provider string) (string, error)
    HandleOAuthCallback(ctx context.Context, provider, code string) (string, error)
    ResetPassword(ctx context.Context, email string) error
    UpdatePassword(ctx context.Context, token, newPassword string) error
}
```

#### 3. Document Service Interface
```go
type DocumentService interface {
    CreateFromEditor(ctx context.Context, userID, content string) (*Document, error)
    CreateFromFile(ctx context.Context, userID string, file multipart.File, filename string) (*Document, error)
    GetDocument(ctx context.Context, documentID string) (*Document, error)
    ListUserDocuments(ctx context.Context, userID string) ([]*Document, error)
}
```

#### 4. Processing Service Interface
```go
type ProcessingService interface {
    ChunkDocument(ctx context.Context, content string) ([]string, error)
    CleanText(ctx context.Context, text string) (string, error)
    ProcessDocument(ctx context.Context, userID, documentID, content string) error
}

// Chunking strategy:
// - Max 10,000 characters per chunk
// - Preserve sentence boundaries
// - Maintain context overlap (200 chars)
```

#### 5. Zep Service Interface
```go
type ZepService interface {
    AddMemory(ctx context.Context, userID string, chunks []string, metadata map[string]interface{}) error
    GetGraph(ctx context.Context, userID string) (*GraphData, error)
    SearchMemory(ctx context.Context, userID, query string) ([]MemoryResult, error)
}

// Uses Zep Go SDK v3
// API endpoint: https://api.getzep.com/api/v2
```

#### 6. Repository Interfaces
```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, userID string) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
}

type DocumentRepository interface {
    Create(ctx context.Context, doc *Document) error
    GetByID(ctx context.Context, docID string) (*Document, error)
    ListByUserID(ctx context.Context, userID string) ([]*Document, error)
    Update(ctx context.Context, doc *Document) error
}
```

## Data Models

### Database Schema

```sql
-- Users table
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

-- Documents table
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename VARCHAR(255),
    content_type VARCHAR(100),
    storage_key VARCHAR(500) NOT NULL, -- S3 key
    size_bytes BIGINT,
    source VARCHAR(50) NOT NULL, -- 'editor' or 'upload'
    status VARCHAR(50) DEFAULT 'processing', -- 'processing', 'completed', 'failed'
    zep_session_id VARCHAR(255), -- Zep session reference
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_documents_user ON documents(user_id);
CREATE INDEX idx_documents_status ON documents(status);

-- Password reset tokens table
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_reset_tokens_token ON password_reset_tokens(token);
```

### Go Models

```go
type User struct {
    ID            string    `json:"id" db:"id"`
    Email         string    `json:"email" db:"email"`
    PasswordHash  *string   `json:"-" db:"password_hash"`
    FirstName     *string   `json:"firstName" db:"first_name"`
    LastName      *string   `json:"lastName" db:"last_name"`
    OAuthProvider *string   `json:"oauthProvider" db:"oauth_provider"`
    OAuthID       *string   `json:"oauthId" db:"oauth_id"`
    CreatedAt     time.Time `json:"createdAt" db:"created_at"`
    UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}

type Document struct {
    ID           string    `json:"id" db:"id"`
    UserID       string    `json:"userId" db:"user_id"`
    Filename     *string   `json:"filename" db:"filename"`
    ContentType  *string   `json:"contentType" db:"content_type"`
    StorageKey   string    `json:"storageKey" db:"storage_key"`
    SizeBytes    int64     `json:"sizeBytes" db:"size_bytes"`
    Source       string    `json:"source" db:"source"` // "editor" or "upload"
    Status       string    `json:"status" db:"status"`
    ZepSessionID *string   `json:"zepSessionId" db:"zep_session_id"`
    CreatedAt    time.Time `json:"createdAt" db:"created_at"`
    UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

type GraphData struct {
    Nodes []GraphNode `json:"nodes"`
    Edges []GraphEdge `json:"edges"`
}

type GraphNode struct {
    ID    string  `json:"id"`
    Label string  `json:"label"`
    Size  float64 `json:"size"`
    Color string  `json:"color"`
}

type GraphEdge struct {
    ID     string  `json:"id"`
    Source string  `json:"source"`
    Target string  `json:"target"`
    Label  *string `json:"label,omitempty"`
}
```

### TypeScript Types

```typescript
export interface User {
  id: string;
  email: string;
  firstName?: string;
  lastName?: string;
  oauthProvider?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Document {
  id: string;
  userId: string;
  filename?: string;
  contentType?: string;
  sizeBytes: number;
  source: 'editor' | 'upload';
  status: 'processing' | 'completed' | 'failed';
  createdAt: string;
  updatedAt: string;
}

export interface GraphData {
  nodes: GraphNode[];
  edges: GraphEdge[];
}

export interface GraphNode {
  id: string;
  label: string;
  size: number;
  color: string;
}

export interface GraphEdge {
  id: string;
  source: string;
  target: string;
  label?: string;
}
```

## Error Handling

### Frontend Error Handling

```typescript
// Centralized error handling utility
export class APIError extends Error {
  constructor(
    public statusCode: number,
    public message: string,
    public details?: any
  ) {
    super(message);
  }
}

// API client with error handling
export async function apiCall<T>(
  endpoint: string,
  options?: RequestInit
): Promise<T> {
  const token = getJWTToken();
  
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token && { Authorization: `Bearer ${token}` }),
      ...options?.headers,
    },
  });

  if (!response.ok) {
    if (response.status === 401) {
      // Token expired or invalid
      clearJWTToken();
      window.location.href = '/signin';
    }
    
    const error = await response.json();
    throw new APIError(response.status, error.message, error.details);
  }

  return response.json();
}

// Component error boundaries
// - Wrap authenticated routes with error boundary
// - Display user-friendly error messages
// - Log errors to monitoring service
```

### Backend Error Handling

```go
// Custom error types
type AppError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
    Status  int    `json:"-"`
}

func (e *AppError) Error() string {
    return e.Message
}

// Common errors
var (
    ErrUnauthorized     = &AppError{Code: "UNAUTHORIZED", Message: "Authentication required", Status: 401}
    ErrForbidden        = &AppError{Code: "FORBIDDEN", Message: "Access denied", Status: 403}
    ErrNotFound         = &AppError{Code: "NOT_FOUND", Message: "Resource not found", Status: 404}
    ErrInvalidInput     = &AppError{Code: "INVALID_INPUT", Message: "Invalid input data", Status: 400}
    ErrInternalServer   = &AppError{Code: "INTERNAL_ERROR", Message: "Internal server error", Status: 500}
)

// Error handling middleware
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            
            var appErr *AppError
            if errors.As(err, &appErr) {
                c.JSON(appErr.Status, appErr)
                return
            }
            
            // Log unexpected errors
            log.Printf("Unexpected error: %v", err)
            c.JSON(500, ErrInternalServer)
        }
    }
}

// Service-level error handling
// - Wrap external service errors (S3, Zep) with context
// - Retry logic for transient failures (3 attempts with exponential backoff)
// - Circuit breaker for Zep API calls
```

## Testing Strategy

### Frontend Testing

1. **Unit Tests** (Jest + React Testing Library)
   - Component rendering and interactions
   - API client functions
   - JWT token management utilities
   - Form validation logic

2. **Integration Tests**
   - Authentication flows (sign-up, sign-in, OAuth)
   - Document submission (editor and upload)
   - Graph visualization rendering

3. **E2E Tests** (Playwright)
   - Complete user journeys
   - OAuth callback handling
   - Document processing workflow

### Backend Testing

1. **Unit Tests** (Go testing package)
   - Service layer business logic
   - Repository CRUD operations
   - Utility functions (JWT, chunking)
   - Middleware (auth validation)

2. **Integration Tests**
   - Database operations with test database
   - S3 storage operations (using MinIO for local testing)
   - API endpoint handlers

3. **Mock Testing**
   - Mock Zep SDK for testing without external API calls
   - Mock OAuth providers for authentication testing
   - Mock storage service for handler testing

### Test Data Strategy

- Use test fixtures for consistent test data
- Seed test database with known user/document data
- Use factory functions for creating test models
- Clean up test data after each test run

## Security Considerations

1. **Authentication**
   - JWT tokens with 24-hour expiration
   - Refresh token mechanism for extended sessions
   - Secure password hashing (bcrypt, cost factor 12)
   - OAuth state parameter validation

2. **Authorization**
   - Middleware validates JWT on all authenticated routes
   - User ID extracted from JWT claims
   - Document access restricted to owner

3. **Data Protection**
   - HTTPS only in production
   - Environment variables for all secrets
   - S3 bucket with private ACL
   - Database connection with SSL

4. **Input Validation**
   - File type whitelist for uploads
   - File size limits (50MB max)
   - Content sanitization before storage
   - SQL injection prevention (parameterized queries)

5. **Rate Limiting**
   - API rate limiting per user (100 req/min)
   - OAuth callback rate limiting
   - File upload rate limiting

## Deployment Considerations

### Environment Variables

```bash
# Backend (.env)
DATABASE_URL=postgresql://user:pass@host:5432/orgmind
JWT_SECRET=<random-secret>
JWT_EXPIRATION_HOURS=24

# AWS S3
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=<key>
AWS_SECRET_ACCESS_KEY=<secret>
AWS_S3_BUCKET=orgmind-documents

# Zep Cloud
ZEP_API_KEY=<zep-api-key>
ZEP_API_URL=https://api.getzep.com/api/v2

# OAuth
GOOGLE_CLIENT_ID=<client-id>
GOOGLE_CLIENT_SECRET=<secret>
OKTA_DOMAIN=<domain>
OKTA_CLIENT_ID=<client-id>
OKTA_CLIENT_SECRET=<secret>
OFFICE365_CLIENT_ID=<client-id>
OFFICE365_CLIENT_SECRET=<secret>
OAUTH_REDIRECT_URL=https://orgmind.com/auth/callback

# Frontend (.env.local)
NEXT_PUBLIC_API_URL=https://api.orgmind.com
NEXT_PUBLIC_OAUTH_REDIRECT_URL=https://orgmind.com/auth/callback
```

### Infrastructure

- **Frontend**: Vercel or AWS Amplify
- **Backend**: AWS ECS/Fargate or Google Cloud Run
- **Database**: AWS RDS PostgreSQL or Google Cloud SQL
- **Storage**: AWS S3
- **CDN**: CloudFront for static assets

### Monitoring

- Application logs (structured JSON logging)
- Error tracking (Sentry or similar)
- Performance monitoring (APM)
- Zep API usage tracking
- S3 storage metrics
