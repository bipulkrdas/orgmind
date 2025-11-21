'use client';

import { useEffect, useRef } from 'react';
import { ChatMessage as ChatMessageType } from '@/lib/types';
import { ChatMessage } from './ChatMessage';
import { TypingIndicator } from './TypingIndicator';

interface ChatMessageListProps {
  messages: ChatMessageType[];
  isLoading: boolean;
  streamingMessage: string | null;
}

export function ChatMessageList({ messages, isLoading, streamingMessage }: ChatMessageListProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to latest message when new messages arrive or streaming updates
  useEffect(() => {
    if (messagesEndRef.current) {
      // Use requestAnimationFrame for smoother scrolling
      requestAnimationFrame(() => {
        messagesEndRef.current?.scrollIntoView({ 
          behavior: 'smooth',
          block: 'nearest'
        });
      });
    }
  }, [messages.length, streamingMessage]); // Scroll on message count change or streaming update

  // Handle empty state
  if (!messages || messages.length === 0) {
    return (
      <div 
        className="flex items-center justify-center h-full p-4"
        role="status"
        aria-label="Empty chat"
      >
        <div className="text-center max-w-md">
          <svg
            className="mx-auto h-10 w-10 text-gray-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z"
            />
          </svg>
          <h3 className="mt-3 text-base font-medium text-gray-900">Start a conversation</h3>
          <p className="mt-1 text-sm text-gray-500">
            Ask questions about your documents
          </p>
        </div>
      </div>
    );
  }

  return (
    <div 
      ref={containerRef} 
      className="p-4 space-y-3"
      role="log"
      aria-live="polite"
      aria-label="Chat messages"
    >
      {/* Render all messages */}
      {messages.map((message) => (
        <ChatMessage key={message.id} message={message} />
      ))}

      {/* Show streaming message with typing indicator */}
      {streamingMessage !== null && (
        <div className="flex justify-start animate-fadeIn">
          <div className="flex flex-col w-full items-start">
            {streamingMessage === '' ? (
              // Show typing indicator when streaming just started
              <TypingIndicator />
            ) : (
              // Show streaming content
              <div className="px-3 py-2 rounded-2xl bg-gray-100 text-gray-900 rounded-bl-sm shadow-sm w-full">
                <div className="text-sm whitespace-pre-wrap break-words prose prose-sm max-w-none leading-relaxed">
                  {streamingMessage}
                  <span className="inline-block w-0.5 h-4 bg-gray-600 ml-1 animate-pulse"></span>
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Scroll anchor - invisible element at the bottom */}
      <div ref={messagesEndRef} className="h-px" />
    </div>
  );
}
