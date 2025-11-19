# Implementation Plan: AI-Powered Graph Chat with Gemini File Search

## Overview

This implementation plan breaks down the AI-powered chat feature into discrete, manageable tasks. The plan follows a bottom-up approach: database → backend services → API handlers → frontend components → integration.

---

## Phase 1: Database and Models

- [x] 1. Create database schema for chat
  - [x] 1.1 Create chat_threads table migration
    - Write up migration with id, graph_id, user_id, summary, timestamps
    - Add foreign key constraints to graphs and users tables
    - Create indexes on graph_id, user_id, and created_at
    - Write down migration for rollback
    - _Requirements: 6.1, 6.2, 6.3_

  - [x] 1.2 Create chat_messages table migration
    - Write up migration with id, thread_id, role, content, created_at
    - Add foreign key constraint to chat_threads table
    - Add CHECK constraint for role (user/assistant)
    - Create indexes on thread_id and created_at
    - Write down migration for rollback
    - _Requirements: 6.1, 6.2, 6.3_

  - [x] 1.3 Modify documents table for Gemini tracking
    - Add gemini_file_id column (VARCHAR 255, nullable)
    - Create index on gemini_file_id
    - Write down migration for rollback
    - _Requirements: 7.3_

  - [x] 1.4 Modify graphs table for File Search store tracking
    - Add gemini_store_id column (VARCHAR 255, nullable)
    - Create index on gemini_store_id
    - Write down migration for rollback
    - _Requirements: 7.2_

- [x] 2. Create Go models for chat
  - [x] 2.1 Create ChatThread model
    - Define struct with JSON and DB tags
    - Add validation methods
    - Create helper methods for summary generation
    - _Requirements: 2.2_

  - [x] 2.2 Create ChatMessage model
    - Define struct with JSON and DB tags
    - Add role validation (user/assistant)
    - Add content length validation
    - _Requirements: 3.2_

  - [x] 2.3 Update Document model
    - Add GeminiFileID field
    - Update JSON tags
    - _Requirements: 7.3_

  - [x] 2.4 Update Graph model
    - Add GeminiStoreID field
    - Update JSON tags
    - _Requirements: 7.2_

---

## Phase 2: Backend Repository Layer

- [x] 3. Implement chat repository
  - [x] 3.1 Create chat_repository.go interface
    - Define ChatRepository interface with all methods
    - Add method signatures for thread operations
    - Add method signatures for message operations
    - _Requirements: 6.1, 6.2_

  - [x] 3.2 Implement thread CRUD operations
    - Implement CreateThread with UUID generation
    - Implement GetThreadByID with error handling
    - Implement ListThreadsByGraphID with ordering
    - Implement UpdateThread for summary updates
    - Implement DeleteThread with cascade handling
    - _Requirements: 2.1, 2.2, 2.3_

  - [x] 3.3 Implement message CRUD operations
    - Implement CreateMessage with timestamp
    - Implement GetMessagesByThreadID with pagination
    - Implement DeleteMessagesByThreadID for cleanup
    - Add proper error wrapping
    - _Requirements: 3.1, 3.2_

  - [ ]* 3.4 Write repository unit tests
    - Test thread creation and retrieval
    - Test message pagination
    - Test foreign key constraints
    - Test concurrent operations
    - _Requirements: 6.1, 6.2_

---

## Phase 3: Gemini Service Integration

- [x] 4. Implement Gemini service
  - [x] 4.1 Create gemini_service.go interface
    - Define GeminiService interface
    - Add File Search store management methods
    - Add document upload methods
    - Add chat generation methods
    - _Requirements: 7.1, 7.2, 8.1_

  - [x] 4.2 Implement File Search store management
    - Implement CreateFileSearchStore using go-genai SDK
    - Implement GetFileSearchStore with caching
    - Implement DeleteFileSearchStore with cleanup
    - Add error handling and retries
    - Store store ID in graphs table
    - _Requirements: 7.2, 7.4_

  - [x] 4.3 Implement document upload to File Search
    - Implement UploadDocument method
    - Convert document content to appropriate format
    - Handle MIME type detection
    - Implement retry logic (3 attempts)
    - Store file ID in documents table
    - _Requirements: 7.1, 7.3, 7.4_

  - [x] 4.4 Implement streaming response generation
    - Implement GenerateStreamingResponse method
    - Query File Search store for relevant content
    - Call Gemini API with context
    - Stream response chunks to channel
    - Handle API errors gracefully
    - _Requirements: 8.1, 8.2, 8.3, 8.4_

  - [x] 4.5 Add configuration and error handling
    - Load API key from environment
    - Implement exponential backoff for retries
    - Create custom error types for Gemini errors
    - Add logging for all operations
    - _Requirements: 7.4, 10.1_

  - [ ]* 4.6 Write Gemini service unit tests
    - Mock go-genai SDK calls
    - Test store creation and retrieval
    - Test document upload with retries
    - Test streaming response handling
    - Test error scenarios
    - _Requirements: 7.1, 7.2, 8.1_

