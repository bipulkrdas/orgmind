# Chat API Documentation

## Overview

The Chat API provides AI-powered conversational capabilities for knowledge graphs using Google Gemini with File Search. Users can ask questions about their documents and receive intelligent, context-aware responses streamed in real-time via Server-Sent Events (SSE).

## Base URL

```
http://localhost:8080/api/graphs/:graphId/chat
```

## Authentication

All chat endpoints require JWT authentication. Include the JWT token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## Access Control

- Users must be members of the graph to access chat functionality
- Thread access is restricted to the user who created the thread
- Rate limiting: 20 messages per minute per user

---

## Endpoints

### 1. Create Chat Thread

Creates a new chat thread for a graph.

**Endpoint:** `POST /api/graphs/:graphId/chat/threads`

**Path Parameters:**
- `graphId` (string, required): UUID of the graph

**Request Headers:**
```
Authorization: Bearer <jwt-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "summary": "Optional thread summary (max 200 characters)"
}
```

**Success Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "graphId": "123e4567-e89b-12d3-a456-426614174000",
  "userId": "789e0123-e89b-12d3-a456-426614174000",
  "summary": null,
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**

```json
// 400 Bad Request - Invalid graph ID
{
  "error": "Invalid graph ID format"
}

// 401 Unauthorized - Missing or invalid token
{
  "error": "Unauthorized"
}

// 403 Forbidden - User not a member of graph
{
  "error": "Access denied: user is not a member of this graph"
}

// 404 Not Found - Graph doesn't exist
{
  "error": "Graph not found"
}

// 500 Internal Server Error
{
  "error": "Failed to create chat thread"
}
```

**Example cURL:**
```bash
curl -X POST http://localhost:8080/api/graphs/123e4567-e89b-12d3-a456-426614174000/chat/threads \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{}'
```

---

### 2. Get Thread Messages

Retrieves messages for a specific chat thread with pagination support.

**Endpoint:** `GET /api/graphs/:graphId/chat/threads/:threadId/messages`

**Path Parameters:**
- `graphId` (string, required): UUID of the graph
- `threadId` (string, required): UUID of the chat thread

**Query Parameters:**
- `limit` (integer, optional): Number of messages to return (default: 50, max: 100)
- `offset` (integer, optional): Number of messages to skip (default: 0)

**Request Headers:**
```
Authorization: Bearer <jwt-token>
```

**Success Response (200 OK):**
```json
{
  "messages": [
    {
      "id": "msg-001",
      "threadId": "550e8400-e29b-41d4-a716-446655440000",
      "role": "user",
      "content": "What is this document about?",
      "createdAt": "2024-01-15T10:31:00Z"
    },
    {
      "id": "msg-002",
      "threadId": "550e8400-e29b-41d4-a716-446655440000",
      "role": "assistant",
      "content": "Based on the documents in your graph, this appears to be about...",
      "createdAt": "2024-01-15T10:31:05Z"
    }
  ],
  "total": 2,
  "limit": 50,
  "offset": 0,
  "hasMore": false
}
```

**Error Responses:**

```json
// 400 Bad Request - Invalid parameters
{
  "error": "Invalid limit parameter: must be between 1 and 100"
}

// 401 Unauthorized
{
  "error": "Unauthorized"
}

// 403 Forbidden - User doesn't have access to thread
{
  "error": "Access denied: thread belongs to another user"
}

// 404 Not Found - Thread doesn't exist
{
  "error": "Thread not found"
}

// 500 Internal Server Error
{
  "error": "Failed to retrieve messages"
}
```

**Example cURL:**
```bash
curl -X GET "http://localhost:8080/api/graphs/123e4567-e89b-12d3-a456-426614174000/chat/threads/550e8400-e29b-41d4-a716-446655440000/messages?limit=50&offset=0" \
  -H "Authorization: Bearer eyJhbGc..."
```

---

### 3. Send Message

Sends a user message to a chat thread and triggers AI response generation.

**Endpoint:** `POST /api/graphs/:graphId/chat/threads/:threadId/messages`

**Path Parameters:**
- `graphId` (string, required): UUID of the graph
- `threadId` (string, required): UUID of the chat thread

**Request Headers:**
```
Authorization: Bearer <jwt-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "content": "What are the key findings in the research paper?"
}
```

**Validation Rules:**
- `content` is required
- `content` must be between 1 and 4000 characters
- Rate limit: 20 messages per minute per user

**Success Response (202 Accepted):**
```json
{
  "messageId": "msg-003",
  "status": "processing",
  "message": "Message received and AI response is being generated"
}
```

**Error Responses:**

