import { create } from 'zustand';
import { ChatMessage } from '../types';

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

  // Thread actions
  setActiveThread: (threadId: string) => void;

  // Message actions
  addMessage: (threadId: string, message: ChatMessage) => void;
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

  // Thread actions
  setActiveThread: (threadId: string) => {
    set({ 
      activeThreadId: threadId,
      streamingMessage: null,
      error: null 
    });
  },

  // Message actions
  addMessage: (threadId: string, message: ChatMessage) => {
    set((state) => {
      const newMessages = new Map(state.messages);
      const threadMessages = newMessages.get(threadId) || [];
      newMessages.set(threadId, [...threadMessages, message]);
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