---

## Phase 4: Chat Service Layer

- [x] 5. Implement chat service
  - [x] 5.1 Create chat_service.go interface
    - Define ChatService interface
    - Add thread management methods
    - Add message management methods
    - Add AI interaction methods
    - _Requirements: 2.1, 3.1, 8.1_

  - [x] 5.2 Implement thread management
    - Implement CreateThread with user validation
    - Implement GetThread with access control
    - Implement ListThreads with filtering
    - Generate summary from first message
    - _Requirements: 2.1, 2.2, 2.3, 2.4_

  - [x] 5.3 Implement message management
    - Implement GetMessages with pagination
    - Implement SaveMessage with validation
    - Sanitize message content
    - Validate message length (max 4000 chars)
    - _Requirements: 3.1, 3.2, 12.4_

  - [x] 5.4 Implement AI response generation
    - Implement GenerateResponse method
    - Verify graph membership
    - Get File Search store ID
    - Call Gemini service for streaming response
    - Save user and assistant messages
    - Handle streaming errors
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

  - [x] 5.5 Add rate limiting
    - Implement rate limiter (20 messages/minute per user)
    - Store rate limit state in memory or Redis
    - Return appropriate error when limit exceeded
    - _Requirements: 12.5_

  - [ ]* 5.6 Write chat service unit tests
    - Test thread creation and retrieval
    - Test message validation
    - Test rate limiting
    - Test AI response generation flow
    - Mock Gemini service calls
    - _Requirements: 2.1, 3.1, 8.1_

---

## Phase 5: Document Processing Integration

- [x] 6. Integrate Gemini with document upload
  - [x] 6.1 Update document_service.go
    - Add GeminiService dependency injection
    - Add method to upload document to File Search
    - _Requirements: 14.1, 14.2_

  - [x] 6.2 Implement async File Search upload
    - Launch goroutine after document creation
    - Get or create File Search store for graph
    - Extract plain text content
    - Upload to Gemini File Search
    - Update document with gemini_file_id
    - Log errors without failing main flow
    - _Requirements: 14.1, 14.2, 14.3, 14.4, 14.5_

  - [x] 6.3 Handle upload failures gracefully
    - Continue with Zep processing if Gemini fails
    - Log detailed error information
    - Mark document for retry if needed
    - _Requirements: 14.4, 14.5_

  - [ ]* 6.4 Write integration tests
    - Test document upload triggers File Search upload
    - Test failure handling
    - Test concurrent uploads
    - _Requirements: 14.1, 14.2_

---

## Phase 6: API Handler Layer

- [x] 7. Implement chat HTTP handlers
  - [x] 7.1 Create chat_handler.go
    - Define ChatHandler struct with dependencies
    - Add constructor function
    - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5_

  - [x] 7.2 Implement CreateThread handler
    - Handle POST /api/graphs/:graphId/chat/threads
    - Extract user ID from JWT
    - Verify graph membership
    - Call chat service to create thread
    - Return thread JSON
    - _Requirements: 11.2, 12.1, 12.2_

  - [x] 7.3 Implement GetThreadMessages handler
    - Handle GET /api/graphs/:graphId/chat/threads/:threadId/messages
    - Parse pagination parameters (limit, offset)
    - Verify user access to thread
    - Call chat service to get messages
    - Return messages JSON with pagination metadata
    - _Requirements: 11.3, 12.1, 12.3_

  - [x] 7.4 Implement SendMessage handler
    - Handle POST /api/graphs/:graphId/chat/threads/:threadId/messages
    - Parse message content from request body
    - Validate content length
    - Check rate limit
    - Save user message
    - Return 202 Accepted with message ID
    - Trigger async AI response generation
    - _Requirements: 11.4, 12.4, 12.5_

  - [x] 7.5 Implement SSE streaming handler
    - Handle GET /api/graphs/:graphId/chat/stream
    - Set SSE headers (Content-Type: text/event-stream)
    - Parse threadId and messageId from query params
    - Verify user access
    - Create response channel
    - Call chat service to generate streaming response
    - Stream chunks as SSE events
    - Send "done" event on completion
    - Handle errors and send error events
    - Implement proper connection cleanup
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 11.5_

  - [x] 7.6 Add error handling and validation
    - Validate all input parameters
    - Return appropriate HTTP status codes
    - Format error responses consistently
    - Log all errors with context
    - _Requirements: 10.1, 10.4_

  - [ ]* 7.7 Write handler integration tests
    - Test all endpoints with valid requests
    - Test authentication and authorization
    - Test error scenarios
    - Test SSE streaming
    - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5_

