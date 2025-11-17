# Design Document: AI-Powered Graph Chat with Gemini File Search

## Overview

This feature integrates Google Gemini's File Search capabilities with OrgMind's knowledge graphs to provide an AI-powered chat interface. Users can ask questions about their documents and receive intelligent, context-aware responses streamed in real-time. The system uses Server-Sent Events (SSE) for streaming, Zustand for state management, and follows a layered architecture pattern.

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend Layer                        │
│  ┌────────────────┐  ┌──────────────┐  ┌─────────────────┐ │
│  │ Graph Detail   │  │ Chat         │  │ Zustand Store   │ │
│  │ Page Component │──│ Components   │──│ (State Mgmt)    │ │
│  └────────────────┘  └──────────────┘  └─────────────────┘ │
│           │                  │                               │
│           └──────────────────┴───────── SSE / REST API      │
└───────────────────────────────┬─────────────────────────────┘
                                │
┌───────────────────────────────┴─────────────────────────────┐
│                        Backend Layer                         │
│  ┌────────────────┐  ┌──────────────┐  ┌─────────────────┐ │
│  │ Chat Handler   │  │ Chat Service │  │ Gemini Service  │ │
│  │ (HTTP/SSE)     │──│ (Business    │──│ (File Search)   │ │
│  └────────────────┘  │  Logic)      │  └─────────────────┘ │
│                      └──────┬───────┘                       │
│                             │                                │
│  ┌────────────────┐  ┌──────┴───────┐  ┌─────────────────┐ │
│  │ Chat           │  │ Document     │  │ Graph           │ │
│  │ Repository     │  │ Service      │  │ Service         │ │
│  └────────────────┘  └──────────────┘  └─────────────────┘ │
└───────────────────────────────┬─────────────────────────────┘
                                │
┌───────────────────────────────┴─────────────────────────────┐
│                      Data Layer                              │
│  ┌────────────────┐  ┌──────────────┐  ┌─────────────────┐ │
│  │ PostgreSQL     │  │ Gemini File  │  │ Zep Cloud       │ │
│  │ (Chat Data)    │  │ Search Store │  │ (Knowledge      │ │
│  │                │  │              │  │  Graph)         │ │
│  └────────────────┘  └──────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### Frontend Components

#### 1. GraphDetailPage (Modified)

**Location:** `frontend/app/(auth)/graphs/[graphId]/page.tsx`

**Responsibilities:**
- Manage two-column layout (documents + chat)
- Initialize SSE connection for chat streaming
- Coordinate between document list and chat interface
- Handle responsive layout switching

**Key Changes:**
```typescript
interface GraphDetailPageState {
  graph: Graph | null;
  documents: Document[];
  activeThreadId: string | null;
  showChat: boolean;
}
```

#### 2. ChatInterface Component

**Location:** `frontend/components/chat/ChatInterface.tsx`

**Responsibilities:**
- Display chat messages with proper styling
- Handle message input and submission
- Manage SSE connection lifecycle
- Display loading and error states

**Props:**
```typescript
interface ChatInterfaceProps {
  graphId: string;
  threadId: string | null;
  onThreadCreate: (threadId: string) => void;
}
```

#### 3. ChatMessageList Component

**Location:** `frontend/components/chat/ChatMessageList.tsx`

**Responsibilities:**
- Render message history
- Auto-scroll to latest message
- Display user vs AI messages differently
- Show timestamps and status indicators

**Props:**
```typescript
interface ChatMessageListProps {
  messages: ChatMessage[];
  isLoading: boolean;
  streamingMessage: string | null;
}
```

#### 4. ChatInput Component

**Location:** `frontend/components/chat/ChatInput.tsx`

**Responsibilities:**
- Capture user input
- Handle Enter key submission
- Disable during AI response
- Show character count

**Props:**
```typescript
interface ChatInputProps {
  onSend: (message: string) => void;
  disabled: boolean;
  placeholder?: string;
}
```

#### 5. Zustand Chat Store

**Location:** `frontend/lib/stores/chatStore.ts`

**State Structure:**
```typescript
interface ChatStore {
  // State
  messages: Map<string, ChatMessage[]>; // threadId -> messages
  activeThreadId: string | null;
  streamingMessage: string | null;
  isLoading: boolean;
  error: string | null;
  
  // Actions
  setActiveThread: (threadId: string) => void;
  addMessage: (threadId: string, message: ChatMessage) => void;
  updateStreamingMessage: (content: string) => void;
  finalizeStreamingMessage: () => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  clearMessages: (threadId: string) => void;
}
```

