# Requirements Document: AI-Powered Graph Chat with Gemini File Search

## Introduction

This feature adds AI-powered chat capabilities to knowledge graphs, enabling users to ask questions about their documents and receive intelligent responses powered by Google Gemini with File Search. The system integrates document content into Gemini File Search stores and provides real-time streaming responses through Server-Sent Events (SSE).

## Glossary

- **Chat System**: The AI-powered conversational interface for querying graph documents
- **Chat Thread**: A conversation session containing multiple messages between user and AI
- **Chat Message**: A single message in a thread from either user or AI agent
- **Gemini File Search**: Google's AI service for semantic search across uploaded documents
- **File Search Store**: Google Gemini's storage container for document embeddings and search
- **SSE (Server-Sent Events)**: HTTP streaming protocol for real-time server-to-client updates
- **Zustand Store**: React state management library for chat message state
- **Graph Detail Page**: The page displaying graph information, documents, and chat interface
- **Document Upload Pipeline**: Background process that uploads documents to Gemini File Search

## Requirements

### Requirement 1: Chat Interface Layout

**User Story:** As a user, I want to see documents and chat side-by-side, so that I can reference documents while chatting with the AI.

#### Acceptance Criteria

1. WHEN the user navigates to a graph detail page, THE Chat System SHALL display a two-column layout with documents on the left and chat on the right
2. THE Chat System SHALL allocate 40% width to the document list column and 60% width to the chat column on desktop screens
3. WHEN the screen width is below 768px, THE Chat System SHALL stack the columns vertically with documents above chat
4. THE Chat System SHALL maintain the existing document list functionality in the left column
5. THE Chat System SHALL display a chat interface with message history and input box in the right column

### Requirement 2: Chat Thread Management

**User Story:** As a user, I want to start and manage chat conversations, so that I can organize my questions about different topics.

#### Acceptance Criteria

1. WHEN a user first opens a graph detail page, THE Chat System SHALL create a new chat thread automatically
2. THE Chat System SHALL store each chat thread with a unique thread_id, graph_id, user_id, summary, and timestamps
3. WHEN a user sends the first message, THE Chat System SHALL generate a summary from the first message (max 100 characters)
4. THE Chat System SHALL associate all messages with the active thread_id
5. THE Chat System SHALL display the thread summary in the chat header

### Requirement 3: Message Display and History

**User Story:** As a user, I want to see my conversation history, so that I can review previous questions and answers.

#### Acceptance Criteria

1. WHEN a chat thread loads, THE Chat System SHALL retrieve all messages for that thread from the database
2. THE Chat System SHALL display messages in chronological order with oldest at top
3. THE Chat System SHALL visually distinguish user messages from AI messages using different alignments and colors
4. THE Chat System SHALL display timestamps for each message
5. THE Chat System SHALL auto-scroll to the latest message when new messages arrive

### Requirement 4: Message Input and Sending

**User Story:** As a user, I want to type and send messages easily, so that I can ask questions about my documents.

#### Acceptance Criteria

1. THE Chat System SHALL provide a text input box fixed at the bottom of the chat column
2. WHEN the user types a message and presses Enter or clicks Send, THE Chat System SHALL send the message to the backend
3. THE Chat System SHALL disable the input box while waiting for AI response
4. THE Chat System SHALL display a loading indicator while the AI is generating a response
5. THE Chat System SHALL clear the input box after successfully sending a message

### Requirement 5: Real-Time Streaming Responses

**User Story:** As a user, I want to see AI responses as they are generated, so that I get immediate feedback.

#### Acceptance Criteria

1. WHEN the backend generates an AI response, THE Chat System SHALL stream the response using Server-Sent Events (SSE)
2. THE Chat System SHALL display partial responses as they arrive in real-time
3. THE Chat System SHALL append new text chunks to the current AI message without flickering
4. WHEN the stream completes, THE Chat System SHALL mark the message as complete
5. IF the SSE connection fails, THE Chat System SHALL display an error message and allow retry

### Requirement 6: Database Schema for Chat

**User Story:** As a system, I need to persist chat data, so that users can access their conversation history.

#### Acceptance Criteria

1. THE Chat System SHALL create a chat_threads table with columns: id, graph_id, user_id, summary, created_at, updated_at
2. THE Chat System SHALL create a chat_messages table with columns: id, thread_id, role (user/assistant), content, created_at
3. THE Chat System SHALL create a foreign key relationship between chat_messages.thread_id and chat_threads.id
4. THE Chat System SHALL create a foreign key relationship between chat_threads.graph_id and graphs.id
5. THE Chat System SHALL create indexes on graph_id, user_id, and thread_id for efficient queries

### Requirement 7: Gemini File Search Integration

**User Story:** As a system, I need to upload documents to Gemini File Search, so that the AI can answer questions based on document content.

#### Acceptance Criteria