---

## Phase 7: API Routes Configuration

- [x] 8. Configure API routes
  - [x] 8.1 Update router.go
    - Add chat handler initialization
    - Register chat routes under /api/graphs/:graphId/chat
    - Apply authentication middleware
    - _Requirements: 11.1, 12.2_

  - [x] 8.2 Add route definitions
    - POST /api/graphs/:graphId/chat/threads
    - GET /api/graphs/:graphId/chat/threads/:threadId/messages
    - POST /api/graphs/:graphId/chat/threads/:threadId/messages
    - GET /api/graphs/:graphId/chat/stream
    - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5_

---

## Phase 8: Frontend State Management

- [x] 9. Implement Zustand chat store
  - [x] 9.1 Create chatStore.ts
    - Define ChatStore interface
    - Initialize store with create()
    - Add state properties (messages, activeThreadId, etc.)
    - _Requirements: 9.1, 9.2_

  - [x] 9.2 Implement message actions
    - Implement setActiveThread action
    - Implement addMessage action
    - Implement updateStreamingMessage action
    - Implement finalizeStreamingMessage action
    - Use Map for efficient message lookup by threadId
    - _Requirements: 9.3, 9.4_

  - [x] 9.3 Implement loading and error actions
    - Implement setLoading action
    - Implement setError action
    - Implement clearMessages action
    - _Requirements: 9.2, 9.5_

  - [x] 9.4 Add selectors for performance
    - Create selector for active thread messages
    - Create selector for loading state
    - Create selector for streaming message
    - Optimize re-renders with shallow equality
    - _Requirements: 9.5_

  - [ ]* 9.5 Write store unit tests
    - Test all actions
    - Test selectors
    - Test state updates
    - _Requirements: 9.1, 9.2, 9.3_

---

## Phase 9: Frontend API Client

- [x] 10. Implement chat API client
  - [x] 10.1 Create chat.ts API module
    - Add createThread function
    - Add getThreadMessages function
    - Add sendMessage function
    - Use existing apiClient for requests
    - _Requirements: 11.1, 11.2, 11.3, 11.4_

  - [x] 10.2 Implement SSE client
    - Create connectChatStream function
    - Handle SSE connection lifecycle
    - Parse SSE events (chunk, done, error)
    - Implement reconnection logic
    - Return cleanup function
    - _Requirements: 5.1, 5.2, 10.2_

  - [x] 10.3 Add error handling
    - Handle network errors
    - Handle API errors
    - Implement exponential backoff for retries
    - _Requirements: 10.1, 10.2_

---

## Phase 10: Frontend Chat Components

- [x] 11. Create chat UI components
  - [x] 11.1 Create ChatInterface component
    - Create component file and interface
    - Set up component state (input value, loading)
    - Initialize SSE connection on mount
    - Clean up connection on unmount
    - Handle thread creation if needed
    - Integrate with Zustand store
    - _Requirements: 1.1, 4.1, 5.1_

  - [x] 11.2 Create ChatMessageList component
    - Create component file and props interface
    - Render messages from Zustand store
    - Differentiate user vs assistant messages
    - Display timestamps
    - Implement auto-scroll to latest message
    - Show streaming message with typing indicator
    - Handle empty state
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

  - [x] 11.3 Create ChatMessage component
    - Create component for individual messages
    - Style user messages (right-aligned, blue)
    - Style assistant messages (left-aligned, gray)
    - Display timestamp
    - Add markdown rendering for assistant messages
    - _Requirements: 3.3_

  - [x] 11.4 Create ChatInput component
    - Create component file and props interface
    - Add textarea with auto-resize
    - Handle Enter key to send (Shift+Enter for new line)
    - Add Send button
    - Disable during loading
    - Show character count
    - Clear input after send
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

  - [x] 11.5 Create TypingIndicator component
    - Create animated typing indicator
    - Show while AI is generating response
    - Use CSS animations for dots
    - _Requirements: 4.4_

  - [x] 11.6 Add error display components
    - Create ErrorMessage component
    - Display inline errors in chat
    - Add retry button for failed messages
    - Show connection status
    - _Requirements: 10.1, 10.2, 10.4_

