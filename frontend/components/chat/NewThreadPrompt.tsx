'use client';

import { useState } from 'react';
import { ChatInput } from './ChatInput';
import { ChatMessageList } from './ChatMessageList';
import { createThread, sendMessage, connectChatStream } from '@/lib/api/chat';
import { useChatStore, selectActiveThreadMessages, selectStreamingMessage } from '@/lib/stores/chatStore';

interface NewThreadPromptProps {
  graphId: string;
  onThreadCreated: (threadId: string) => void;
}

export function NewThreadPrompt({ graphId, onThreadCreated }: NewThreadPromptProps) {
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [pendingMessage, setPendingMessage] = useState<string>('');

  // Zustand store selectors
  const messages = useChatStore(selectActiveThreadMessages);
  const streamingMessage = useChatStore(selectStreamingMessage);

  // Zustand store actions
  const addMessage = useChatStore((state) => state.addMessage);
  const updateStreamingMessage = useChatStore((state) => state.updateStreamingMessage);
  const finalizeStreamingMessage = useChatStore((state) => state.finalizeStreamingMessage);
  const setActiveThread = useChatStore((state) => state.setActiveThread);
  const setMessages = useChatStore((state) => state.setMessages);

  const handleSendFirstMessage = async (content: string) => {
    setIsCreating(true);
    setError(null);
    setPendingMessage(content); // Preserve message for retry

    try {
      // Step 1: Create the thread
      const thread = await createThread(graphId);
      const threadId = thread.id;

      // Step 2: Set as active thread and initialize empty messages
      setActiveThread(threadId);
      setMessages(threadId, []);

      // Step 3: Send the first message
      const userMessage = await sendMessage(graphId, threadId, content);

      // Step 4: Add user message to store
      addMessage(threadId, userMessage);

      // Step 5: Initialize streaming message
      updateStreamingMessage('');

      // Step 6: Connect to SSE stream for AI response
      connectChatStream(
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
          setIsCreating(false);
          
          // Step 7: Notify parent component (thread is now ready with messages)
          setPendingMessage(''); // Clear on success
          onThreadCreated(threadId);
        },
        // onError: handle streaming errors
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
          setIsCreating(false);
          
          // Still notify parent so thread appears in list, even if AI response failed
          onThreadCreated(threadId);
        }
      );
    } catch (err) {
      // Handle errors with user-friendly messages
      let errorMessage = 'Failed to create conversation. Please try again.';
      
      if (err instanceof Error) {
        if (err.message.includes('fetch') || err.message.includes('network')) {
          errorMessage = 'Unable to connect to the server. Please check your internet connection and try again.';
        } else if (err.message.includes('401') || err.message.includes('403')) {
          errorMessage = 'Your session has expired. Please refresh the page and sign in again.';
        } else if (err.message.includes('timeout')) {
          errorMessage = 'The request took too long. Please try again.';
        } else {
          errorMessage = err.message;
        }
      }
      
      setError(errorMessage);
      setIsCreating(false);
    }
  };

  // Retry with the pending message
  const handleRetry = () => {
    if (pendingMessage) {
      handleSendFirstMessage(pendingMessage);
    } else {
      setError(null);
    }
  };

  // Show messages if conversation has started
  const hasMessages = messages.length > 0 || streamingMessage !== null;

  return (
    <div className="flex flex-col h-full">
      {/* Show messages if conversation started, otherwise show welcome screen */}
      {hasMessages ? (
        <div className="flex-1 overflow-y-auto" style={{ minHeight: 0 }}>
          <ChatMessageList
            messages={messages}
            isLoading={isCreating}
            streamingMessage={streamingMessage}
          />
        </div>
      ) : (
        /* Welcome message area - centered */
        <div className="flex-1 flex items-center justify-center p-8">
        <div className="max-w-2xl w-full text-center space-y-4">
          {/* Welcome message with fade-in animation */}
          <h2 className="text-2xl font-semibold text-gray-900 animate-fadeIn">
            Start a new conversation
          </h2>
          
          {/* Subtitle with delayed fade-in */}
          <p 
            className="text-base text-gray-600 animate-fadeIn"
            style={{ animationDelay: '100ms' }}
          >
            Ask questions about your documents
          </p>

          {/* Icon with delayed fade-in and subtle pulse when creating */}
          <div 
            className="mt-8 mb-4 animate-fadeIn"
            style={{ animationDelay: '200ms' }}
          >
            <svg
              className={`mx-auto h-16 w-16 text-gray-400 transition-all duration-300 ${
                isCreating ? 'animate-pulse-subtle scale-110' : ''
              }`}
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              aria-hidden="true"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={1.5}
                d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z"
              />
            </svg>
          </div>

          {/* Loading indicator when creating thread */}
          {isCreating && (
            <div 
              className="mt-4 animate-fadeIn"
              role="status"
              aria-live="polite"
            >
              <div className="flex items-center justify-center space-x-2">
                <div className="w-2 h-2 bg-blue-600 rounded-full animate-bounce" style={{ animationDelay: '0ms' }}></div>
                <div className="w-2 h-2 bg-blue-600 rounded-full animate-bounce" style={{ animationDelay: '150ms' }}></div>
                <div className="w-2 h-2 bg-blue-600 rounded-full animate-bounce" style={{ animationDelay: '300ms' }}></div>
              </div>
              <p className="mt-2 text-sm text-gray-600">Creating your conversation...</p>
              <span className="sr-only">Creating conversation, please wait</span>
            </div>
          )}
        </div>
      </div>
      )}

      {/* Error message with slide-up animation and retry button */}
      {error && (
        <div className="px-4 pb-2 animate-slideUp">
          <div className="max-w-4xl mx-auto">
            <div 
              className="bg-red-50 border border-red-200 rounded-lg p-3 flex items-start space-x-2"
              role="alert"
              aria-live="assertive"
            >
              <svg
                className="h-5 w-5 text-red-600 flex-shrink-0 mt-0.5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              <div className="flex-1">
                <p className="text-sm text-red-800 mb-2">{error}</p>
                {pendingMessage && (
                  <button
                    onClick={handleRetry}
                    className="inline-flex items-center px-3 py-1.5 text-xs font-medium text-red-700 bg-red-100 rounded hover:bg-red-200 transition-colors focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2"
                    aria-label="Retry sending message"
                  >
                    <svg
                      className="h-4 w-4 mr-1"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      aria-hidden="true"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                      />
                    </svg>
                    Retry
                  </button>
                )}
              </div>
              <button
                onClick={() => {
                  setError(null);
                  setPendingMessage('');
                }}
                className="text-red-600 hover:text-red-800 transition-colors focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 rounded"
                aria-label="Dismiss error"
              >
                <svg
                  className="h-5 w-5"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Chat input at bottom */}
      <div className="border-t border-gray-200">
        <div className="max-w-4xl mx-auto">
          <ChatInput
            onSend={handleSendFirstMessage}
            disabled={isCreating}
            placeholder={isCreating ? 'Creating conversation...' : 'Type your first message...'}
          />
        </div>
      </div>
    </div>
  );
}