1. WHEN a document is uploaded to a graph, THE Chat System SHALL upload the document content to a Gemini File Search store
2. THE Chat System SHALL create one File Search store per graph using the graph_id as identifier
3. THE Chat System SHALL store the Gemini file_id in the documents table for reference
4. THE Chat System SHALL handle upload failures gracefully and retry up to 3 times
5. THE Chat System SHALL process document uploads asynchronously without blocking the main upload flow

### Requirement 8: AI Response Generation

**User Story:** As a system, I need to generate intelligent responses, so that users get accurate answers from their documents.

#### Acceptance Criteria

1. WHEN a user sends a message, THE Chat System SHALL query the Gemini File Search store for relevant document content
2. THE Chat System SHALL use the Gemini API with the retrieved context to generate a response
3. THE Chat System SHALL include the graph's File Search store ID in the API request
4. THE Chat System SHALL stream the response back to the frontend using SSE
5. THE Chat System SHALL store both user message and AI response in the chat_messages table

### Requirement 9: State Management with Zustand

**User Story:** As a frontend developer, I need efficient state management, so that chat updates are performant and reliable.

#### Acceptance Criteria

1. THE Chat System SHALL use Zustand to manage chat messages state on the graph detail page
2. THE Chat System SHALL store messages, loading state, and error state in the Zustand store
3. THE Chat System SHALL provide actions for adding messages, updating streaming messages, and clearing errors
4. THE Chat System SHALL persist the active thread_id in the store
5. THE Chat System SHALL optimize re-renders by using Zustand selectors

### Requirement 10: Error Handling and User Feedback

**User Story:** As a user, I want clear error messages, so that I understand what went wrong and how to fix it.

#### Acceptance Criteria

1. WHEN the Gemini API returns an error, THE Chat System SHALL display a user-friendly error message
2. WHEN the SSE connection fails, THE Chat System SHALL attempt to reconnect automatically
3. WHEN document upload to File Search fails, THE Chat System SHALL log the error and continue with Zep processing
4. THE Chat System SHALL display inline error messages in the chat interface
5. THE Chat System SHALL provide a retry button for failed messages

### Requirement 11: API Endpoints

**User Story:** As a system, I need well-defined API endpoints, so that frontend and backend communicate effectively.

#### Acceptance Criteria

1. THE Chat System SHALL provide GET /api/graphs/:graphId/chat/threads endpoint to list threads
2. THE Chat System SHALL provide POST /api/graphs/:graphId/chat/threads endpoint to create a new thread
3. THE Chat System SHALL provide GET /api/graphs/:graphId/chat/threads/:threadId/messages endpoint to retrieve messages
4. THE Chat System SHALL provide POST /api/graphs/:graphId/chat/threads/:threadId/messages endpoint to send a message
5. THE Chat System SHALL provide GET /api/graphs/:graphId/chat/stream endpoint for SSE streaming responses

### Requirement 12: Security and Access Control

**User Story:** As a system, I need to enforce access control, so that users can only access their own chat data.

#### Acceptance Criteria

1. THE Chat System SHALL verify that the user is a member of the graph before allowing chat access
2. THE Chat System SHALL validate JWT tokens for all chat API requests
3. THE Chat System SHALL prevent users from accessing other users' chat threads
4. THE Chat System SHALL sanitize user input to prevent injection attacks
5. THE Chat System SHALL rate-limit chat requests to prevent abuse (max 20 messages per minute per user)

### Requirement 13: Performance and Scalability

**User Story:** As a system, I need to handle multiple concurrent chats, so that the platform scales with user growth.

#### Acceptance Criteria

1. THE Chat System SHALL process document uploads to File Search asynchronously in background goroutines
2. THE Chat System SHALL use connection pooling for database queries
3. THE Chat System SHALL implement pagination for message history (50 messages per page)
4. THE Chat System SHALL cache File Search store IDs to avoid repeated lookups
5. THE Chat System SHALL handle SSE connections efficiently with proper timeout and cleanup

### Requirement 14: Document Processing Pipeline

**User Story:** As a system, I need to automatically upload documents to File Search, so that they are immediately available for chat.

#### Acceptance Criteria

1. WHEN a document is created via editor or upload, THE Chat System SHALL trigger File Search upload in a background goroutine
2. THE Chat System SHALL extract plain text content for File Search upload
3. THE Chat System SHALL update the document record with gemini_file_id after successful upload
4. THE Chat System SHALL continue with Zep processing even if File Search upload fails
5. THE Chat System SHALL log all File Search upload operations for debugging

### Requirement 15: Responsive Design

**User Story:** As a user on mobile, I want a usable chat interface, so that I can chat on any device.

#### Acceptance Criteria

1. WHEN the viewport width is below 768px, THE Chat System SHALL display documents and chat in a vertical stack
2. THE Chat System SHALL make the chat input box sticky at the bottom on mobile
3. THE Chat System SHALL ensure touch-friendly button sizes (minimum 44x44px)
4. THE Chat System SHALL optimize message rendering for mobile performance
5. THE Chat System SHALL handle virtual keyboard appearance gracefully on mobile devices