```json
// 400 Bad Request - Invalid content
{
  "error": "Message content is required"
}

// 400 Bad Request - Content too long
{
  "error": "Message content exceeds maximum length of 4000 characters"
}

// 401 Unauthorized
{
  "error": "Unauthorized"
}

// 403 Forbidden - User doesn't have access
{
  "error": "Access denied: thread belongs to another user"
}

// 404 Not Found - Thread doesn't exist
{
  "error": "Thread not found"
}

// 429 Too Many Requests - Rate limit exceeded
{
  "error": "Rate limit exceeded: maximum 20 messages per minute"
}

// 500 Internal Server Error
{
  "error": "Failed to process message"
}
```

**Example cURL:**
```bash
curl -X POST http://localhost:8080/api/graphs/123e4567-e89b-12d3-a456-426614174000/chat/threads/550e8400-e29b-41d4-a716-446655440000/messages \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{"content": "What are the key findings?"}'
```

**Notes:**
- This endpoint returns immediately with 202 Accepted
- The AI response is generated asynchronously
- Use the SSE streaming endpoint to receive the AI response in real-time

---

### 4. Stream AI Response (SSE)

Establishes a Server-Sent Events (SSE) connection to receive AI-generated responses in real-time.

**Endpoint:** `GET /api/graphs/:graphId/chat/stream`

**Path Parameters:**
- `graphId` (string, required): UUID of the graph

**Query Parameters:**
- `threadId` (string, required): UUID of the chat thread
- `messageId` (string, required): UUID of the user message to respond to

**Request Headers:**
```
Authorization: Bearer <jwt-token>
Accept: text/event-stream
```

**Response Headers:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
X-Accel-Buffering: no
```

**SSE Event Types:**

#### 1. Chunk Event
Sent for each piece of the AI response as it's generated.

```
event: chunk
data: {"content": "Based on"}

event: chunk
data: {"content": " the documents"}

event: chunk
data: {"content": " in your graph"}
```

#### 2. Done Event
Sent when the AI response is complete.

```
event: done
data: {"messageId": "msg-004", "status": "completed"}
```

#### 3. Error Event
Sent if an error occurs during response generation.

```
event: error
data: {"error": "Failed to generate response", "code": "GEMINI_API_ERROR"}
```

**Complete SSE Stream Example:**
```
event: chunk
data: {"content": "Based on"}

event: chunk
data: {"content": " the research"}

event: chunk
data: {"content": " paper you"}

event: chunk
data: {"content": " uploaded,"}

event: chunk
data: {"content": " the key"}

event: chunk
data: {"content": " findings are:"}

event: chunk
data: {"content": "\n\n1. The study"}

event: chunk
data: {"content": " demonstrates..."}

event: done
data: {"messageId": "msg-004", "status": "completed"}
```

**Error Responses:**

```json
// 400 Bad Request - Missing parameters
{
  "error": "Missing required parameter: threadId"
}

// 401 Unauthorized
{
  "error": "Unauthorized"
}

// 403 Forbidden - User doesn't have access
{
  "error": "Access denied: thread belongs to another user"
}

// 404 Not Found - Thread or message not found
{
  "error": "Thread not found"
}

// 500 Internal Server Error
{
  "error": "Failed to establish SSE connection"
}
```

**Example JavaScript Client:**
```javascript
const eventSource = new EventSource(
  `http://localhost:8080/api/graphs/${graphId}/chat/stream?threadId=${threadId}&messageId=${messageId}`,
  {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  }
);

eventSource.addEventListener('chunk', (event) => {
  const data = JSON.parse(event.data);
  console.log('Received chunk:', data.content);
  // Append to UI
});

eventSource.addEventListener('done', (event) => {
  const data = JSON.parse(event.data);
  console.log('Response complete:', data.messageId);
  eventSource.close();
});

eventSource.addEventListener('error', (event) => {
  const data = JSON.parse(event.data);
  console.error('Error:', data.error);
  eventSource.close();
});

eventSource.onerror = (error) => {
  console.error('SSE connection error:', error);
  eventSource.close();
};
```

**Example cURL (for testing):**
```bash
curl -N -H "Authorization: Bearer eyJhbGc..." \
  -H "Accept: text/event-stream" \
  "http://localhost:8080/api/graphs/123e4567-e89b-12d3-a456-426614174000/chat/stream?threadId=550e8400-e29b-41d4-a716-446655440000&messageId=msg-003"
