# Implementation Plan

- [x] 1. Add backend API endpoint for listing threads
  - Add `ListThreads` handler to `backend/internal/handler/chat_handler.go`
  - Add route `GET /api/graphs/:id/chat/threads` to `backend/internal/router/auth.go`
  - Return threads sorted by `updatedAt` descending
  - _Requirements: 1.1, 1.2_

- [x] 2. Add frontend API function for listing threads
  - Add `listThreads(graphId)` function to `frontend/lib/api/chat.ts`
  - Implement defensive response parsing (handle array or wrapped object)
  - Return empty array as fallback
  - _Requirements: 1.1, 1.2, 1.5_

- [x] 3. Extend Zustand chat store for thread management
  - Add `threads`, `selectedThreadId` state to `frontend/lib/stores/chatStore.ts`
  - Add `setThreads`, `addThread`, `selectThread`, `updateThreadTimestamp` actions
  - Add selectors for thread list state
  - _Requirements: 1.1, 2.1, 3.4, 6.1, 6.2, 6.3_

- [x] 4. Create ThreadList component
- [x] 4.1 Create base ThreadList component structure
  - Create `frontend/components/chat/ThreadList.tsx`
  - Implement props interface and component skeleton
  - Add loading skeleton UI
  - Add empty state UI
  - Add error state with retry button
  - _Requirements: 1.1, 1.3, 1.4, 1.5_

- [x] 4.2 Implement thread item rendering
  - Create thread item display logic
  - Show thread preview (first 50 chars or "New conversation")
  - Show relative timestamp (e.g., "2 hours ago")
  - Highlight selected thread
  - Add click handler for thread selection
  - _Requirements: 1.3, 2.1, 2.2_

- [x] 4.3 Add "New Chat" button
  - Add button at top of ThreadList
  - Call `onNewThread` callback when clicked
  - Style consistently with OrgMind design
  - _Requirements: 3.1, 3.2_

- [x] 5. Create NewThreadPrompt component
  - Create `frontend/components/chat/NewThreadPrompt.tsx`
  - Show welcome message and subtitle
  - Integrate ChatInput component
  - Handle first message submission
  - Call `createThread` API then `sendMessage` API
  - Show loading state during creation
  - Handle errors with user-friendly messages
  - Call `onThreadCreated` callback with new thread ID
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 5.1, 5.2, 5.3, 5.4, 5.5_

- [x] 6. Create ThreadConversation component
  - Create `frontend/components/chat/ThreadConversation.tsx`
  - Wrap existing ChatMessageList and ChatInput
  - Load messages for selected thread
  - Handle message sending with existing SSE logic
  - Call `onMessageSent` callback after sending
  - Reuse existing chat store logic
  - _Requirements: 2.3, 2.4, 2.5, 6.2_

- [x] 7. Create ChatPanel component
  - Create `frontend/components/chat/ChatPanel.tsx`
  - Conditionally render NewThreadPrompt or ThreadConversation
  - Show NewThreadPrompt when `selectedThreadId` is null
  - Show ThreadConversation when thread is selected
  - Pass appropriate props to child components
  - _Requirements: 2.1, 2.4, 3.1, 3.4_

- [x] 8. Refactor ChatInterface to use new components
- [x] 8.1 Update ChatInterface component structure
  - Modify `frontend/components/chat/ChatInterface.tsx`
  - Remove `threadId` and `onThreadCreate` props
  - Add state for `selectedThreadId`, `threads`, `isLoadingThreads`, `threadsError`
  - Implement two-column layout (ThreadList + ChatPanel)
  - _Requirements: 1.1, 2.1, 3.1_

- [x] 8.2 Implement thread list loading logic
  - Fetch threads on component mount using `listThreads` API
  - Store threads in component state
  - Handle loading and error states
  - Pass threads to ThreadList component
  - _Requirements: 1.1, 1.2, 1.4, 1.5_

- [x] 8.3 Implement thread selection logic
  - Handle thread selection from ThreadList
  - Update `selectedThreadId` state
  - Update Zustand store with `selectThread` action
  - Pass selected thread to ChatPanel
  - _Requirements: 2.1, 2.2, 6.4_

- [x] 8.4 Implement new thread creation logic
  - Handle "New Chat" button click (deselect thread)
  - Handle thread creation from NewThreadPrompt
  - Add new thread to threads list
  - Select newly created thread
  - Update thread list order
  - _Requirements: 3.3, 3.4, 3.5, 6.1, 6.3_

- [x] 8.5 Implement thread list updates
  - Update thread timestamp when message is sent
  - Re-sort threads by most recent activity
  - Maintain scroll position in thread list
  - _Requirements: 6.2, 6.3, 6.5_

- [x] 9. Add responsive mobile layout
- [x] 9.1 Implement mobile sidebar toggle
  - Add `isMobileSidebarOpen` state to ChatInterface
  - Add hamburger menu button in mobile header
  - Toggle sidebar visibility on button click
  - _Requirements: 4.2, 4.3_

- [x] 9.2 Implement mobile sidebar overlay
  - Show ThreadList as overlay on mobile when open
  - Add backdrop to close sidebar
  - Auto-close sidebar when thread is selected
  - Use CSS transforms for smooth animation
  - _Requirements: 4.2, 4.3, 4.4_

- [x] 9.3 Add responsive CSS
  - Use Tailwind breakpoints (md:768px)
  - Two-column layout on desktop
  - Full-width ChatPanel on mobile
  - Proper spacing and touch targets
  - _Requirements: 4.1, 4.2, 4.5_

- [x] 10. Update parent page component
  - Modify `frontend/app/(auth)/graphs/[graphId]/page.tsx`
  - Remove `threadId` state management
  - Remove `onThreadCreate` callback
  - Pass only `graphId` and `ready` props to ChatInterface
  - Simplify component logic
  - _Requirements: 1.1, 3.1_

- [x] 11. Add loading states and animations
  - Add skeleton loaders for thread list
  - Add loading indicators for thread creation
  - Add smooth transitions for thread selection
  - Add fade-in animations for new threads
  - Ensure loading states are accessible
  - _Requirements: 1.4, 5.1, 5.3_

- [x] 12. Add error handling and recovery
  - Handle thread list fetch errors with retry
  - Handle thread creation errors with user feedback
  - Handle message send errors (already implemented)
  - Show user-friendly error messages
  - Provide retry mechanisms
  - _Requirements: 1.5, 5.4, 5.5_

- [x] 13. Update component exports
  - Export new components from `frontend/components/chat/index.ts`
  - Ensure proper TypeScript types are exported
  - Update any import paths if needed
  - _Requirements: All_
