'use client';

import { NewThreadPrompt } from './NewThreadPrompt';
import { ThreadConversation } from './ThreadConversation';

interface ChatPanelProps {
  graphId: string;
  selectedThreadId: string | null;
  onThreadCreated: (threadId: string) => void;
  onMessageSent: () => void;
}

/**
 * ChatPanel Component
 * 
 * Right column of the chat interface that conditionally renders:
 * - NewThreadPrompt when no thread is selected (selectedThreadId is null)
 * - ThreadConversation when a thread is selected
 * 
 * This component acts as a router between the two states of the chat interface.
 * Includes smooth transitions when switching between states.
 * 
 * Requirements: 2.1, 2.4, 3.1, 3.4
 */
export function ChatPanel({
  graphId,
  selectedThreadId,
  onThreadCreated,
  onMessageSent,
}: ChatPanelProps) {
  // Show NewThreadPrompt when no thread is selected
  if (!selectedThreadId) {
    return (
      <div className="h-full animate-fadeIn">
        <NewThreadPrompt
          graphId={graphId}
          onThreadCreated={onThreadCreated}
        />
      </div>
    );
  }

  // Show ThreadConversation when thread is selected
  return (
    <div className="h-full animate-fadeIn">
      <ThreadConversation
        graphId={graphId}
        threadId={selectedThreadId}
        onMessageSent={onMessageSent}
      />
    </div>
  );
}