```

**Connection Management:**
- The SSE connection remains open until the response is complete or an error occurs
- Client should handle reconnection if the connection drops
- Server automatically closes the connection after sending the `done` or `error` event
- Implement exponential backoff for reconnection attempts (recommended: 1s, 2s, 4s)

---

## Error Codes

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | OK - Request successful |
| 201 | Created - Resource created successfully |
| 202 | Accepted - Request accepted for processing |
| 400 | Bad Request - Invalid request parameters or body |
| 401 | Unauthorized - Missing or invalid authentication token |
| 403 | Forbidden - User doesn't have permission to access resource |
| 404 | Not Found - Requested resource doesn't exist |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error - Server-side error occurred |
| 503 | Service Unavailable - External service (Gemini) unavailable |

### Application Error Codes

| Code | Description | HTTP Status |
|------|-------------|-------------|
| `INVALID_GRAPH_ID` | Graph ID format is invalid | 400 |
| `INVALID_THREAD_ID` | Thread ID format is invalid | 400 |
| `INVALID_MESSAGE_ID` | Message ID format is invalid | 400 |
| `MISSING_CONTENT` | Message content is required | 400 |
| `CONTENT_TOO_LONG` | Message exceeds 4000 character limit | 400 |
| `INVALID_PAGINATION` | Invalid limit or offset parameter | 400 |
| `UNAUTHORIZED` | Missing or invalid JWT token | 401 |
| `ACCESS_DENIED` | User not authorized for this resource | 403 |
| `GRAPH_NOT_FOUND` | Graph doesn't exist | 404 |
| `THREAD_NOT_FOUND` | Thread doesn't exist | 404 |
| `MESSAGE_NOT_FOUND` | Message doesn't exist | 404 |
| `RATE_LIMIT_EXCEEDED` | Too many requests in time window | 429 |
| `GEMINI_API_ERROR` | Error communicating with Gemini API | 500 |
| `FILE_SEARCH_ERROR` | Error with File Search store | 500 |
| `DATABASE_ERROR` | Database operation failed | 500 |
| `INTERNAL_ERROR` | Unexpected server error | 500 |

---

## Rate Limiting

### Limits

- **Messages per user**: 20 messages per minute
- **SSE connections per user**: 5 concurrent connections

### Rate Limit Headers

When rate limited, the response includes:

```
X-RateLimit-Limit: 20
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1705318260
```

### Rate Limit Response

```json
{
  "error": "Rate limit exceeded: maximum 20 messages per minute",
  "retryAfter": 45
}
```

---

## Complete Usage Flow

### 1. Create a Thread

```bash
# Create thread
THREAD_ID=$(curl -s -X POST \
  http://localhost:8080/api/graphs/$GRAPH_ID/chat/threads \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}' | jq -r '.id')
```

### 2. Send a Message

```bash
# Send message
MESSAGE_ID=$(curl -s -X POST \
  http://localhost:8080/api/graphs/$GRAPH_ID/chat/threads/$THREAD_ID/messages \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"content": "What is this document about?"}' | jq -r '.messageId')
```

### 3. Stream the Response

```bash
# Stream AI response
curl -N -H "Authorization: Bearer $TOKEN" \
  -H "Accept: text/event-stream" \
  "http://localhost:8080/api/graphs/$GRAPH_ID/chat/stream?threadId=$THREAD_ID&messageId=$MESSAGE_ID"
```

### 4. Retrieve Message History

```bash
# Get all messages
curl -s -X GET \
  "http://localhost:8080/api/graphs/$GRAPH_ID/chat/threads/$THREAD_ID/messages?limit=50&offset=0" \
  -H "Authorization: Bearer $TOKEN" | jq
```

---

## Frontend Integration

### React/TypeScript Example

```typescript
import { useState, useEffect } from 'react';

interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  createdAt: string;
}

function ChatInterface({ graphId, token }: { graphId: string; token: string }) {
  const [threadId, setThreadId] = useState<string | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [streaming, setStreaming] = useState(false);

  // Create thread on mount
  useEffect(() => {
    async function createThread() {
      const response = await fetch(
        `http://localhost:8080/api/graphs/${graphId}/chat/threads`,
        {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({}),
        }
      );
      const data = await response.json();
      setThreadId(data.id);
    }
    createThread();
  }, [graphId, token]);

  // Send message and stream response
  async function sendMessage() {
    if (!threadId || !input.trim()) return;

    // Add user message to UI
    const userMessage: Message = {
      id: `temp-${Date.now()}`,
      role: 'user',
      content: input,
      createdAt: new Date().toISOString(),
    };
    setMessages(prev => [...prev, userMessage]);
    setInput('');

    // Send message to backend
    const response = await fetch(
      `http://localhost:8080/api/graphs/${graphId}/chat/threads/${threadId}/messages`,
      {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ content: input }),
      }
    );
    const { messageId } = await response.json();

    // Stream AI response
    setStreaming(true);
    const eventSource = new EventSource(
      `http://localhost:8080/api/graphs/${graphId}/chat/stream?threadId=${threadId}&messageId=${messageId}`,
      {
        headers: { 'Authorization': `Bearer ${token}` },
      }
    );

    let assistantContent = '';
    const assistantMessage: Message = {
      id: `temp-assistant-${Date.now()}`,
      role: 'assistant',
      content: '',
      createdAt: new Date().toISOString(),
    };
    setMessages(prev => [...prev, assistantMessage]);

    eventSource.addEventListener('chunk', (event) => {
      const data = JSON.parse(event.data);
      assistantContent += data.content;
      setMessages(prev => 
        prev.map(msg => 
          msg.id === assistantMessage.id 
            ? { ...msg, content: assistantContent }
            : msg
        )
      );
    });

    eventSource.addEventListener('done', (event) => {
      const data = JSON.parse(event.data);
      setMessages(prev => 
        prev.map(msg => 
          msg.id === assistantMessage.id 
            ? { ...msg, id: data.messageId }
            : msg
        )
      );
      setStreaming(false);
      eventSource.close();
    });

    eventSource.addEventListener('error', (event) => {
      console.error('SSE error:', event);
      setStreaming(false);
      eventSource.close();
    });
  }

  return (
    <div>
      <div className="messages">
        {messages.map(msg => (
          <div key={msg.id} className={`message ${msg.role}`}>
            {msg.content}
          </div>
        ))}
      </div>
      <input
        value={input}
        onChange={(e) => setInput(e.target.value)}
        disabled={streaming}
        onKeyPress={(e) => e.key === 'Enter' && sendMessage()}
      />
      <button onClick={sendMessage} disabled={streaming}>
        Send
      </button>
    </div>
  );
}
```

---

## Testing

### Using Postman

1. **Set up environment variables:**
   - `BASE_URL`: `http://localhost:8080`
   - `TOKEN`: Your JWT token
   - `GRAPH_ID`: A valid graph UUID