### Backend Components

#### 1. Chat Handler

**Location:** `backend/internal/handler/chat_handler.go`

**Responsibilities:**
- Handle HTTP requests for chat operations
- Manage SSE connections for streaming
- Validate user permissions
- Format responses

**Key Methods:**
```go
type ChatHandler struct {
    chatService    service.ChatService
    graphService   service.GraphService
}

// CreateThread handles POST /api/graphs/:graphId/chat/threads
func (h *ChatHandler) CreateThread(c *gin.Context)

// GetThreadMessages handles GET /api/graphs/:graphId/chat/threads/:threadId/messages
func (h *ChatHandler) GetThreadMessages(c *gin.Context)

// SendMessage handles POST /api/graphs/:graphId/chat/threads/:threadId/messages
func (h *ChatHandler) SendMessage(c *gin.Context)

// StreamResponse handles GET /api/graphs/:graphId/chat/stream
func (h *ChatHandler) StreamResponse(c *gin.Context)
```

#### 2. Chat Service

**Location:** `backend/internal/service/chat_service.go`

**Responsibilities:**
- Business logic for chat operations
- Coordinate with Gemini service
- Manage chat threads and messages
- Generate thread summaries

**Interface:**
```go
type ChatService interface {
    // Thread management
    CreateThread(ctx context.Context, graphID, userID string) (*models.ChatThread, error)
    GetThread(ctx context.Context, threadID, userID string) (*models.ChatThread, error)
    ListThreads(ctx context.Context, graphID, userID string) ([]*models.ChatThread, error)
    
    // Message management
    GetMessages(ctx context.Context, threadID string, limit, offset int) ([]*models.ChatMessage, error)
    SaveMessage(ctx context.Context, message *models.ChatMessage) error
    
    // AI interaction
    GenerateResponse(ctx context.Context, threadID, userMessage string, responseChan chan<- string) error
}
```

#### 3. Gemini Service

**Location:** `backend/internal/service/gemini_service.go`

**Responsibilities:**
- Interact with Google Gemini API
- Manage File Search stores
- Upload documents to File Search
- Query File Search and generate responses

**Interface:**
```go
type GeminiService interface {
    // File Search store management
    CreateFileSearchStore(ctx context.Context, graphID string) (string, error)
    GetFileSearchStore(ctx context.Context, graphID string) (string, error)
    DeleteFileSearchStore(ctx context.Context, storeID string) error
    
    // Document management
    UploadDocument(ctx context.Context, storeID, documentID string, content []byte, mimeType string) (string, error)
    DeleteDocument(ctx context.Context, storeID, fileID string) error
    
    // Chat interaction
    GenerateStreamingResponse(ctx context.Context, storeID, query string, responseChan chan<- string) error
}
```

#### 4. Chat Repository

**Location:** `backend/internal/repository/chat_repository.go`

**Responsibilities:**
- Database operations for chat data
- CRUD operations for threads and messages
- Efficient querying with pagination

**Interface:**
```go
type ChatRepository interface {
    // Thread operations
    CreateThread(ctx context.Context, thread *models.ChatThread) error
    GetThreadByID(ctx context.Context, threadID string) (*models.ChatThread, error)
    ListThreadsByGraphID(ctx context.Context, graphID string) ([]*models.ChatThread, error)
    UpdateThread(ctx context.Context, thread *models.ChatThread) error
    DeleteThread(ctx context.Context, threadID string) error
    
    // Message operations
    CreateMessage(ctx context.Context, message *models.ChatMessage) error
    GetMessagesByThreadID(ctx context.Context, threadID string, limit, offset int) ([]*models.ChatMessage, error)
    DeleteMessagesByThreadID(ctx context.Context, threadID string) error
}
```

## Data Models

### Database Schema

#### chat_threads Table

```sql
CREATE TABLE chat_threads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    graph_id UUID NOT NULL REFERENCES graphs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    summary VARCHAR(200),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    INDEX idx_chat_threads_graph_id (graph_id),
    INDEX idx_chat_threads_user_id (user_id),
    INDEX idx_chat_threads_created_at (created_at DESC)
);
```

