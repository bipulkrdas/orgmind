'use client';

import { useEffect, useState, useCallback } from 'react';
import { useChatStore, selectActiveThreadMessages, selectIsLoading, selectStreamingMessage, selectError } from '@/lib/stores/chatStore';
import { createThread, getThreadMessages, sendMessage, connectChatStream } from '@/lib/api/chat';
import { ChatMessageList } from '@/components/chat/ChatMessageList';
import { ChatInput } from '@/components/chat/ChatInput';
import { ErrorMessage } from '@/components/chat/ErrorMessage';

interface ChatInterfaceProps {
  graphId: string;
  threadId: string | null;
  onThreadCreate: (threadId: string) => void;
  ready?: boolean; // Prevents initialization until parent is ready (avoids Neon DB race conditions)
}

export function ChatInterface({ graphId, threadId, onThreadCreate, ready = true }: ChatInterfaceProps) {
  const [isInitializing, setIsInitializing] = useState(true);
  const [sseCleanup, setSseCleanup] = useState<(() => void) | null>(null);

  // Zustand store selectors
  const messages = useChatStore(selectActiveThreadMessages);
  const isLoading = useChatStore(selectIsLoading);
  const streamingMessage = useChatStore(selectStreamingMessage);
  const error = useChatStore(selectError);

  // Zustand store actions
  const setActiveThread = useChatStore((state) => state.setActiveThread);
  const addMessage = useChatStore((state) => state.addMessage);
  const updateStreamingMessage = useChatStore((state) => state.updateStreamingMessage);
  const finalizeStreamingMessage = useChatStore((state) => state.finalizeStreamingMessage);
  const setLoading = useChatStore((state) => state.setLoading);
  const setError = useChatStore((state) => state.setError);

  // Initialize thread and load messages
  useEffect(() => {
    // WORKAROUND: Wait for parent to finish loading to avoid Neon DB prepared statement race conditions
    if (!ready) {
      return;
    }

    const initializeChat = async () => {
      try {
        setIsInitializing(true);
        setError(null);

        let currentThreadId = threadId;

        // Create thread if needed
        if (!currentThreadId) {
          const newThread = await createThread(graphId);
          currentThreadId = newThread.id;
          onThreadCreate(newThread.id);
        }

        // Set active thread in store
        setActiveThread(currentThreadId);

        // Load existing messages
        const response = await getThreadMessages(graphId, currentThreadId);
        
        // Add messages to store
        if (response.messages && response.messages.length > 0) {
          response.messages.forEach((msg) => {
            addMessage(currentThreadId!, msg);
          });
        }
      } catch (err) {
        console.error('Failed to initialize chat:', err);
        setError(err instanceof Error ? err.message : 'Failed to initialize chat');
      } finally {
        setIsInitializing(false);
      }
    };

    initializeChat();
  }, [graphId, threadId, onThreadCreate, setActiveThread, addMessage, setError, ready]);

  // Cleanup SSE connection on unmount
  useEffect(() => {
    return () => {
      if (sseCleanup) {
        sseCleanup();
      }
    };
  }, [sseCleanup]);

  // Handle sending a message
  const handleSendMessage = useCallback(async (content: string) => {
    if (!threadId || !content.trim()) return;

    try {
      setError(null);
      setLoading(true);

      // Send message to backend (saves user message and returns it)
      const userMessage = await sendMessage(graphId, threadId, content.trim());

      // Add user message to store
      addMessage(threadId, userMessage);

      // Initialize streaming message
      updateStreamingMessage('');

      // Connect to SSE stream for AI response
      const cleanup = connectChatStream(
        graphId,
        threadId,
        userMessage.id, // Pass the user message ID
        // onChunk: append content to streaming message
        (chunk: string) => {
          updateStreamingMessage(chunk);
        },
        // onDone: finalize the streaming message
        (assistantMessageId: string) => {
          // Get the complete streaming message content
          const completeContent = useChatStore.getState().streamingMessage || '';
          
          // Only add assistant message if we have content
          if (completeContent.trim()) {
            const assistantMessage = {
              id: assistantMessageId,
              threadId: threadId,
              role: 'assistant' as const,
              content: completeContent,
              createdAt: new Date().toISOString(),
            };
            addMessage(threadId, assistantMessage);
          }
          
          // Clear streaming state
          finalizeStreamingMessage();
        },
        // onError: handle streaming errors
        // ROBUST: If chunks were received before error, they're already displayed
        // Only show error if no content was received
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
      );

      setSseCleanup(() => cleanup);
    } catch (err) {
      console.error('Failed to send message:', err);
      setError(err instanceof Error ? err.message : 'Failed to send message');
      setLoading(false);
    }
  }, [graphId, threadId, addMessage, updateStreamingMessage, finalizeStreamingMessage, setLoading, setError]);

  // Handle retry on error
  const handleRetry = useCallback(() => {
    setError(null);
    // User can try sending the message again
  }, [setError]);

  if (isInitializing) {
    return (
      <div className="flex flex-col h-full bg-white rounded-lg shadow animate-fadeIn">
        {/* Skeleton Header */}
        <div className="px-4 sm:px-6 py-4 border-b border-gray-200">
          <div className="h-6 w-32 bg-gray-200 rounded animate-pulse mb-2"></div>
          <div className="h-4 w-48 bg-gray-200 rounded animate-pulse" style={{ animationDelay: '150ms' }}></div>
        </div>

        {/* Skeleton Messages */}
        <div className="flex-1 p-4 sm:p-6 space-y-4">
          <div className="flex justify-start animate-slideUp">
            <div className="max-w-xs">
              <div className="h-16 w-64 bg-gray-200 rounded-2xl animate-pulse"></div>
            </div>
          </div>
          <div className="flex justify-end animate-slideUp" style={{ animationDelay: '100ms' }}>
            <div className="max-w-xs">
              <div className="h-12 w-48 bg-blue-100 rounded-2xl animate-pulse"></div>
            </div>
          </div>
          <div className="flex justify-start animate-slideUp" style={{ animationDelay: '200ms' }}>
            <div className="max-w-xs">
              <div className="h-20 w-72 bg-gray-200 rounded-2xl animate-pulse"></div>
            </div>
          </div>
        </div>

        {/* Skeleton Input */}
        <div className="border-t border-gray-200 p-3 sm:p-4 bg-gray-50">
          <div className="flex items-end space-x-2 sm:space-x-3">
            <div className="flex-1 h-11 bg-gray-200 rounded-xl animate-pulse"></div>
            <div className="h-11 w-11 bg-gray-200 rounded-xl animate-pulse" style={{ animationDelay: '100ms' }}></div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div 
      className="flex flex-col h-full bg-white rounded-lg shadow animate-fadeIn"
      role="region"
      aria-label="AI Chat Interface"
    >
      {/* Compact Chat Header */}
      <header className="px-4 py-2 border-b border-gray-200 flex-shrink-0">
        <h2 className="text-sm font-semibold text-gray-900" id="chat-title">
          AI Chat
        </h2>
      </header>

      {/* Messages Area - Scrollable with fixed height */}
      <main 
        className="flex-1 overflow-y-auto"
        style={{ minHeight: 0 }} // Critical for flex scrolling
        aria-labelledby="chat-title"
        aria-describedby="chat-description"
      >
        <ChatMessageList
          messages={messages}
          isLoading={isLoading}
          streamingMessage={streamingMessage}
        />
      </main>

      {/* Error Display */}
      {error && (
        <div className="px-4 py-2 border-t border-gray-100 flex-shrink-0" role="alert" aria-live="assertive">
          <ErrorMessage message={error} onRetry={handleRetry} />
        </div>
      )}

      {/* Input Area - Fixed at bottom */}
      <footer className="border-t border-gray-200 flex-shrink-0 bg-white">
        <ChatInput
          onSend={handleSendMessage}
          disabled={isLoading}
          placeholder="Ask a question..."
        />
      </footer>
    </div>
  );
}
