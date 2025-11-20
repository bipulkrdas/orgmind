package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bipulkrdas/orgmind/backend/internal/middleware"
	"github.com/bipulkrdas/orgmind/backend/internal/models"
	"github.com/bipulkrdas/orgmind/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// ChatHandler handles chat-related HTTP requests
type ChatHandler struct {
	chatService  service.ChatService
	graphService service.GraphService
}

// NewChatHandler creates a new instance of ChatHandler
func NewChatHandler(chatService service.ChatService, graphService service.GraphService) *ChatHandler {
	return &ChatHandler{
		chatService:  chatService,
		graphService: graphService,
	}
}

// ChatThreadResponse represents a chat thread in API responses
type ChatThreadResponse struct {
	ID        string  `json:"id"`
	GraphID   string  `json:"graphId"`
	UserID    string  `json:"userId"`
	Summary   *string `json:"summary,omitempty"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

// ChatMessageResponse represents a chat message in API responses
type ChatMessageResponse struct {
	ID        string `json:"id"`
	ThreadID  string `json:"threadId"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

// SendMessageRequest represents the request body for sending a message
type SendMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

// SendMessageResponse represents the response for sending a message
type SendMessageResponse struct {
	MessageID string `json:"messageId"`
	Status    string `json:"status"`
}

// MessagesResponse represents the paginated messages response
type MessagesResponse struct {
	Messages []ChatMessageResponse `json:"messages"`
	Total    int                   `json:"total"`
	HasMore  bool                  `json:"hasMore"`
}

// ListThreads handles GET /api/graphs/:id/chat/threads
func (h *ChatHandler) ListThreads(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Get graph ID from URL parameter
	graphID := c.Param("id")
	if graphID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Graph ID is required"})
		return
	}

	// List threads for the graph
	threads, err := h.chatService.ListThreads(c.Request.Context(), graphID, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotGraphMember) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to this graph"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list threads", "details": err.Error()})
		return
	}

	// Convert to response format
	response := make([]ChatThreadResponse, len(threads))
	for i, thread := range threads {
		response[i] = ChatThreadResponse{
			ID:        thread.ID,
			GraphID:   thread.GraphID,
			UserID:    thread.UserID,
			Summary:   thread.Summary,
			CreatedAt: thread.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt: thread.UpdatedAt.UTC().Format(time.RFC3339),
		}
	}

	// Return threads array directly (not wrapped)
	c.JSON(http.StatusOK, response)
}

// CreateThread handles POST /api/graphs/:id/chat/threads
func (h *ChatHandler) CreateThread(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Get graph ID from URL parameter
	graphID := c.Param("id")
	if graphID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Graph ID is required"})
		return
	}

	// Verify graph membership (this is done in the service, but we can also verify here)
	_, err := h.graphService.GetByID(c.Request.Context(), graphID, userID)
	if err != nil {
		if errors.Is(err, service.ErrGraphNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Graph not found"})
			return
		}
		if errors.Is(err, service.ErrNotGraphMember) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to this graph"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify graph access", "details": err.Error()})
		return
	}

	// Create thread
	thread, err := h.chatService.CreateThread(c.Request.Context(), graphID, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotGraphMember) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to this graph"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chat thread", "details": err.Error()})
		return
	}

	// Return thread response
	c.JSON(http.StatusCreated, ChatThreadResponse{
		ID:        thread.ID,
		GraphID:   thread.GraphID,
		UserID:    thread.UserID,
		Summary:   thread.Summary,
		CreatedAt: thread.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: thread.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// GetThreadMessages handles GET /api/graphs/:id/chat/threads/:threadId/messages
func (h *ChatHandler) GetThreadMessages(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Get graph ID from URL parameter
	graphID := c.Param("id")
	if graphID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Graph ID is required"})
		return
	}

	// Get thread ID from URL parameter
	threadID := c.Param("threadId")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thread ID is required"})
		return
	}

	// Parse pagination parameters
	limit := 50 // default
	offset := 0 // default

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Verify user access to thread
	thread, err := h.chatService.GetThread(c.Request.Context(), threadID, userID)
	if err != nil {
		if errors.Is(err, service.ErrChatThreadNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Chat thread not found"})
			return
		}
		if errors.Is(err, service.ErrChatUnauthorized) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to this chat thread"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify thread access", "details": err.Error()})
		return
	}

	// Verify thread belongs to the graph
	if thread.GraphID != graphID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thread does not belong to this graph"})
		return
	}

	// Get messages with pagination
	messages, err := h.chatService.GetMessages(c.Request.Context(), threadID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages", "details": err.Error()})
		return
	}

	// Convert to response format
	response := make([]ChatMessageResponse, len(messages))
	for i, msg := range messages {
		response[i] = ChatMessageResponse{
			ID:        msg.ID,
			ThreadID:  msg.ThreadID,
			Role:      msg.Role,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt.UTC().Format(time.RFC3339),
		}
	}

	// Determine if there are more messages
	hasMore := len(messages) == limit

	// Return messages with pagination metadata
	c.JSON(http.StatusOK, MessagesResponse{
		Messages: response,
		Total:    len(messages),
		HasMore:  hasMore,
	})
}

