'use client';

import { useState, useRef, useEffect, KeyboardEvent, ChangeEvent } from 'react';

interface ChatInputProps {
  onSend: (message: string) => void;
  disabled: boolean;
  placeholder?: string;
}

const MAX_CHARS = 4000;

export function ChatInput({ onSend, disabled, placeholder = 'Type a message...' }: ChatInputProps) {
  const [value, setValue] = useState('');
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Auto-resize textarea
  useEffect(() => {
    if (textareaRef.current) {
      // Reset height to auto to get the correct scrollHeight
      textareaRef.current.style.height = 'auto';
      // Set height to scrollHeight (content height)
      const newHeight = Math.min(textareaRef.current.scrollHeight, 200); // Max 200px
      textareaRef.current.style.height = `${newHeight}px`;
    }
  }, [value]);

  const handleChange = (e: ChangeEvent<HTMLTextAreaElement>) => {
    const newValue = e.target.value;
    // Enforce character limit
    if (newValue.length <= MAX_CHARS) {
      setValue(newValue);
    }
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    // Send on Enter (without Shift)
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
    // Allow Shift+Enter for new line (default behavior)
  };

  const handleSend = () => {
    const trimmedValue = value.trim();
    if (trimmedValue && !disabled) {
      onSend(trimmedValue);
      setValue(''); // Clear input after send
    }
  };

  const charCount = value.length;
  const isOverLimit = charCount > MAX_CHARS;
  const showCharCount = charCount > MAX_CHARS * 0.8; // Show when 80% full

  return (
    <div className="p-2 bg-gray-50">
      <div className="flex items-end space-x-2">
        {/* Textarea */}
        <div className="flex-1 relative">
          <textarea
            ref={textareaRef}
            value={value}
            onChange={handleChange}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            disabled={disabled}
            rows={1}
            aria-label="Message input"
            className={`w-full px-3 py-2 text-sm border rounded-lg resize-none transition-all duration-200 ${
              disabled
                ? 'bg-gray-100 text-gray-500 cursor-not-allowed border-gray-200 opacity-60'
                : 'bg-white text-gray-900 border-gray-300 hover:border-gray-400 focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500'
            } ${isOverLimit ? 'border-red-500 focus:ring-red-500 focus:border-red-500' : ''}`}
            style={{
              minHeight: '36px',
              maxHeight: '120px',
            }}
          />
          
          {/* Character count */}
          {showCharCount && (
            <div
              className={`absolute bottom-1.5 right-1.5 text-xs font-medium px-1.5 py-0.5 rounded ${
                isOverLimit 
                  ? 'text-red-700 bg-red-50' 
                  : 'text-gray-600 bg-gray-100'
              }`}
            >
              {charCount}/{MAX_CHARS}
            </div>
          )}
        </div>

        {/* Send Button - Compact but touch-friendly */}
        <button
          onClick={handleSend}
          disabled={disabled || !value.trim() || isOverLimit}
          aria-label="Send message"
          className={`flex-shrink-0 flex items-center justify-center p-2 rounded-lg font-medium transition-all duration-200 ${
            disabled || !value.trim() || isOverLimit
              ? 'bg-gray-300 text-gray-500 cursor-not-allowed opacity-60'
              : 'bg-blue-600 text-white hover:bg-blue-700 active:bg-blue-800 active:scale-95 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-1'
          }`}
          style={{ minHeight: '36px', minWidth: '36px' }}
        >
          {disabled ? (
            <svg
              className="animate-spin h-4 w-4"
              fill="none"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <circle
                className="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                strokeWidth="4"
              ></circle>
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
          ) : (
            <svg
              className="h-4 w-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              aria-hidden="true"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
              />
            </svg>
          )}
        </button>
      </div>
    </div>
  );
}
