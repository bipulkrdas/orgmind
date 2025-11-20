# Design Document

## Overview

This design implements a two-column chat interface with a thread list sidebar. The left column displays all chat threads for a graph, while the right column shows the selected thread's conversation or a new message input for creating threads. The design follows the existing OrgMind architecture patterns and integrates seamlessly with the current chat implementation.

## Architecture

### Component Hierarchy

```
ChatInterface (Container)
├── ThreadList (Left Column)
│   ├── ThreadListHeader
│   ├── ThreadListItems
│   │   └── ThreadItem (repeating)
│   └── ThreadListEmpty
└── ChatPanel (Right Column)
    ├── NewThreadPrompt (when no thread selected)
    │   └── ChatInput
    └── ThreadConversation (when thread selected)
        ├── ChatMessageList
        └── ChatInput
```

### Layout Structure

**Desktop (≥768px):**
```
┌─────────────────────────────────────────┐
│  ThreadList  │  ChatPanel               │
│  (300px)     │  (flex-1)                │
│              │                           │
│  [Threads]   │  [Messages/New Thread]   │
│              │                           │
└─────────────────────────────────────────┘
```

**Mobile (<768px):**
```
┌─────────────────┐
│ [☰] Chat        │  ← Header with toggle
├─────────────────┤
│                 │
│  [Messages]     │  ← Full width
│                 │
└─────────────────┘

When sidebar open:
┌─────────────────┐
│  ThreadList     │  ← Overlay
│  [Threads]      │
│                 │
└─────────────────┘
```

## Components and Interfaces

### 1. ChatInterface (Modified)

**Purpose:** Main container that orchestrates thread list and chat panel

**Props:**
```typescript
interface ChatInterfaceProps {
  graphId: string;
  ready?: boolean; // Prevents initialization until parent is ready
}
```

**State:**
```typescript
{
  selectedThreadId: string | null;
  threads: ChatThread[];
  isLoadingThreads: boolean;
  threadsError: string | null;
  isMobileSidebarOpen: boolean;
}
```

**Key Changes:**
- Remove `threadId` and `onThreadCreate` props (managed internally now)
- Add thread list state management
- Handle thread selection
- Manage mobile sidebar toggle
- Coordinate between ThreadList and ChatPanel

### 2. ThreadList (New Component)

**Purpose:** Displays list of chat threads with selection state

**Props:**
```typescript
interface ThreadListProps {
  graphId: string;
  threads: ChatThread[];
  selectedThreadId: string | null;
  isLoading: boolean;
  error: string | null;
  onThreadSelect: (threadId: string) => void;
  onThreadsRefresh: () => void;
  onNewThread: () => void;
}
```

**Features:**
- Fetch and display threads on mount
- Show thread preview (first message or timestamp)
- Highlight selected thread
- Sort by most recent activity (updatedAt desc)
- Loading skeleton for initial load
- Error state with retry
- "New Chat" button at top
- Empty state when no threads exist

**Thread Item Display:**
```typescript
interface ThreadItemDisplay {
  id: string;
  preview: string; // First 50 chars of first message or "New conversation"
  timestamp: string; // Relative time (e.g., "2 hours ago")
  isSelected: boolean;
}
```

### 3. ChatPanel (New Component)

**Purpose:** Right column that shows either new thread prompt or active conversation

**Props:**
```typescript
interface ChatPanelProps {
  graphId: string;
  selectedThreadId: string | null;
  onThreadCreated: (threadId: string) => void;
  onMessageSent: () => void; // Callback to refresh thread list
}
```

**Modes:**
1. **No Thread Selected:** Shows NewThreadPrompt
2. **Thread Selected:** Shows ThreadConversation

### 4. NewThreadPrompt (New Component)

**Purpose:** Welcome screen with input for starting new conversation

**Props:**
```typescript
interface NewThreadPromptProps {
  graphId: string;
  onThreadCreated: (threadId: string) => void;
}
```

**UI Elements:**
- Welcome message: "Start a new conversation"
- Subtitle: "Ask questions about your documents"
- ChatInput component
- Loading state during thread creation
- Error handling

