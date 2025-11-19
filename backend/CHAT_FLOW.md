# Chat Flow Architecture

This document explains the correct flow for the AI-powered chat feature in OrgMind.

## Overview

The chat system uses a two-step process:
1. **Save user message** (POST request)
2. **Stream AI response** (SSE connection)

This separation allows for:
- Immediate user feedback (message saved)
- Real-time streaming of AI responses
- Better error handling
- Cleaner architecture

## Detailed Flow

### Step 1: User Sends Message

**Frontend:**
```typescript
// In ChatInterface.tsx
const userMessage = await sendMessage(graphId, threadId, content);
// userMessage contains: { id, threadId, role: "user", content, createdAt }
```

**API Call:**
```
POST /api/graphs/:graphId/chat/threads/:threadId/messages
Body: { "content": "What is this document about?" }
```

**Backend Handler:** `SendMessage()`
- Validates user access to thread
- Checks rate limits (20 messages/minute)
- Validates message content (max 4000 chars)
- Saves user message to database
- Returns saved message with ID

**Response:**
```json
{
  "id": "msg_abc123",
  "threadId": "thread_xyz",
  "role": "user",
  "content": "What is this document about?",
  "createdAt": "2024-01-15T10:30:00Z"
}
```

### Step 2: Frontend Opens SSE Stream

**Frontend:**
```typescript
// In ChatInterface.tsx
const cleanup = connectChatStream(
  graphId,
  threadId,
  userMessage.id,  // Pass the user message ID
  onChunk,         // Handle each content chunk
  onDone,          // Handle completion
  onError          // Handle errors
);
```

**API Call:**
```
GET /api/graphs/:graphId/chat/stream?threadId=thread_xyz&userMessageId=msg_abc123
Headers: Authorization: Bearer <token>
```

**Backend Handler:** `StreamResponse()`
- Validates user access to thread
- Retrieves the user message by ID
- Gets the graph's Gemini File Search store ID
- Calls `GenerateResponseForMessage()` service method
- Streams AI response chunks via SSE

### Step 3: AI Response Generation

**Service Method:** `GenerateResponseForMessage()`
```go
func (s *chatService) GenerateResponseForMessage(
    ctx context.Context,
    threadID string,
    userMessageID string,
    graphID string,
    responseChan chan<- string,
) (assistantMessageID string, err error)
```

**Process:**
1. Retrieve user message from database
2. Get graph's Gemini store ID
3. Call Gemini API with File Search
4. Stream response chunks to channel
5. Save complete assistant message to database
6. Return assistant message ID

### Step 4: Frontend Receives Chunks

**SSE Events:**

**Chunk Event:**
```
event: chunk
data: {"content": "This document discusses"}
```

**Chunk Event:**
```
event: chunk
data: {"content": " the architecture of"}
```

**Done Event:**
```
event: done
data: {"content": "msg_def456"}
```

**Error Event (if error occurs):**
```
event: error
data: {"error": "Rate limit exceeded"}
```

**Frontend Handling:**
```typescript
onChunk: (content) => {
  // Append to current message display
  setCurrentResponse(prev => prev + content);
}

onDone: (assistantMessageId) => {
  // Save complete message to state
  addMessage({
    id: assistantMessageId,
    role: "assistant",
    content: currentResponse,
    ...
  });
  setCurrentResponse("");
}

onError: (error) => {
  // Display error to user
  setError(error);
}
```

## Sequence Diagram

```
Frontend                Backend Handler           Chat Service              Gemini Service
   |                          |                         |                         |
   |--POST /messages--------->|                         |                         |
   |                          |--SaveUserMessage()----->|                         |
   |                          |                         |--Save to DB             |
   |                          |                         |<----                    |
   |<--User Message-----------|                         |                         |
   |                          |                         |                         |
   |--GET /stream------------>|                         |                         |
   |                          |--GenerateResponseForMessage()-->                  |
   |                          |                         |--Get user message       |
   |                          |                         |--Get store ID           |
   |                          |                         |--GenerateStreamingResponse()-->
   |                          |                         |                         |
   |<--SSE: chunk-------------|<--chunk-----------------|<--chunk-----------------|
   |<--SSE: chunk-------------|<--chunk-----------------|<--chunk-----------------|
   |<--SSE: chunk-------------|<--chunk-----------------|<--chunk-----------------|
   |                          |                         |--Save assistant message |
   |<--SSE: done--------------|<--assistantMessageID----|<--done------------------|
   |                          |                         |                         |
```

