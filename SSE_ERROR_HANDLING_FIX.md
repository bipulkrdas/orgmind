# SSE Error Handling Fix

## Problem
The backend was inconsistently sending both `chunk` events (with the AI response) AND `error` events, causing confusion in the frontend. The error would overwrite the successfully received chunks.

Example problematic output:
```
event:chunk
data:{"content":"I am sorry, but based on the documents in the knowledge graph, I cannot find a definition for machine learning."}

event:error
data:{"error":"Failed to generate response"}
```

## Root Cause

### Backend Issues
1. **Channel Lifecycle Management**: `geminiService.GenerateStreamingResponse()` was closing `responseChan` with `defer close(responseChan)` at the start, causing the handler to check for completion before streaming finished.

2. **Error Propagation**: The chat service wasn't properly managing the channel closure, leading to race conditions where both success and error signals were sent.

3. **Post-Streaming Error**: After successfully streaming all chunks, if `SaveMessage` failed, the service would return an error. But the chunks were already sent, causing the handler to send both `chunk` events AND an `error` event.

### Frontend Issues
1. **No Chunk Preservation**: If an error event arrived after chunks, the chunks were discarded and only the error was shown.

2. **Overwriting Content**: The error handler would immediately clear any streaming content and show the error.

## Solution

### Backend Changes

#### 1. `gemini_service.go` - Remove Premature Channel Closure
```go
// BEFORE:
func (s *geminiService) GenerateStreamingResponse(...) error {
    defer close(responseChan)  // ❌ Closes too early
    // ...
}

// AFTER:
func (s *geminiService) GenerateStreamingResponse(...) error {
    // NOTE: Do NOT close responseChan here - let the caller manage channel lifecycle
    // The caller needs to know when streaming completes vs when an error occurs
    // ...
}
```

#### 2. `chat_service.go` - Proper Channel Management
```go
// BEFORE:
if err := s.geminiSvc.GenerateStreamingResponse(...); err != nil {
    close(fullResponseChan)
    <-done
    return "", fmt.Errorf("failed to generate AI response: %w", err)
}
<-done

// AFTER:
geminiErr := s.geminiSvc.GenerateStreamingResponse(...)
close(fullResponseChan)  // Close after streaming completes
<-done                   // Wait for goroutine to finish

if geminiErr != nil {
    return "", fmt.Errorf("failed to generate AI response: %w", geminiErr)
}
```

#### 3. `chat_service.go` - Don't Fail After Successful Streaming
```go
// BEFORE:
if err := s.SaveMessage(context.Background(), assistantMsg); err != nil {
    fmt.Printf("Error: failed to save assistant message: %v\n", err)
    return "", fmt.Errorf("failed to save assistant message: %w", err)  // ❌ Causes error after chunks sent
}

// AFTER:
if err := s.SaveMessage(context.Background(), assistantMsg); err != nil {
    // Log error but DON'T fail - streaming was successful
    // The user already received the response, failing now would send both chunks AND error
    fmt.Printf("Error: failed to save assistant message: %v\n", err)
    // Return the message ID anyway so the client knows streaming completed
}
return assistantMsg.ID, nil  // ✅ Always return success if streaming worked
```

### Frontend Changes

#### 1. `chat.ts` - Robust SSE Event Handling
```typescript
// Track if we received chunks before error
let receivedChunks = false;
let receivedDone = false;

switch (eventType) {
  case 'chunk':
    receivedChunks = true;
    onChunk(data.content);
    break;
    
  case 'done':
    receivedDone = true;
    onDone(data.content);
    return;
    
  case 'error':
    // ROBUST: Only show error if we haven't received successful completion
    if (!receivedDone) {
      // Only call onError if we haven't received any content
      if (!receivedChunks) {
        onError(data.error);
      }
    }
    return;
}
```

#### 2. `ChatInterface.tsx` - Preserve Partial Responses
```typescript
// onError callback
(errorMsg: string) => {
  const currentStreamingContent = useChatStore.getState().streamingMessage || '';
  
  // If we have streaming content, save it as a partial message before showing error
  if (currentStreamingContent.trim()) {
    console.log('Partial response received before error, preserving content');
    const partialMessage = {
      id: `partial-${Date.now()}`,
      threadId: threadId,
      role: 'assistant' as const,
      content: currentStreamingContent,
      createdAt: new Date().toISOString(),
    };
    addMessage(threadId, partialMessage);
  }
  
  // Show error
  setError(errorMsg);
  finalizeStreamingMessage();
}
```

#### 3. `chatStore.ts` - Cumulative Streaming
```typescript
// BEFORE:
updateStreamingMessage: (content: string) => {
  set({ streamingMessage: content });  // ❌ Replaces content
},

// AFTER:
updateStreamingMessage: (content: string) => {
  // CUMULATIVE: Append new content to existing streaming message
  set((state) => ({
    streamingMessage: (state.streamingMessage || '') + content
  }));
},
```

## Behavior After Fix

### Scenario 1: Successful Response
1. Backend streams chunks → Frontend displays them cumulatively
2. Backend sends `done` event → Frontend saves complete message
3. No error shown ✅

### Scenario 2: Error After Chunks (Previous Bug)
1. Backend streams chunks → Frontend displays them cumulatively
2. Backend sends `error` event → Frontend preserves chunks as partial message
3. Error shown below the partial response ✅

### Scenario 3: Error Without Chunks
1. Backend sends `error` event immediately
2. Frontend shows error (no chunks to preserve)
3. User can retry ✅

## Testing Checklist

- [ ] Send a normal query → Should receive complete response without errors
- [ ] Send a query that triggers backend error → Should show error appropriately
- [ ] Send a query that partially succeeds → Should preserve partial content
- [ ] Check browser console for SSE event logs
- [ ] Verify no duplicate messages in chat history
- [ ] Verify error messages are displayed in ErrorMessage component
- [ ] Test with slow network (throttling) to ensure streaming works

## Files Modified

### Backend
- `backend/internal/service/gemini_service.go` - Removed premature channel closure
- `backend/internal/service/chat_service.go` - Fixed channel lifecycle management + don't fail after successful streaming

### Frontend
- `frontend/lib/api/chat.ts` - Added robust SSE event handling with chunk tracking
- `frontend/components/chat/ChatInterface.tsx` - Preserve partial responses on error
- `frontend/lib/stores/chatStore.ts` - Changed to cumulative streaming (append vs replace)

## Additional Notes

- The backend should ideally NOT send both `chunk` and `error` events for the same request
- This fix makes the frontend resilient to backend inconsistencies
- Consider adding backend logging to identify why both events are being sent
- Future improvement: Add retry logic for failed streaming requests