**Behavior:**
1. User types first message
2. On send, call `createThread(graphId)` API
3. Then call `sendMessage(graphId, threadId, content)` API
4. Call `onThreadCreated(threadId)` to update parent
5. Parent selects the new thread and shows conversation

### 5. ThreadConversation (New Component)

**Purpose:** Wrapper for existing chat components when thread is selected

**Props:**
```typescript
interface ThreadConversationProps {
  graphId: string;
  threadId: string;
  onMessageSent: () => void;
}
```

**Contains:**
- ChatMessageList (existing)
- ChatInput (existing)
- Uses existing chat store and SSE streaming logic

## Data Models

### ChatThread (Existing)
```typescript
interface ChatThread {
  id: string;
  graphId: string;
  userId: string;
  summary: string | null;
  createdAt: string;
  updatedAt: string;
}
```

### ThreadListItem (Derived)
```typescript
interface ThreadListItem {
  id: string;
  preview: string; // Derived from first message or summary
  timestamp: string; // Formatted relative time
  isSelected: boolean;
}
```

## API Integration

### New Frontend API Function

Add to `frontend/lib/api/chat.ts`:

```typescript
/**
 * List all threads for a graph
 */
export async function listThreads(graphId: string): Promise<ChatThread[]> {
  const response = await apiCall<ChatThread[] | { threads: ChatThread[] }>(
    `/api/graphs/${graphId}/chat/threads`,
    { method: 'GET' }
  );

  // Defensive: Handle both direct array and wrapped response
  if (Array.isArray(response)) {
    return response;
  }
  
  if (response && typeof response === 'object' && 'threads' in response) {
    return Array.isArray(response.threads) ? response.threads : [];
  }
  
  return [];
}
```

### New Backend Handler

Add to `backend/internal/handler/chat_handler.go`:

```go
// ListThreads handles GET /api/graphs/:id/chat/threads
func (h *ChatHandler) ListThreads(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	graphID := c.Param("id")
	if graphID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Graph ID is required"})
		return
	}

	threads, err := h.chatService.ListThreads(c.Request.Context(), graphID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format
	response := make([]ChatThreadResponse, len(threads))
	for i, thread := range threads {
		response[i] = ChatThreadResponse{
			ID:        thread.ID,
			GraphID:   thread.GraphID,
			UserID:    thread.UserID,
			Summary:   thread.Summary,
			CreatedAt: thread.CreatedAt.Format(time.RFC3339),
			UpdatedAt: thread.UpdatedAt.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, response)
}
```

### New Backend Route

Add to `backend/internal/router/auth.go`:

```go
chat.GET("/threads", r.chatHandler.ListThreads)
```

## State Management

### Zustand Store Updates

Extend `frontend/lib/stores/chatStore.ts`:

```typescript
export interface ChatStore {
  // ... existing state ...
  
  // Thread list state
  threads: ChatThread[];
  selectedThreadId: string | null;
  
  // Thread list actions
  setThreads: (threads: ChatThread[]) => void;
  addThread: (thread: ChatThread) => void;
  selectThread: (threadId: string | null) => void;
  updateThreadTimestamp: (threadId: string) => void;
}
```

**Implementation:**
```typescript
// Thread list actions
setThreads: (threads: ChatThread[]) => {
  set({ threads });
},

addThread: (thread: ChatThread) => {
  set((state) => ({
    threads: [thread, ...state.threads] // Add to top
  }));
},

selectThread: (threadId: string | null) => {
  set({ 
    selectedThreadId: threadId,
    activeThreadId: threadId,
    streamingMessage: null,
    error: null 
  });
},

updateThreadTimestamp: (threadId: string) => {
  set((state) => ({
    threads: state.threads.map(t => 
      t.id === threadId 
        ? { ...t, updatedAt: new Date().toISOString() }
        : t
    ).sort((a, b) => 
      new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
    )
  }));
},
```

## Error Handling

### Thread List Errors
- **Network failure:** Show error banner with retry button
- **Empty state:** Show friendly message with "Start new chat" CTA
- **Permission denied:** Redirect to graphs list

### Thread Creation Errors
- **Network failure:** Show error message, keep user input
- **Validation error:** Show inline error below input
- **Timeout:** Show retry option

### Message Errors
- Use existing error handling from current ChatInterface

## Responsive Design