// SendMessage handles POST /api/graphs/:id/chat/threads/:threadId/messages
// This endpoint ONLY saves the user message and returns it immediately
// The AI response generation happens via the SSE stream endpoint
func (h *ChatHandler) SendMessage(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Get graph ID from URL parameter
	graphID := c.Param("id")
	if graphID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Graph ID is required"})
		return
	}

	// Get thread ID from URL parameter
	threadID := c.Param("threadId")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thread ID is required"})
		return
	}

	// Parse request body
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate content length
	if len(req.Content) > 4000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message content exceeds 4000 characters"})
		return
	}

	// Verify user access to thread
	thread, err := h.chatService.GetThread(c.Request.Context(), threadID, userID)
	if err != nil {
		if errors.Is(err, service.ErrChatThreadNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Chat thread not found"})
			return
		}
		if errors.Is(err, service.ErrChatUnauthorized) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to this chat thread"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify thread access", "details": err.Error()})
		return
	}

	// Verify thread belongs to the graph
	if thread.GraphID != graphID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thread does not belong to this graph"})
		return
	}

	// Save the user message (only the user message, not the AI response)
	userMessage, err := h.chatService.SaveUserMessage(c.Request.Context(), threadID, userID, req.Content)
	if err != nil {
		if errors.Is(err, service.ErrRateLimitExceeded) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded: maximum 20 messages per minute"})
			return
		}
		if errors.Is(err, service.ErrMessageTooLong) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Message content exceeds 4000 characters"})
			return
		}
		if errors.Is(err, service.ErrInvalidMessageContent) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Message content is required"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message", "details": err.Error()})
		return
	}

	// Return the saved user message
	// The client will then open an SSE stream to get the AI response
	c.JSON(http.StatusCreated, convertMessageToResponse(userMessage))
}

// StreamResponse handles GET /api/graphs/:id/chat/stream
// This endpoint generates the AI response and streams it back via SSE
// It expects the user message to already be saved (via SendMessage endpoint)
func (h *ChatHandler) StreamResponse(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Get graph ID from URL parameter
	graphID := c.Param("id")
	if graphID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Graph ID is required"})
		return
	}

	// Get threadId and userMessageId from query params
	threadID := c.Query("threadId")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "threadId query parameter is required"})
		return
	}

	userMessageID := c.Query("userMessageId")
	if userMessageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userMessageId query parameter is required"})
		return
	}

	// Verify user access to thread
	thread, err := h.chatService.GetThread(c.Request.Context(), threadID, userID)
	if err != nil {
		if errors.Is(err, service.ErrChatThreadNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Chat thread not found"})
			return
		}
		if errors.Is(err, service.ErrChatUnauthorized) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to this chat thread"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify thread access", "details": err.Error()})
		return
	}

	// Verify thread belongs to the graph
	if thread.GraphID != graphID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thread does not belong to this graph"})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	// Create response channel
	responseChan := make(chan string, 100)
	errorChan := make(chan error, 1)
	assistantMessageIDChan := make(chan string, 1)

	// Start AI response generation in goroutine
	go func() {
		defer close(errorChan)
		defer close(assistantMessageIDChan)
		defer close(responseChan) // Close response channel after everything is done

		// Generate AI response based on the user message ID
		assistantMessageID, err := h.chatService.GenerateResponseForMessage(
			c.Request.Context(),
			threadID,
			userMessageID,
			graphID,
			responseChan,
		)

		if err != nil {
			errorChan <- err
			return
		}

		// Send the assistant message ID when done
		assistantMessageIDChan <- assistantMessageID
	}()

	// Stream chunks to client
	c.Writer.Flush()

	for {
		select {
		case chunk, ok := <-responseChan:
			if !ok {
				// Channel closed, wait for either error or success
				// Use a blocking select to wait for the goroutine to finish
				select {
				case err, ok := <-errorChan:
					// Check if channel was closed without sending an error
					if !ok {
						// Channel closed without error - this shouldn't happen, but treat as success
						// The assistantMessageIDChan should have the result
						select {
						case assistantMessageID := <-assistantMessageIDChan:
							c.SSEvent("done", map[string]interface{}{
								"content": assistantMessageID,
							})
							c.Writer.Flush()
							return
						default:
							// Neither channel has data - unexpected state
							c.SSEvent("error", map[string]interface{}{"error": "Unexpected streaming completion state"})
							c.Writer.Flush()
							return
						}
					}

					// Channel had an actual error
					if errors.Is(err, service.ErrRateLimitExceeded) {
						c.SSEvent("error", map[string]interface{}{"error": "Rate limit exceeded"})
					} else {
						c.SSEvent("error", map[string]interface{}{"error": "Failed to generate response"})
					}
					c.Writer.Flush()
					return

				case assistantMessageID, ok := <-assistantMessageIDChan:
					// Check if channel was closed without sending a message ID
					if !ok {
						// Channel closed without message ID - check for error
						select {
						case err, ok := <-errorChan:
							if ok && err != nil {
								// There was an error
								if errors.Is(err, service.ErrRateLimitExceeded) {
									c.SSEvent("error", map[string]interface{}{"error": "Rate limit exceeded"})
								} else {
									c.SSEvent("error", map[string]interface{}{"error": "Failed to generate response"})
								}
							} else {
								// No error either - unexpected state
								c.SSEvent("error", map[string]interface{}{"error": "Unexpected streaming completion state"})
							}
							c.Writer.Flush()
							return
						default:
							// No error - unexpected state
							c.SSEvent("error", map[string]interface{}{"error": "Unexpected streaming completion state"})
							c.Writer.Flush()
							return
						}
					}

					// Got a valid message ID - success!
					c.SSEvent("done", map[string]interface{}{
						"content": assistantMessageID,
					})
					c.Writer.Flush()
					return

				case <-c.Request.Context().Done():
					// Client disconnected while waiting for completion
					return
				}
			}

			// Send chunk event
			c.SSEvent("chunk", map[string]interface{}{
				"content": chunk,
			})
			c.Writer.Flush()

		case <-c.Request.Context().Done():
			// Client disconnected
			return
		}
	}
}

