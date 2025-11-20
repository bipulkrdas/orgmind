'use client';

import { useEffect, useState, useCallback } from 'react';
import { useChatStore, selectActiveThreadMessages, selectIsLoading, selectStreamingMessage, selectError } from '@/lib/stores/chatStore';
import { getThreadMessages, sendMessage, connectChatStream } from '@/lib/api/chat';
import { ChatMessageList } from './ChatMessageList';
import { ChatInput } from './ChatInput';
import { ErrorMessage } from './ErrorMessage';

interface ThreadConversationProps {
  graphId: string;
  threadId: string;
  onMessageSent: () => void;
}

/**
 * ThreadConversation Component
 * 
 * Wraps ChatMessageList and ChatInput to display and interact with a selected thread.
 * Loads messages for the thread, handles message sending with SSE streaming,
 * and reuses existing chat store logic.
 * 
 * Requirements: 2.3, 2.4, 2.5, 6.2
 */
export function ThreadConversation({ graphId, threadId, onMessageSent }: ThreadConversationProps) {
  const [isLoadingMessages, setIsLoadingMessages] = useState(true);
  const [sseCleanup, setSseCleanup] = useState<(() => void) | null>(null);

  // Zustand store selectors
  const messages = useChatStore(selectActiveThreadMessages);
  const isLoading = useChatStore(selectIsLoading);
  const streamingMessage = useChatStore(selectStreamingMessage);
  const error = useChatStore(selectError);

  // Zustand store actions
  const setActiveThread = useChatStore((state) => state.setActiveThread);
  const addMessage = useChatStore((state) => state.addMessage);
  const setMessages = useChatStore((state) => state.setMessages);
  const updateStreamingMessage = useChatStore((state) => state.updateStreamingMessage);
  const finalizeStreamingMessage = useChatStore((state) => state.finalizeStreamingMessage);
  const setLoading = useChatStore((state) => state.setLoading);
  const setError = useChatStore((state) => state.setError);
  const clearMessages = useChatStore((state) => state.clearMessages);

  // Load messages when thread changes with retry logic
  useEffect(() => {
    let retryCount = 0;
    const maxRetries = 3;
    const baseDelay = 1000;

    const loadMessages = async () => {
      try {
        setIsLoadingMessages(true);
        setError(null);

        // Set active thread in store
        setActiveThread(threadId);

        // Check if messages already exist in store (e.g., from NewThreadPrompt)
        const existingMessages = useChatStore.getState().messages.get(threadId);
        if (existingMessages && existingMessages.length > 0) {
          // Messages already loaded, skip API call
          console.log('Messages already in store, skipping API call');
          setIsLoadingMessages(false);
          return;
        }

        // Clear previous messages for this thread
        clearMessages(threadId);

        // Load messages from API
        const response = await getThreadMessages(graphId, threadId);
        
        // Set all messages at once (more efficient than adding one by one)
        if (response.messages && Array.isArray(response.messages)) {
          setMessages(threadId, response.messages);
        } else {
          setMessages(threadId, []);
        }
        
        setIsLoadingMessages(false);
      } catch (err) {
        console.error('Failed to load messages:', err);
        
        // Implement exponential backoff retry
        if (retryCount < maxRetries) {
          const delay = baseDelay * Math.pow(2, retryCount);
          retryCount++;
          console.log(`Retrying message load in ${delay}ms (attempt ${retryCount}/${maxRetries})`);
          
          setTimeout(() => {
            loadMessages();
          }, delay);
        } else {
          // All retries exhausted, show user-friendly error
          let errorMessage = 'Unable to load messages. Please try again.';
          
          if (err instanceof Error) {
            if (err.message.includes('fetch') || err.message.includes('network')) {
              errorMessage = 'Unable to connect to the server. Please check your internet connection.';
            } else if (err.message.includes('401') || err.message.includes('403')) {
              errorMessage = 'Your session has expired. Please refresh the page and sign in again.';
            } else if (err.message.includes('404')) {
              errorMessage = 'This conversation no longer exists.';
            } else {
              errorMessage = err.message;
            }
          }
          
          setError(errorMessage);
          setIsLoadingMessages(false);
        }
      }
    };

    loadMessages();
  }, [graphId, threadId, setActiveThread, setMessages, clearMessages, setError]);

  // Cleanup SSE connection on unmount or thread change
  useEffect(() => {
    return () => {
      if (sseCleanup) {
        sseCleanup();
      }
    };
  }, [sseCleanup, threadId]);

  // Handle sending a message with improved error handling
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
        userMessage.id,
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
          
          // Notify parent that message was sent (for thread list update)
          onMessageSent();
        },
        // onError: handle streaming errors with user-friendly messages
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
          
          // Show user-friendly error message
          let friendlyError = errorMsg;
          if (errorMsg.includes('fetch') || errorMsg.includes('network')) {
            friendlyError = 'Connection lost. Please check your internet connection and try again.';
          } else if (errorMsg.includes('timeout')) {
            friendlyError = 'The AI is taking longer than expected. Please try again.';
          } else if (errorMsg.includes('401') || errorMsg.includes('403')) {
            friendlyError = 'Your session has expired. Please refresh the page and sign in again.';
          }
          
          setError(friendlyError);
          finalizeStreamingMessage();
        }
      );

      setSseCleanup(() => cleanup);
    } catch (err) {
      console.error('Failed to send message:', err);
      
      // User-friendly error messages
      let errorMessage = 'Failed to send message. Please try again.';
      
      if (err instanceof Error) {
        if (err.message.includes('fetch') || err.message.includes('network')) {
          errorMessage = 'Unable to connect to the server. Please check your internet connection.';
        } else if (err.message.includes('401') || err.message.includes('403')) {
          errorMessage = 'Your session has expired. Please refresh the page and sign in again.';
        } else if (err.message.includes('timeout')) {
          errorMessage = 'The request took too long. Please try again.';
        } else {
          errorMessage = err.message;
        }
      }
      
      setError(errorMessage);
      setLoading(false);
    }
  }, [graphId, threadId, addMessage, updateStreamingMessage, finalizeStreamingMessage, setLoading, setError, onMessageSent]);

  // Handle retry on error - reload messages
  const handleRetry = useCallback(() => {
    setError(null);
    setIsLoadingMessages(true);
    
    // Trigger a re-load by clearing and re-fetching
    const loadMessages = async () => {
      try {
        clearMessages(threadId);
        const response = await getThreadMessages(graphId, threadId);
        
        if (response.messages && Array.isArray(response.messages)) {
          setMessages(threadId, response.messages);
        } else {
          setMessages(threadId, []);
        }
        setIsLoadingMessages(false);
      } catch (err) {
        console.error('Retry failed to load messages:', err);
        
        let errorMessage = 'Unable to load messages. Please try again.';
        if (err instanceof Error) {
          if (err.message.includes('fetch') || err.message.includes('network')) {
            errorMessage = 'Unable to connect to the server. Please check your internet connection.';
          } else if (err.message.includes('401') || err.message.includes('403')) {
            errorMessage = 'Your session has expired. Please refresh the page and sign in again.';
          }
        }
        
        setError(errorMessage);
        setIsLoadingMessages(false);
      }
    };
    
    loadMessages();
  }, [graphId, threadId, setMessages, clearMessages, setError]);

  // Loading state while fetching messages
  if (isLoadingMessages) {
    return (
      <div className="flex flex-col h-full bg-white">
        {/* Skeleton Messages */}
        <div className="flex-1 p-4 space-y-4 overflow-y-auto">
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
        <div className="border-t border-gray-200 p-3 bg-gray-50">
          <div className="flex items-end space-x-2">
            <div className="flex-1 h-11 bg-gray-200 rounded-xl animate-pulse"></div>
            <div className="h-11 w-11 bg-gray-200 rounded-xl animate-pulse" style={{ animationDelay: '100ms' }}></div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div 
      className="flex flex-col h-full bg-white"
      role="region"
      aria-label="Thread conversation"
    >
      {/* Messages Area - Scrollable */}
      <main 
        className="flex-1 overflow-y-auto"
        style={{ minHeight: 0 }}
        aria-label="Thread messages"
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
          placeholder="Type a message..."
        />
      </footer>
    </div>
  );
}
