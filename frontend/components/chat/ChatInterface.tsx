'use client';

import { useEffect, useState, useCallback } from 'react';
import { useChatStore } from '@/lib/stores/chatStore';
import { listThreads } from '@/lib/api/chat';
import { ThreadList, ChatPanel, NetworkStatus } from '@/components/chat';
import type { ChatThread } from '@/lib/types';

interface ChatInterfaceProps {
  graphId: string;
  ready?: boolean; // Prevents initialization until parent is ready (avoids Neon DB race conditions)
}

export function ChatInterface({ graphId, ready = true }: ChatInterfaceProps) {
  // Component state for thread list
  const [threads, setThreadsState] = useState<ChatThread[]>([]);
  const [isLoadingThreads, setIsLoadingThreads] = useState(true);
  const [threadsError, setThreadsError] = useState<string | null>(null);
  
  // Subtask 9.1: Mobile sidebar toggle state
  const [isMobileSidebarOpen, setIsMobileSidebarOpen] = useState(false);

  // Zustand store state and actions
  const selectedThreadId = useChatStore((state) => state.selectedThreadId);
  const setThreadsStore = useChatStore((state) => state.setThreads);
  const selectThread = useChatStore((state) => state.selectThread);
  const updateThreadTimestamp = useChatStore((state) => state.updateThreadTimestamp);

  // Subtask 8.2: Fetch threads on component mount with retry logic
  const loadThreads = useCallback(async (retryCount = 0) => {
    // WORKAROUND: Wait for parent to finish loading to avoid Neon DB prepared statement race conditions
    if (!ready) {
      return;
    }

    const maxRetries = 3;
    const baseDelay = 1000; // 1 second

    try {
      setIsLoadingThreads(true);
      setThreadsError(null);

      const fetchedThreads = await listThreads(graphId);
      
      // Update both component state and Zustand store
      setThreadsState(fetchedThreads);
      setThreadsStore(fetchedThreads);
      setIsLoadingThreads(false);
    } catch (err) {
      console.error('Failed to load threads:', err);
      
      // Implement exponential backoff retry
      if (retryCount < maxRetries) {
        const delay = baseDelay * Math.pow(2, retryCount);
        console.log(`Retrying thread list fetch in ${delay}ms (attempt ${retryCount + 1}/${maxRetries})`);
        
        setTimeout(() => {
          loadThreads(retryCount + 1);
        }, delay);
      } else {
        // All retries exhausted, show user-friendly error
        const errorMessage = err instanceof Error 
          ? err.message.includes('fetch') || err.message.includes('network')
            ? 'Unable to connect to the server. Please check your internet connection and try again.'
            : err.message.includes('401') || err.message.includes('403')
            ? 'Your session has expired. Please refresh the page and sign in again.'
            : err.message
          : 'Unable to load conversations. Please try again.';
        
        setThreadsError(errorMessage);
        setIsLoadingThreads(false);
      }
    }
  }, [graphId, ready, setThreadsStore]);

  useEffect(() => {
    loadThreads();
  }, [loadThreads]);

  // Subtask 8.3: Handle thread selection
  const handleThreadSelect = useCallback((threadId: string) => {
    selectThread(threadId);
    // Subtask 9.2: Auto-close sidebar on mobile when thread is selected
    setIsMobileSidebarOpen(false);
  }, [selectThread]);

  // Subtask 8.4: Handle new thread creation
  const handleNewThread = useCallback(() => {
    // Deselect current thread to show NewThreadPrompt
    selectThread(null);
  }, [selectThread]);

  const handleThreadCreated = useCallback((threadId: string) => {
    // Reload threads to get the new thread from backend
    loadThreads().then(() => {
      // Select the newly created thread
      selectThread(threadId);
    });
  }, [loadThreads, selectThread]);

  // Subtask 8.5: Handle message sent (update thread timestamp)
  const handleMessageSent = useCallback(() => {
    if (selectedThreadId) {
      // Update thread timestamp and re-sort
      updateThreadTimestamp(selectedThreadId);
      
      // Also update local state to keep it in sync
      setThreadsState((prevThreads) => {
        const updated = prevThreads.map(t => 
          t.id === selectedThreadId 
            ? { ...t, updatedAt: new Date().toISOString() }
            : t
        );
        // Sort by most recent activity
        return updated.sort((a, b) => 
          new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
        );
      });
    }
  }, [selectedThreadId, updateThreadTimestamp]);

  // Subtask 9.1: Toggle mobile sidebar
  const toggleMobileSidebar = useCallback(() => {
    setIsMobileSidebarOpen((prev) => !prev);
  }, []);

  // Subtask 9.2: Close sidebar when clicking backdrop
  const closeMobileSidebar = useCallback(() => {
    setIsMobileSidebarOpen(false);
  }, []);

  // Subtask 8.1 & 9.3: Responsive two-column layout
  return (
    <>
      {/* Network status indicator */}
      <NetworkStatus />
      
      <div 
        className="flex h-full bg-white rounded-lg shadow overflow-hidden animate-fadeIn"
        role="region"
        aria-label="AI Chat Interface"
      >
      {/* Subtask 9.2: Mobile backdrop overlay with fade-in animation */}
      {isMobileSidebarOpen && (
        <div
          className="fixed inset-0 bg-black bg-opacity-50 z-40 md:hidden animate-fadeIn"
          onClick={closeMobileSidebar}
          aria-hidden="true"
          style={{ animationDuration: '200ms' }}
        />
      )}

      {/* Subtask 9.3: Left Column - Thread List (responsive) */}
      {/* Desktop: 300px fixed width, always visible */}
      {/* Mobile: Overlay sidebar with slide-in animation */}
      <div
        className={`
          w-80 flex-shrink-0 bg-white
          md:block md:relative md:z-auto
          fixed inset-y-0 left-0 z-50
          transform transition-transform duration-300 ease-in-out
          ${isMobileSidebarOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0'}
        `}
      >
        <ThreadList
          graphId={graphId}
          threads={threads}
          selectedThreadId={selectedThreadId}
          isLoading={isLoadingThreads}
          error={threadsError}
          onThreadSelect={handleThreadSelect}
          onThreadsRefresh={loadThreads}
          onNewThread={handleNewThread}
        />
      </div>

      {/* Subtask 9.3: Right Column - Chat Panel (responsive, full-width on mobile) */}
      <div className="flex-1 flex flex-col" style={{ minWidth: 0 }}>
        {/* Subtask 9.1: Mobile header with hamburger menu button */}
        <div className="md:hidden flex items-center px-4 py-3 border-b border-gray-200 bg-white">
          <button
            onClick={toggleMobileSidebar}
            className="p-2 rounded-lg hover:bg-gray-100 active:bg-gray-200 transition-all duration-200 touch-manipulation focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
            aria-label="Toggle thread list"
            aria-expanded={isMobileSidebarOpen}
            aria-controls="thread-list-sidebar"
          >
            <svg
              className={`w-6 h-6 text-gray-700 transition-transform duration-200 ${
                isMobileSidebarOpen ? 'rotate-90' : ''
              }`}
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M4 6h16M4 12h16M4 18h16"
              />
            </svg>
          </button>
          <h1 className="ml-3 text-lg font-semibold text-gray-900">Chat</h1>
        </div>

        <ChatPanel
          graphId={graphId}
          selectedThreadId={selectedThreadId}
          onThreadCreated={handleThreadCreated}
          onMessageSent={handleMessageSent}
        />
      </div>
    </div>
    </>
  );
}