// Helper functions for error handling and validation

// validateGraphID validates that a graph ID is provided
func validateGraphID(graphID string) error {
	if graphID == "" {
		return fmt.Errorf("graph ID is required")
	}
	return nil
}

// validateThreadID validates that a thread ID is provided
func validateThreadID(threadID string) error {
	if threadID == "" {
		return fmt.Errorf("thread ID is required")
	}
	return nil
}

// validateMessageContent validates message content
func validateMessageContent(content string) error {
	if content == "" {
		return fmt.Errorf("message content is required")
	}
	if len(content) > 4000 {
		return fmt.Errorf("message content exceeds 4000 characters")
	}
	return nil
}

// handleServiceError maps service errors to appropriate HTTP responses
func handleServiceError(c *gin.Context, err error, operation string) {
	switch {
	case errors.Is(err, service.ErrGraphNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Graph not found"})
	case errors.Is(err, service.ErrNotGraphMember):
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to this graph"})
	case errors.Is(err, service.ErrChatThreadNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat thread not found"})
	case errors.Is(err, service.ErrChatUnauthorized):
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to this chat thread"})
	case errors.Is(err, service.ErrRateLimitExceeded):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded: maximum 20 messages per minute"})
	case errors.Is(err, service.ErrMessageTooLong):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message content exceeds 4000 characters"})
	case errors.Is(err, service.ErrInvalidMessageContent):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message content is required"})
	default:
		// Log the error with context
		fmt.Printf("Error in %s: %v\n", operation, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   fmt.Sprintf("Failed to %s", operation),
			"details": err.Error(),
		})
	}
}

// convertThreadToResponse converts a ChatThread model to response format
func convertThreadToResponse(thread *models.ChatThread) ChatThreadResponse {
	return ChatThreadResponse{
		ID:        thread.ID,
		GraphID:   thread.GraphID,
		UserID:    thread.UserID,
		Summary:   thread.Summary,
		CreatedAt: thread.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: thread.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// convertMessageToResponse converts a ChatMessage model to response format
func convertMessageToResponse(message *models.ChatMessage) ChatMessageResponse {
	return ChatMessageResponse{
		ID:        message.ID,
		ThreadID:  message.ThreadID,
		Role:      message.Role,
		Content:   message.Content,
		CreatedAt: message.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// convertMessagesToResponse converts a slice of ChatMessage models to response format
func convertMessagesToResponse(messages []*models.ChatMessage) []ChatMessageResponse {
	response := make([]ChatMessageResponse, len(messages))
	for i, msg := range messages {
		response[i] = convertMessageToResponse(msg)
	}
	return response
}