## Key Design Decisions

### Why Two Separate Endpoints?

1. **Immediate Feedback**: User sees their message saved immediately
2. **Error Handling**: Can handle save errors separately from generation errors
3. **Rate Limiting**: Can rate limit message sending independently
4. **Reconnection**: If SSE connection drops, can reconnect without resending message
5. **Scalability**: Can queue AI generation separately from message storage

### Why Pass User Message ID to Stream?

1. **Idempotency**: Can reconnect to same stream if connection drops
2. **Context**: Backend knows exactly which message to respond to
3. **History**: Can retrieve full conversation history for context
4. **Tracking**: Can track which user message triggered which AI response

### Why Not Generate Response in SendMessage?

1. **Timeout**: HTTP requests have timeouts, SSE doesn't
2. **Streaming**: Can't stream response chunks over regular HTTP POST
3. **User Experience**: User sees message immediately, then sees AI typing
4. **Error Recovery**: If generation fails, user message is still saved

## Error Handling

### Message Save Errors (Step 1)
- **Rate limit exceeded**: 429 Too Many Requests
- **Invalid content**: 400 Bad Request
- **Unauthorized**: 401/403
- **Server error**: 500 Internal Server Error

Frontend should display error and not proceed to streaming.

### Streaming Errors (Step 2)
- **Connection failed**: Retry with exponential backoff (max 3 attempts)
- **Generation error**: Display error message from SSE error event
- **Timeout**: Close connection and show timeout message
- **Client disconnect**: Backend cleans up resources

## Rate Limiting

- **Limit**: 20 messages per minute per user per thread
- **Checked in**: `SaveUserMessage()` service method
- **Response**: 429 Too Many Requests
- **Frontend**: Display countdown timer before allowing next message

## Database Schema

### chat_messages table
```sql
CREATE TABLE chat_messages (
    id VARCHAR(255) PRIMARY KEY,
    thread_id VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,  -- 'user' or 'assistant'
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (thread_id) REFERENCES chat_threads(id)
);
```

### Message Relationships
- Each user message triggers one assistant message
- Messages are ordered by `created_at`
- Thread contains alternating user/assistant messages

## Testing the Flow

### Manual Testing

1. **Send a message:**
   ```bash
   curl -X POST http://localhost:8080/api/graphs/graph_123/chat/threads/thread_456/messages \
     -H "Authorization: Bearer <token>" \
     -H "Content-Type: application/json" \
     -d '{"content": "What is this about?"}'
   ```

2. **Open SSE stream:**
   ```bash
   curl -N http://localhost:8080/api/graphs/graph_123/chat/stream?threadId=thread_456&userMessageId=msg_789 \
     -H "Authorization: Bearer <token>"
   ```

### Frontend Testing

1. Open chat interface
2. Send a message
3. Verify message appears immediately
4. Verify AI response streams in real-time
5. Verify both messages saved to history

## Common Issues

### Issue: "userMessageId query parameter is required"
**Cause**: Frontend not passing user message ID to stream
**Fix**: Ensure `connectChatStream()` receives message ID from `sendMessage()` response

### Issue: Stream never completes
**Cause**: `GenerateResponseForMessage()` not closing channel or returning
**Fix**: Ensure channel is closed and assistant message ID is returned

### Issue: Duplicate messages
**Cause**: Calling `GenerateResponse()` in both handlers
**Fix**: Only call in `StreamResponse()`, not in `SendMessage()`

### Issue: Message content not found
**Cause**: Trying to stream before message is saved
**Fix**: Ensure `sendMessage()` completes before calling `connectChatStream()`

## Future Enhancements

1. **Message Editing**: Allow users to edit their messages
2. **Regenerate Response**: Allow users to regenerate AI responses
3. **Message Reactions**: Add thumbs up/down for feedback
4. **Thread Summaries**: Auto-generate thread summaries
5. **Message Search**: Search within thread messages
6. **Export Thread**: Export conversation as PDF/Markdown
