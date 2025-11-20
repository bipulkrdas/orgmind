'use client';

import { ChatThread } from '@/lib/types';

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

/**
 * Format a date as relative time (e.g., "2 hours ago", "Yesterday")
 */
function formatRelativeTime(dateString: string): string {
  // Parse the date string - handle both ISO format and timezone offsets
  const date = new Date(dateString);
  
  // Check if date is valid
  if (isNaN(date.getTime())) {
    console.error('Invalid date string:', dateString);
    return 'Unknown';
  }
  
  // Debug logging to identify timezone issues
  // console.log('Date string:', dateString, '| Parsed date:', date.toISOString(), '| Now:', new Date().toISOString());
  
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  // Handle negative differences (future dates - shouldn't happen but be defensive)
  if (diffSecs < 0) {
    return 'Just now';
  }

  if (diffSecs < 60) {
    return 'Just now';
  } else if (diffMins < 60) {
    return `${diffMins} ${diffMins === 1 ? 'minute' : 'minutes'} ago`;
  } else if (diffHours < 24) {
    return `${diffHours} ${diffHours === 1 ? 'hour' : 'hours'} ago`;
  } else if (diffDays === 1) {
    return 'Yesterday';
  } else if (diffDays < 7) {
    return `${diffDays} days ago`;
  } else {
    // Format as date for older threads
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  }
}

/**
 * Get thread preview text (first 50 chars of summary or default text)
 */
function getThreadPreview(thread: ChatThread): string {
  if (thread.summary && thread.summary.trim()) {
    const preview = thread.summary.trim();
    return preview.length > 50 ? preview.substring(0, 50) + '...' : preview;
  }
  return 'New conversation';
}

export function ThreadList({
  threads,
  selectedThreadId,
  isLoading,
  error,
  onThreadSelect,
  onThreadsRefresh,
  onNewThread,
}: ThreadListProps) {
  // Loading skeleton with staggered animation and shimmer effect
  if (isLoading) {
    return (
      <div 
        className="flex flex-col h-full bg-white border-r border-gray-200"
        role="status"
        aria-label="Loading threads"
      >
        {/* Header with New Chat button skeleton */}
        <div className="p-3 border-b border-gray-200 flex-shrink-0">
          <div className="h-9 bg-gray-200 rounded-lg animate-shimmer"></div>
        </div>

        {/* Thread list skeleton with staggered fade-in and shimmer */}
        <div className="flex-1 overflow-y-auto p-2 space-y-1">
          {[1, 2, 3, 4, 5].map((i) => (
            <div
              key={i}
              className="p-3 rounded-lg opacity-0"
              style={{ 
                animation: 'fadeIn 0.5s ease-out forwards',
                animationDelay: `${i * 100}ms` 
              }}
            >
              <div className="h-4 bg-gray-200 rounded mb-2 w-full animate-shimmer"></div>
              <div className="h-3 bg-gray-100 rounded w-2/3 animate-shimmer" style={{ animationDelay: '0.1s' }}></div>
            </div>
          ))}
        </div>
        
        {/* Screen reader announcement */}
        <span className="sr-only">Loading conversation threads</span>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className="flex flex-col h-full bg-white border-r border-gray-200">
        {/* Header with New Chat button */}
        <div className="p-3 border-b border-gray-200 flex-shrink-0">
          <button
            onClick={onNewThread}
            className="w-full px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 active:bg-blue-800 transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
            aria-label="Start new chat"
          >
            + New Chat
          </button>
        </div>

        {/* Error message */}
        <div className="flex-1 flex flex-col items-center justify-center p-4 text-center">
          <div className="mb-4">
            <svg
              className="w-12 h-12 text-red-400 mx-auto"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              aria-hidden="true"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
              />
            </svg>
          </div>
          <h3 className="text-sm font-medium text-gray-900 mb-1">
            Failed to load threads
          </h3>
          <p className="text-xs text-gray-500 mb-4">{error}</p>
          <button
            onClick={onThreadsRefresh}
            className="px-4 py-2 bg-white text-sm font-medium text-gray-700 border border-gray-300 rounded-lg hover:bg-gray-50 active:bg-gray-100 transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
          >
            Try Again
          </button>
        </div>
      </div>
    );
  }

  // Empty state
  if (!threads || !Array.isArray(threads) || threads.length === 0) {
    return (
      <div className="flex flex-col h-full bg-white border-r border-gray-200">
        {/* Header with New Chat button */}
        <div className="p-3 border-b border-gray-200 flex-shrink-0">
          <button
            onClick={onNewThread}
            className="w-full px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 active:bg-blue-800 transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
            aria-label="Start new chat"
          >
            + New Chat
          </button>
        </div>

        {/* Empty state message */}
        <div className="flex-1 flex flex-col items-center justify-center p-4 text-center">
          <div className="mb-4">
            <svg
              className="w-12 h-12 text-gray-300 mx-auto"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              aria-hidden="true"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
              />
            </svg>
          </div>
          <h3 className="text-sm font-medium text-gray-900 mb-1">
            No conversations yet
          </h3>
          <p className="text-xs text-gray-500">
            Start a new chat to begin
          </p>
        </div>
      </div>
    );
  }

  // Thread list
  return (
    <nav
      className="flex flex-col h-full bg-white border-r border-gray-200"
      aria-label="Chat threads"
    >
      {/* Header with New Chat button */}
      <div className="p-3 border-b border-gray-200 flex-shrink-0">
        <button
          onClick={onNewThread}
          className="w-full px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 active:bg-blue-800 transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
          aria-label="Start new chat"
        >
          + New Chat
        </button>
      </div>

      {/* Thread list */}
      <div className="flex-1 overflow-y-auto p-2" style={{ minHeight: 0 }}>
        <ul role="list" className="space-y-1">
          {Array.isArray(threads) &&
            threads.map((thread, index) => {
              const isSelected = thread.id === selectedThreadId;
              const preview = getThreadPreview(thread);
              const timestamp = formatRelativeTime(thread.updatedAt);

              return (
                <li 
                  key={thread.id} 
                  role="listitem"
                  className="animate-fadeIn"
                  style={{ animationDelay: `${index * 30}ms` }}
                >
                  <button
                    onClick={() => onThreadSelect(thread.id)}
                    aria-selected={isSelected}
                    aria-label={`Thread: ${preview}, ${timestamp}`}
                    className={`w-full text-left p-3 rounded-lg transition-all duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-1 ${
                      isSelected
                        ? 'bg-blue-50 border border-blue-200 shadow-sm scale-[1.02]'
                        : 'hover:bg-gray-50 active:bg-gray-100 border border-transparent hover:scale-[1.01]'
                    }`}
                  >
                    {/* Thread preview */}
                    <div
                      className={`text-sm font-medium mb-1 line-clamp-2 transition-colors duration-150 ${
                        isSelected ? 'text-blue-900' : 'text-gray-900'
                      }`}
                    >
                      {preview}
                    </div>

                    {/* Timestamp */}
                    <div
                      className={`text-xs transition-colors duration-150 ${
                        isSelected ? 'text-blue-600' : 'text-gray-500'
                      }`}
                    >
                      {timestamp}
                    </div>
                  </button>
                </li>
              );
            })}
        </ul>
      </div>
    </nav>
  );
}