2. **Create a thread:**
   - Method: POST
   - URL: `{{BASE_URL}}/api/graphs/{{GRAPH_ID}}/chat/threads`
   - Headers: `Authorization: Bearer {{TOKEN}}`
   - Save `id` from response as `THREAD_ID`

3. **Send a message:**
   - Method: POST
   - URL: `{{BASE_URL}}/api/graphs/{{GRAPH_ID}}/chat/threads/{{THREAD_ID}}/messages`
   - Headers: `Authorization: Bearer {{TOKEN}}`
   - Body: `{"content": "Test message"}`
   - Save `messageId` from response

4. **Stream response (use curl or custom client):**
   - Postman doesn't support SSE well, use curl or browser

### Using curl

See the complete usage flow examples above.

---

## Troubleshooting

### SSE Connection Issues

**Problem:** SSE connection drops immediately

**Solutions:**
- Check that `Accept: text/event-stream` header is set
- Verify JWT token is valid and not expired
- Ensure threadId and messageId are valid UUIDs
- Check server logs for errors

**Problem:** No chunks received

**Solutions:**
- Verify the message was sent successfully (check 202 response)
- Check that Gemini API key is configured correctly
- Ensure File Search store exists for the graph
- Check server logs for Gemini API errors

### Rate Limiting

**Problem:** Getting 429 Too Many Requests

**Solutions:**
- Implement exponential backoff in client
- Check `X-RateLimit-Reset` header for retry time
- Reduce message frequency
- Consider implementing client-side rate limiting

### Authentication Errors

**Problem:** Getting 401 Unauthorized

**Solutions:**
- Verify JWT token is included in Authorization header
- Check token hasn't expired
- Ensure token format is `Bearer <token>`
- Verify user account is active

### Access Denied Errors

**Problem:** Getting 403 Forbidden

**Solutions:**
- Verify user is a member of the graph
- Check that thread belongs to the authenticated user
- Ensure graph exists and user has access

---

## Performance Considerations

### Pagination

- Use appropriate `limit` values (default: 50, max: 100)
- Implement cursor-based pagination for large message histories
- Cache message history on client side

### SSE Connections

- Limit concurrent SSE connections per user (max: 5)
- Implement connection pooling on client side
- Close connections when component unmounts
- Use exponential backoff for reconnection

### Message Content

- Keep messages under 4000 characters
- Consider chunking very long questions
- Sanitize and validate input on client side

---

## Security Best Practices

1. **Always use HTTPS in production**
2. **Validate JWT tokens on every request**
3. **Sanitize user input** to prevent injection attacks
4. **Implement rate limiting** to prevent abuse
5. **Log all API access** for audit trails
6. **Use environment variables** for sensitive configuration
7. **Rotate JWT secrets** regularly
8. **Monitor API usage** for anomalies

---

## Support

For issues or questions:
- Check server logs: `docker-compose logs backend`
- Review Gemini API status: https://status.cloud.google.com/
- Check Zep Cloud status: https://status.getzep.com/
- Contact support: support@orgmind.com