#### chat_messages Table

```sql
CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thread_id UUID NOT NULL REFERENCES chat_threads(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant')),
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    INDEX idx_chat_messages_thread_id (thread_id),
    INDEX idx_chat_messages_created_at (created_at ASC)
);
```

#### documents Table (Modified)

Add new column for Gemini file tracking:

```sql
ALTER TABLE documents ADD COLUMN gemini_file_id VARCHAR(255);
CREATE INDEX idx_documents_gemini_file_id ON documents(gemini_file_id);
```

#### graphs Table (Modified)

Add new column for File Search store tracking:

```sql
ALTER TABLE graphs ADD COLUMN gemini_store_id VARCHAR(255);
CREATE INDEX idx_graphs_gemini_store_id ON graphs(gemini_store_id);
```

### Go Models

#### ChatThread Model

```go
type ChatThread struct {
    ID        string    `json:"id" db:"id"`
    GraphID   string    `json:"graphId" db:"graph_id"`
    UserID    string    `json:"userId" db:"user_id"`
    Summary   *string   `json:"summary" db:"summary"`
    CreatedAt time.Time `json:"createdAt" db:"created_at"`
    UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}
```

#### ChatMessage Model

```go
type ChatMessage struct {
    ID        string    `json:"id" db:"id"`
    ThreadID  string    `json:"threadId" db:"thread_id"`
    Role      string    `json:"role" db:"role"` // "user" or "assistant"
    Content   string    `json:"content" db:"content"`
    CreatedAt time.Time `json:"createdAt" db:"created_at"`
}
```

### TypeScript Types

```typescript
interface ChatThread {
  id: string;
  graphId: string;
  userId: string;
  summary: string | null;
  createdAt: string;
  updatedAt: string;
}

interface ChatMessage {
  id: string;
  threadId: string;
  role: 'user' | 'assistant';
  content: string;
  createdAt: string;
}

interface StreamEvent {
  type: 'chunk' | 'done' | 'error';
  content?: string;
  error?: string;
}
```

## API Endpoints

### REST Endpoints

#### 1. Create Chat Thread

```
POST /api/graphs/:graphId/chat/threads
Authorization: Bearer <token>

Response 201:
{
  "id": "thread-uuid",
  "graphId": "graph-uuid",
  "userId": "user-uuid",
  "summary": null,
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

#### 2. Get Thread Messages

```
GET /api/graphs/:graphId/chat/threads/:threadId/messages?limit=50&offset=0
Authorization: Bearer <token>

Response 200:
{
  "messages": [
    {
      "id": "msg-uuid",
      "threadId": "thread-uuid",
      "role": "user",
      "content": "What is this document about?",
      "createdAt": "2024-01-01T00:00:00Z"
    },
    {
      "id": "msg-uuid-2",
      "threadId": "thread-uuid",
      "role": "assistant",
      "content": "Based on the documents...",
      "createdAt": "2024-01-01T00:00:01Z"
    }
  ],
  "total": 2,
  "hasMore": false
}
```

#### 3. Send Message

```
POST /api/graphs/:graphId/chat/threads/:threadId/messages
Authorization: Bearer <token>
Content-Type: application/json

{
  "content": "What is this document about?"
}

Response 202:
{
  "messageId": "msg-uuid",
  "status": "processing"
}
```

### SSE Endpoint

#### Stream AI Response

```
GET /api/graphs/:graphId/chat/stream?threadId=<thread-uuid>&messageId=<msg-uuid>
Authorization: Bearer <token>
Accept: text/event-stream

Response (SSE stream):
event: chunk
data: {"content": "Based on"}

event: chunk
data: {"content": " the documents"}

event: chunk
data: {"content": " in your graph..."}

event: done
data: {"messageId": "assistant-msg-uuid"}
```

## Integration Flow

### Document Upload Flow

```
1. User uploads document
   ↓
2. Document Service saves to storage & DB
   ↓
3. Launch goroutine for Zep processing
   ↓
4. Launch goroutine for Gemini File Search upload
   ↓
5. Gemini Service:
   - Get or create File Search store for graph
   - Upload document content to store
   - Store gemini_file_id in documents table
   ↓
6. Document ready for chat queries
```

### Chat Message Flow

```
1. User types message and clicks Send
   ↓
2. Frontend sends POST to /api/chat/threads/:threadId/messages
   ↓
