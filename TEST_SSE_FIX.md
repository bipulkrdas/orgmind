# Testing SSE Error Handling Fix

## Quick Test Steps

### 1. Start the Backend
```bash
cd backend
go run cmd/server/main.go
```

### 2. Start the Frontend
```bash
cd frontend
npm run dev
```

### 3. Test Scenarios

#### Test 1: Normal Query (Should Work)
1. Navigate to a graph with documents
2. Open the chat interface
3. Send a query: "What is machine learning?"
4. **Expected**: Response streams in smoothly, no errors
5. **Check**: Browser console should show SSE events being received
6. **Check**: Message appears in chat history

#### Test 2: Query with No Results (Previous Bug Scenario)
1. Send a query about something NOT in your documents
2. **Expected**: Either a "no information found" response OR an error
3. **Check**: If chunks are received, they should be preserved
4. **Check**: Error (if any) should appear in ErrorMessage component, not overwrite chunks

#### Test 3: Network Interruption
1. Open browser DevTools → Network tab
2. Send a query
3. While streaming, throttle network to "Slow 3G"
4. **Expected**: Streaming continues (may be slower)
5. **Check**: No duplicate error events

### 4. Backend Logs to Monitor

Look for these log patterns in backend console:

```
[Gemini] Query Filtering: Starting query execution | Store: ... | Graph ID: ...
[Gemini] Query Filtering: Initiating streaming response for graph '...'
[Gemini] Query Filtering: SUCCESS - Streaming complete for graph '...' | Total chunks: X
```

### 5. Frontend Console Logs to Monitor

Look for these patterns in browser console:

```javascript
// Normal flow:
SSE chunk received: "content here"
SSE done received: "message-id-here"

// Error after chunks (should preserve chunks):
SSE chunk received: "partial content"
SSE error event received: "error message" (chunks were received before error)
Partial response received before error, preserving content

// Error without chunks:
SSE error event received: "error message" (no chunks received)
```

## Debugging Commands

### Check Backend SSE Stream
```bash
# Replace with your actual values
curl -N -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  "http://localhost:8080/api/graphs/GRAPH_ID/chat/stream?threadId=THREAD_ID&userMessageId=MESSAGE_ID"
```

Expected output:
```
event:chunk
data:{"content":"First part of response"}

event:chunk
data:{"content":" second part"}

event:done
data:{"content":"assistant-message-id"}
```

### Check for Error Events
If you see this pattern, there's still a backend issue:
```
event:chunk
data:{"content":"Response text"}

event:error
data:{"error":"Failed to generate response"}
```

## Success Criteria

✅ **Backend**:
- Only ONE terminal event per request (`done` OR `error`, never both)
- All chunks sent before terminal event
- Logs show "SUCCESS - Streaming complete"

✅ **Frontend**:
- Chunks accumulate in streaming message
- On success: Complete message saved to chat history
- On error after chunks: Partial message saved + error shown
- On error without chunks: Only error shown
- No duplicate messages in chat history

## Common Issues

### Issue: Still seeing both chunk and error events
**Solution**: Check `gemini_service.go` - ensure `responseChan` is NOT closed with `defer close(responseChan)` at the start of `GenerateStreamingResponse()`

### Issue: Chunks not accumulating
**Solution**: Check `chatStore.ts` - `updateStreamingMessage` should APPEND, not replace:
```typescript
streamingMessage: (state.streamingMessage || '') + content
```

### Issue: Error overwrites chunks
**Solution**: Check `ChatInterface.tsx` - error handler should preserve `currentStreamingContent` before calling `setError()`

### Issue: Duplicate messages
**Solution**: Check that `onDone` callback only adds message if content exists:
```typescript
if (completeContent.trim()) {
  addMessage(threadId, assistantMessage);
}
```

## Performance Check

Monitor these metrics:
- **Time to First Chunk**: Should be < 2 seconds
- **Streaming Rate**: Chunks should arrive smoothly (not in bursts)
- **Memory Usage**: Should not grow unbounded during long conversations
- **Connection Cleanup**: SSE connections should close properly (check Network tab)

## Rollback Plan

If issues persist:
1. Revert `gemini_service.go` changes
2. Revert `chat_service.go` changes
3. Keep frontend defensive changes (they're always beneficial)
4. Investigate backend channel management more deeply