---

## Phase 11: Frontend Layout Integration

- [x] 12. Update graph detail page layout
  - [x] 12.1 Modify GraphDetail component
    - Change layout to two-column grid
    - Set 40% width for documents column
    - Set 60% width for chat column
    - Add responsive breakpoint for mobile (stack vertically)
    - _Requirements: 1.1, 1.2, 1.3, 15.1_

  - [x] 12.2 Update page.tsx to include chat
    - Import ChatInterface component
    - Pass graphId prop to ChatInterface
    - Handle thread creation
    - Manage chat visibility state
    - _Requirements: 1.1, 1.4, 1.5_

  - [x] 12.3 Add responsive styles
    - Use Tailwind responsive classes
    - Stack columns on mobile (< 768px)
    - Make chat input sticky on mobile
    - Ensure touch-friendly button sizes
    - Handle virtual keyboard on mobile
    - _Requirements: 15.1, 15.2, 15.3, 15.4, 15.5_

  - [x] 12.4 Add loading and error states
    - Show skeleton loader while initializing
    - Display error message if chat fails to load
    - Add retry functionality
    - _Requirements: 10.1, 10.2_

---

## Phase 12: Styling and Polish

- [x] 13. Style chat interface
  - [x] 13.1 Design message bubbles
    - Style user messages (blue, right-aligned)
    - Style assistant messages (gray, left-aligned)
    - Add proper spacing and padding
    - Add subtle shadows
    - _Requirements: 3.3_

  - [x] 13.2 Style chat input area
    - Design textarea with border
    - Style Send button
    - Add focus states
    - Add disabled states
    - _Requirements: 4.1_

  - [x] 13.3 Add animations
    - Fade in new messages
    - Animate typing indicator
    - Smooth scroll to new messages
    - Add loading spinners
    - _Requirements: 3.5, 4.4_

  - [x] 13.4 Ensure accessibility
    - Add ARIA labels
    - Ensure keyboard navigation
    - Add focus indicators
    - Test with screen readers
    - _Requirements: 15.3_

---

## Phase 13: Testing and Quality Assurance

- [ ]* 14. End-to-end testing
  - [ ]* 14.1 Test complete chat flow
    - Test sending message and receiving response
    - Test SSE streaming
    - Test message persistence
    - Test pagination
    - _Requirements: 1.1, 3.1, 4.1, 5.1_

  - [ ]* 14.2 Test document integration
    - Upload document and verify File Search upload
    - Ask question about uploaded document
    - Verify AI uses document content in response
    - _Requirements: 7.1, 8.1, 14.1_

  - [ ]* 14.3 Test error scenarios
    - Test network failures
    - Test SSE reconnection
    - Test rate limiting
    - Test invalid inputs
    - _Requirements: 10.1, 10.2, 12.5_

  - [ ]* 14.4 Test multi-user scenarios
    - Test concurrent chat sessions
    - Test access control
    - Test thread isolation
    - _Requirements: 12.1, 12.3_

  - [ ]* 14.5 Test mobile responsiveness
    - Test on various screen sizes
    - Test touch interactions
    - Test virtual keyboard handling
    - _Requirements: 15.1, 15.2, 15.3, 15.4, 15.5_

---

## Phase 14: Documentation and Deployment

- [ ] 15. Documentation and deployment
  - [x] 15.1 Update API documentation
    - Document all chat endpoints
    - Add request/response examples
    - Document SSE event format
    - Add error codes and messages
    - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5_

  - [x] 15.2 Update environment configuration
    - Add GEMINI_API_KEY to .env.example
    - Add GEMINI_PROJECT_ID to .env.example
    - Add GEMINI_LOCATION to .env.example
    - Document required environment variables
    - _Requirements: 4.5_

  - [ ] 15.3 Create deployment guide
    - Document database migration steps
    - Document Gemini API setup
    - Add monitoring recommendations
    - Add scaling considerations
    - _Requirements: 13.1, 13.2_

  - [ ] 15.4 Add usage documentation
    - Create user guide for chat feature
    - Add screenshots
    - Document best practices
    - Add troubleshooting section
    - _Requirements: 1.1, 4.1_

---

## Notes

- Tasks marked with `*` are optional testing tasks that can be skipped for MVP
- Each task should be completed and tested before moving to the next
- Database migrations should be tested with both up and down migrations
- All API endpoints should be tested with Postman or similar tool before frontend integration
- SSE implementation requires careful connection management and cleanup
- Gemini API costs should be monitored during development and production
