'use client';

import { ChatMessage as ChatMessageType } from '@/lib/types';

interface ChatMessageProps {
  message: ChatMessageType;
}

export function ChatMessage({ message }: ChatMessageProps) {
  const isUser = message.role === 'user';
  const timestamp = new Date(message.createdAt).toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
  });

  return (
    <div 
      className={`flex ${isUser ? 'justify-end' : 'justify-start'} mb-6 animate-fadeIn`}
      role="article"
      aria-label={`${isUser ? 'Your' : 'Assistant'} message`}
    >
      <div className={`flex flex-col ${isUser ? 'max-w-[75%] sm:max-w-[70%] items-end' : 'w-full items-start'}`}>
        {/* Message Bubble */}
        <div
          className={`px-4 py-3 rounded-2xl transition-all duration-200 ${
            isUser
              ? 'bg-blue-600 text-white rounded-br-sm shadow-md hover:shadow-lg'
              : 'bg-gray-100 text-gray-900 rounded-bl-sm shadow-sm hover:shadow-md w-full'
          }`}
        >
          {isUser ? (
            // User messages: plain text
            <p className="text-sm sm:text-base whitespace-pre-wrap break-words leading-relaxed">
              {message.content}
            </p>
          ) : (
            // Assistant messages: render with basic markdown support
            <div className="text-sm sm:text-base whitespace-pre-wrap break-words prose prose-sm max-w-none leading-relaxed">
              {renderMarkdown(message.content)}
            </div>
          )}
        </div>

        {/* Timestamp */}
        <span className="text-xs text-gray-500 mt-1.5 px-2 opacity-75">
          {timestamp}
        </span>
      </div>
    </div>
  );
}

/**
 * Simple markdown renderer for assistant messages
 * Supports basic formatting: bold, italic, code blocks, inline code, links
 */
function renderMarkdown(content: string): JSX.Element {
  // Split by code blocks first
  const codeBlockRegex = /```(\w+)?\n([\s\S]*?)```/g;
  const parts: JSX.Element[] = [];
  let lastIndex = 0;
  let match;
  let key = 0;

  while ((match = codeBlockRegex.exec(content)) !== null) {
    // Add text before code block
    if (match.index > lastIndex) {
      const textBefore = content.substring(lastIndex, match.index);
      parts.push(
        <span key={`text-${key++}`} dangerouslySetInnerHTML={{ __html: formatInlineMarkdown(textBefore) }} />
      );
    }

    // Add code block
    const language = match[1] || '';
    const code = match[2];
    parts.push(
      <pre key={`code-${key++}`} className="bg-gray-800 text-gray-100 p-3 rounded my-2 overflow-x-auto">
        <code className={language ? `language-${language}` : ''}>{code}</code>
      </pre>
    );

    lastIndex = match.index + match[0].length;
  }

  // Add remaining text
  if (lastIndex < content.length) {
    const textAfter = content.substring(lastIndex);
    parts.push(
      <span key={`text-${key++}`} dangerouslySetInnerHTML={{ __html: formatInlineMarkdown(textAfter) }} />
    );
  }

  return <>{parts}</>;
}

/**
 * Format inline markdown (bold, italic, inline code, links)
 */
function formatInlineMarkdown(text: string): string {
  return text
    // Escape HTML
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    // Bold: **text** or __text__
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/__(.+?)__/g, '<strong>$1</strong>')
    // Italic: *text* or _text_
    .replace(/\*(.+?)\*/g, '<em>$1</em>')
    .replace(/_(.+?)_/g, '<em>$1</em>')
    // Inline code: `code`
    .replace(/`(.+?)`/g, '<code class="bg-gray-200 text-gray-800 px-1 py-0.5 rounded text-xs">$1</code>')
    // Links: [text](url)
    .replace(/\[(.+?)\]\((.+?)\)/g, '<a href="$2" target="_blank" rel="noopener noreferrer" class="text-blue-600 hover:underline">$1</a>')
    // Line breaks
    .replace(/\n/g, '<br />');
}
