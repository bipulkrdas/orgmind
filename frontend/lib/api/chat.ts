import { apiCall, APIError, API_BASE_URL } from './client';
import { getJWTToken } from '../auth/jwt';
import type { ChatThread, ChatMessage, StreamEvent } from '../types';

/**
 * List all threads for a graph
 * Implements defensive response parsing to handle multiple response formats
 */
export async function listThreads(graphId: string): Promise<ChatThread[]> {
  const response = await apiCall<ChatThread[] | { threads: ChatThread[] }>(
    `/api/graphs/${graphId}/chat/threads`,
    { method: 'GET' }
  );

  // Defensive: Handle direct array response
  if (Array.isArray(response)) {
    return response;
  }

  // Defensive: Handle wrapped response format
  if (response && typeof response === 'object' && 'threads' in response) {
    return Array.isArray(response.threads) ? response.threads : [];
  }

  // Fallback to empty array
  return [];
}

/**
 * Create a new chat thread for a graph
 */
export async function createThread(graphId: string): Promise<ChatThread> {
  return apiCall<ChatThread>(`/api/graphs/${graphId}/chat/threads`, {
    method: 'POST',
  });
}

/**
 * Get messages for a specific thread with pagination
 */
export async function getThreadMessages(
  graphId: string,
  threadId: string,
  limit: number = 50,
  offset: number = 0
): Promise<{ messages: ChatMessage[]; total: number; hasMore: boolean }> {
  const response = await apiCall<{
    messages: ChatMessage[];
    total: number;
    hasMore: boolean;
  }>(`/api/graphs/${graphId}/chat/threads/${threadId}/messages?limit=${limit}&offset=${offset}`, {
    method: 'GET',
  });

  // Handle defensive response format
  if (response && typeof response === 'object') {
    return {
      messages: Array.isArray(response.messages) ? response.messages : [],
      total: response.total || 0,
      hasMore: response.hasMore || false,
    };
  }

  // Fallback to empty response
  return {
    messages: [],
    total: 0,
    hasMore: false,
  };
}

/**
 * Send a message to a chat thread
 * Returns the saved user message with its ID
 */
export async function sendMessage(
  graphId: string,
  threadId: string,
  content: string
): Promise<ChatMessage> {
  return apiCall<ChatMessage>(
    `/api/graphs/${graphId}/chat/threads/${threadId}/messages`,
    {
      method: 'POST',
      body: JSON.stringify({ content }),
    }
  );
}

/**
 * Connect to SSE stream for AI response
 * Returns a cleanup function to close the connection
 * 
 * ROBUST ERROR HANDLING:
 * - If chunks are received followed by an error, the chunks are preserved
 * - If only an error is received, it's displayed
 * - The streaming message is finalized only after all events are processed
 * 
 * @param graphId - The graph ID
 * @param threadId - The chat thread ID
 * @param userMessageId - The ID of the user message that triggered this response
 * @param onChunk - Callback for each content chunk
 * @param onDone - Callback when streaming is complete (receives assistant message ID)
 * @param onError - Callback for errors
 */
export function connectChatStream(
  graphId: string,
  threadId: string,
  userMessageId: string,
  onChunk: (content: string) => void,
  onDone: (messageId: string) => void,
  onError: (error: string) => void
): () => void {
  const token = getJWTToken();
  
  if (!token) {
    onError('Authentication required');
    return () => {};
  }

  // Construct SSE URL with query parameters
  // The backend will use the userMessageId to generate a response
  const url = `${API_BASE_URL}/api/graphs/${graphId}/chat/stream?threadId=${threadId}&userMessageId=${userMessageId}`;
  
  let abortController = new AbortController();
  let reconnectAttempts = 0;
  const maxReconnectAttempts = 3;
  const baseDelay = 1000; // 1 second
  
  // Track if we received any chunks before an error
  let receivedChunks = false;
  let receivedDone = false;
  
  const connect = async () => {
    try {
      const response = await fetch(url, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Accept': 'text/event-stream',
        },
        signal: abortController.signal,
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      if (!response.body) {
        throw new Error('Response body is null');
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';

      while (true) {
        const { done, value } = await reader.read();
        
        if (done) {
          break;
        }

        // Decode the chunk and add to buffer
        buffer += decoder.decode(value, { stream: true });
        
        // Process complete SSE messages (separated by \n\n)
        const messages = buffer.split('\n\n');
        buffer = messages.pop() || ''; // Keep incomplete message in buffer

        for (const message of messages) {
          if (!message.trim()) continue;

          // Parse SSE format: "event: type\ndata: content"
          const lines = message.split('\n');
          let eventType = 'message';
          let eventData = '';

          for (const line of lines) {
            if (line.startsWith('event:')) {
              eventType = line.substring(6).trim();
            } else if (line.startsWith('data:')) {
              eventData = line.substring(5).trim();
            }
          }

          // Parse the event data
          try {
            const data = JSON.parse(eventData);

            switch (eventType) {
              case 'chunk':
                if (data.content) {
                  receivedChunks = true;
                  onChunk(data.content);
                }
                break;
              case 'done':
                if (data.content) {
                  receivedDone = true;
                  onDone(data.content); // messageId of assistant message
                }
                return; // Close connection on done
              case 'error':
                // ROBUST: Only show error if we haven't received a successful completion
                // If we got chunks but then an error, the chunks are already displayed
                if (!receivedDone) {
                  console.warn('SSE error event received:', data.error, 
                    receivedChunks ? '(chunks were received before error)' : '(no chunks received)');
                  
                  // Only call onError if we haven't received any content
                  // This prevents overwriting a successful response with an error
                  if (!receivedChunks) {
                    onError(data.error || 'Unknown error occurred');
                  }
                }
                return; // Close connection on error
            }
          } catch (parseError) {
            console.error('Failed to parse SSE event:', parseError, 'Raw data:', eventData);
          }
        }
      }
    } catch (error) {
      // Don't retry if aborted
      if (abortController.signal.aborted) {
        return;
      }

      // Implement exponential backoff for reconnection
      if (reconnectAttempts < maxReconnectAttempts) {
        reconnectAttempts++;
        const delay = baseDelay * Math.pow(2, reconnectAttempts - 1);
        
        console.log(`SSE connection failed, retrying in ${delay}ms (attempt ${reconnectAttempts}/${maxReconnectAttempts})`);
        
        setTimeout(() => {
          if (!abortController.signal.aborted) {
            connect();
          }
        }, delay);
      } else {
        onError(
          error instanceof Error 
            ? error.message 
            : 'Failed to connect to chat stream after multiple attempts'
        );
      }
    }
  };

  // Start the connection
  connect();

  // Return cleanup function
  return () => {
    abortController.abort();
  };
}
