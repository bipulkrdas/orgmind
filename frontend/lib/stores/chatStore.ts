import { create } from 'zustand';
import { ChatMessage, ChatThread } from '../types';

/**
 * Chat Store Interface
 * Manages chat state including messages, active thread, loading, and errors
 */
export interface ChatStore {
  // State
  messages: Map<string, ChatMessage[]>; // threadId -> messages array
  activeThreadId: string | null;
  streamingMessage: string | null; // Current message being streamed
  isLoading: boolean;
  error: string | null;

  // Thread list state
  threads: ChatThread[];
  selectedThreadId: string | null;

  // Thread actions
  setActiveThread: (threadId: string) => void;

  // Thread list actions
  setThreads: (threads: ChatThread[]) => void;
  addThread: (thread: ChatThread) => void;
  selectThread: (threadId: string | null) => void;
  updateThreadTimestamp: (threadId: string) => void;

  // Message actions
  addMessage: (threadId: string, message: ChatMessage) => void;
  setMessages: (threadId: string, messages: ChatMessage[]) => void;
  updateStreamingMessage: (content: string) => void;
  finalizeStreamingMessage: () => void;

  // Loading and error actions
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  clearMessages: (threadId: string) => void;
}

/**
 * Zustand Chat Store
 * Provides efficient state management for chat messages with optimized re-renders
 */
export const useChatStore = create<ChatStore>((set, get) => ({
  // Initial state
  messages: new Map<string, ChatMessage[]>(),
  activeThreadId: null,
  streamingMessage: null,
  isLoading: false,
  error: null,

  // Thread list state
  threads: [],
  selectedThreadId: null,

  // Thread actions
  setActiveThread: (threadId: string) => {
    set({ 
      activeThreadId: threadId,
      streamingMessage: null,
      error: null 
    });
  },

  // Thread list actions
  setThreads: (threads: ChatThread[]) => {
    set({ threads });
  },

  addThread: (thread: ChatThread) => {
    set((state) => ({
      threads: [thread, ...state.threads] // Add to top
    }));
  },

  selectThread: (threadId: string | null) => {
    set({ 
      selectedThreadId: threadId,
      activeThreadId: threadId,
      streamingMessage: null,
      error: null 
    });
  },

  updateThreadTimestamp: (threadId: string) => {
    set((state) => ({
      threads: state.threads
        .map(t => 
          t.id === threadId 
            ? { ...t, updatedAt: new Date().toISOString() }
            : t
        )
        .sort((a, b) => 
          new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
        )
    }));
  },

  // Message actions
  addMessage: (threadId: string, message: ChatMessage) => {
    set((state) => {
      const newMessages = new Map(state.messages);
      const threadMessages = newMessages.get(threadId) || [];
      
      // Check if message with this ID already exists to prevent duplicates
      const existingIndex = threadMessages.findIndex(m => m.id === message.id);
      
      if (existingIndex >= 0) {
        // Update existing message instead of adding duplicate
        const updatedMessages = [...threadMessages];
        updatedMessages[existingIndex] = message;
        newMessages.set(threadId, updatedMessages);
      } else {
        // Add new message
        newMessages.set(threadId, [...threadMessages, message]);
      }
      
      return { messages: newMessages };
    });
  },

  setMessages: (threadId: string, messages: ChatMessage[]) => {
    set((state) => {
      const newMessages = new Map(state.messages);
      // Replace all messages for this thread (used when loading from API)
      newMessages.set(threadId, messages);
      return { messages: newMessages };
    });
  },

  updateStreamingMessage: (content: string) => {
    // CUMULATIVE: Append new content to existing streaming message
    set((state) => ({
      streamingMessage: (state.streamingMessage || '') + content
    }));
  },

  finalizeStreamingMessage: () => {
    set({ streamingMessage: null, isLoading: false });
  },

  // Loading and error actions
  setLoading: (loading: boolean) => {
    set({ isLoading: loading });
  },

  setError: (error: string | null) => {
    set({ error, isLoading: false });
  },

  clearMessages: (threadId: string) => {
    set((state) => {
      const newMessages = new Map(state.messages);
      newMessages.delete(threadId);
      return { messages: newMessages };
    });
  },
}));

/**
 * Selectors for optimized re-renders
 * Use these to subscribe to specific parts of the state
 */

// Cache for empty array to avoid creating new references
const EMPTY_MESSAGES: ChatMessage[] = [];

// Get messages for the active thread
export const selectActiveThreadMessages = (state: ChatStore): ChatMessage[] => {
  if (!state.activeThreadId) return EMPTY_MESSAGES;
  return state.messages.get(state.activeThreadId) || EMPTY_MESSAGES;
};

// Get loading state
export const selectIsLoading = (state: ChatStore): boolean => state.isLoading;

// Get streaming message
export const selectStreamingMessage = (state: ChatStore): string | null => 
  state.streamingMessage;

// Get error state
export const selectError = (state: ChatStore): string | null => state.error;

// Get active thread ID
export const selectActiveThreadId = (state: ChatStore): string | null => 
  state.activeThreadId;

// Thread list selectors
export const selectThreads = (state: ChatStore): ChatThread[] => state.threads;

export const selectSelectedThreadId = (state: ChatStore): string | null => 
  state.selectedThreadId;
