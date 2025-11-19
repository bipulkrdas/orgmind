'use client';

export function TypingIndicator() {
  return (
    <div 
      className="flex items-center space-x-2 px-4 py-3 bg-gray-100 rounded-2xl rounded-bl-sm shadow-sm max-w-[80px]"
      role="status"
      aria-label="AI is typing"
    >
      <div className="flex space-x-1">
        <div 
          className="w-2 h-2 bg-gray-500 rounded-full animate-bounce" 
          style={{ animationDelay: '0ms', animationDuration: '1s' }}
        ></div>
        <div 
          className="w-2 h-2 bg-gray-500 rounded-full animate-bounce" 
          style={{ animationDelay: '150ms', animationDuration: '1s' }}
        ></div>
        <div 
          className="w-2 h-2 bg-gray-500 rounded-full animate-bounce" 
          style={{ animationDelay: '300ms', animationDuration: '1s' }}
        ></div>
      </div>
    </div>
  );
}