3. Chat Handler validates user access
   ↓
4. Chat Service saves user message to DB
   ↓
5. Chat Service calls Gemini Service
   ↓
6. Gemini Service:
   - Queries File Search store
   - Generates streaming response
   - Sends chunks to response channel
   ↓
7. Chat Handler streams via SSE to frontend
   ↓
8. Frontend updates UI in real-time
   ↓
9. On completion, save assistant message to DB
```

## Error Handling

### Frontend Error Handling

1. **SSE Connection Errors**
   - Display "Connection lost" message
   - Attempt automatic reconnection (3 retries with exponential backoff)
   - Provide manual retry button

2. **Message Send Errors**
   - Display inline error below input
   - Keep message in input for retry
   - Show specific error message from backend

3. **Loading States**
   - Show typing indicator for AI
   - Disable input during processing
   - Display progress for long operations

### Backend Error Handling

1. **Gemini API Errors**
   - Log detailed error information
   - Return user-friendly error message
   - Implement retry logic with exponential backoff
   - Fall back to generic response if File Search unavailable

2. **Database Errors**
   - Wrap errors with context
   - Return appropriate HTTP status codes
   - Log errors for debugging
   - Implement transaction rollback where needed

3. **File Search Upload Errors**
   - Log error but don't fail document upload
   - Mark document for retry
   - Continue with Zep processing
   - Notify user if document won't be available for chat

## Testing Strategy

### Unit Tests

1. **Frontend**
   - Zustand store actions and selectors
   - Chat component rendering
   - Message formatting utilities
   - SSE event parsing

2. **Backend**
   - Chat service business logic
   - Gemini service API interactions (mocked)
   - Repository database operations
   - Message validation and sanitization

### Integration Tests

1. **API Endpoints**
   - Thread creation and retrieval
   - Message sending and retrieval
   - SSE streaming functionality
   - Authentication and authorization

2. **Database Operations**
   - Thread and message CRUD
   - Foreign key constraints
   - Index performance
   - Pagination correctness

3. **Gemini Integration**
   - File Search store creation
   - Document upload
   - Query and response generation
   - Error handling

### End-to-End Tests

1. Complete chat flow from UI to database
2. Document upload to chat availability
3. Multi-user concurrent chat sessions
4. SSE reconnection scenarios
5. Mobile responsive behavior

## Performance Considerations

### Frontend Optimization

1. **Virtual Scrolling** for long message lists (react-window)
2. **Debounced Input** to prevent excessive API calls
3. **Memoization** of message components
4. **Lazy Loading** of old messages on scroll
5. **Optimistic Updates** for user messages

### Backend Optimization

1. **Connection Pooling** for database and Gemini API
2. **Caching** of File Search store IDs
3. **Batch Processing** for multiple document uploads
4. **Rate Limiting** to prevent API abuse
5. **Async Processing** for non-critical operations

### Database Optimization

1. **Indexes** on frequently queried columns
2. **Pagination** for message retrieval
3. **Archiving** of old threads
4. **Query Optimization** with EXPLAIN ANALYZE
5. **Connection Pooling** configuration

## Security Considerations

1. **Input Sanitization**
   - Escape HTML in messages
   - Validate message length (max 4000 characters)
   - Prevent SQL injection in queries

2. **Access Control**
   - Verify graph membership before chat access
   - Validate JWT tokens on all requests
   - Implement rate limiting per user

3. **Data Privacy**
   - Encrypt sensitive data at rest
   - Use HTTPS for all communications
   - Implement proper CORS policies

4. **API Key Management**
   - Store Gemini API key in environment variables
   - Rotate keys regularly
   - Monitor API usage and costs

## Deployment Considerations

1. **Environment Variables**
   ```
   GEMINI_API_KEY=<api-key>
   GEMINI_PROJECT_ID=<project-id>
   GEMINI_LOCATION=us-central1
   ```

2. **Database Migrations**
   - Run migrations before deployment
   - Test rollback procedures
   - Backup data before schema changes

3. **Monitoring**
   - Track SSE connection count
   - Monitor Gemini API usage and costs
   - Alert on error rate thresholds
   - Log chat interactions for debugging

4. **Scaling**
   - Horizontal scaling for SSE connections
   - Load balancing with sticky sessions
   - Database read replicas for queries
   - CDN for static assets