### Breakpoints
- **Mobile:** < 768px
- **Desktop:** ≥ 768px

### Mobile Behavior
1. **Default:** Show ChatPanel full width, ThreadList hidden
2. **Toggle:** Hamburger menu button in header
3. **Sidebar Open:** ThreadList overlays ChatPanel with backdrop
4. **Thread Select:** Auto-close sidebar, show conversation
5. **Back Button:** Show in ChatPanel header to return to thread list

### Desktop Behavior
1. **Fixed Layout:** ThreadList always visible (300px width)
2. **No Toggle:** Sidebar always shown
3. **Resizable:** Optional future enhancement

## Testing Strategy

### Unit Tests
1. **ThreadList Component:**
   - Renders loading state
   - Renders thread items correctly
   - Handles thread selection
   - Shows empty state
   - Handles errors with retry

2. **NewThreadPrompt Component:**
   - Renders welcome message
   - Handles message input
   - Creates thread on send
   - Shows loading state
   - Handles creation errors

3. **ChatPanel Component:**
   - Shows NewThreadPrompt when no thread selected
   - Shows ThreadConversation when thread selected
   - Passes correct props to children

4. **ChatInterface Integration:**
   - Loads threads on mount
   - Handles thread selection
   - Creates new threads
   - Updates thread list on new messages
   - Mobile sidebar toggle works

### Integration Tests
1. **Thread List API:**
   - Fetches threads successfully
   - Handles empty response
   - Handles API errors
   - Defensive response parsing

2. **Thread Creation Flow:**
   - Creates thread via API
   - Sends first message
   - Updates thread list
   - Selects new thread

3. **Message Flow:**
   - Sends message in selected thread
   - Updates thread timestamp
   - Maintains thread list order

### E2E Tests (Optional)
1. User opens chat interface
2. Sees existing threads
3. Selects a thread
4. Sends a message
5. Creates new thread
6. Switches between threads

## Performance Considerations

### Optimizations
1. **Thread List:**
   - Virtualize list if > 50 threads (react-window)
   - Debounce search/filter (future feature)
   - Cache thread list in store

2. **Message Loading:**
   - Lazy load messages when thread selected
   - Keep messages in store for quick switching
   - Clear old thread messages after threshold (e.g., 10 threads)

3. **Mobile:**
   - Use CSS transforms for sidebar animation
   - Lazy render ChatPanel content until thread selected

### Memory Management
- Clear messages for threads not accessed in 5 minutes
- Limit thread list to 100 most recent
- Implement pagination for older threads (future)

## Accessibility

### Keyboard Navigation
- Tab through thread list items
- Enter to select thread
- Escape to close mobile sidebar
- Arrow keys to navigate threads

### Screen Readers
- Announce thread selection
- Announce new messages
- Label all interactive elements
- Provide skip links

### ARIA Attributes
```html
<nav aria-label="Chat threads">
  <button aria-label="New chat" />
  <ul role="list">
    <li role="listitem" aria-selected="true">
      <button aria-label="Thread from 2 hours ago">
    </li>
  </ul>
</nav>
```

## Migration Strategy

### Phase 1: Backend (Minimal Changes)
1. Add `ListThreads` handler
2. Add route for `GET /api/graphs/:id/chat/threads`
3. Test endpoint

### Phase 2: Frontend Components
1. Create ThreadList component
2. Create NewThreadPrompt component
3. Create ChatPanel component
4. Create ThreadConversation wrapper

### Phase 3: Integration
1. Update ChatInterface to use new components
2. Add thread list state management
3. Update parent page.tsx to remove threadId prop
4. Test thread creation flow

### Phase 4: Polish
1. Add mobile responsive behavior
2. Add loading states and animations
3. Add error handling
4. Accessibility audit

## Future Enhancements

1. **Thread Search:** Filter threads by content
2. **Thread Rename:** Edit thread summary
3. **Thread Delete:** Remove threads
4. **Thread Archive:** Hide old threads
5. **Thread Sharing:** Share thread with other graph members
6. **Infinite Scroll:** Load more threads on scroll
7. **Thread Grouping:** Group by date (Today, Yesterday, Last Week)
8. **Keyboard Shortcuts:** Quick thread switching (Cmd+K)
