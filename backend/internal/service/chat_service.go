package service

import (
	"context"
	"fmt"
	"html"
	"strings"
	"sync"
	"time"

	"github.com/bipulkrdas/orgmind/backend/internal/models"
	"github.com/bipulkrdas/orgmind/backend/internal/repository"
	"github.com/google/uuid"
)

// Custom errors for chat operations
var (
	ErrChatThreadNotFound    = fmt.Errorf("chat thread not found")
	ErrChatUnauthorized      = fmt.Errorf("you don't have access to this chat thread")
	ErrMessageTooLong        = fmt.Errorf("message content exceeds 4000 characters")
	ErrRateLimitExceeded     = fmt.Errorf("rate limit exceeded: maximum 20 messages per minute")
	ErrInvalidMessageContent = fmt.Errorf("message content is required")
)

// chatService implements the ChatService interface
type chatService struct {
	chatRepo    repository.ChatRepository
	graphRepo   repository.GraphRepository
	geminiSvc   GeminiService
	rateLimiter *rateLimiter
}

// NewChatService creates a new chat service instance
func NewChatService(
	chatRepo repository.ChatRepository,
	graphRepo repository.GraphRepository,
	geminiSvc GeminiService,
) ChatService {
	return &chatService{
		chatRepo:    chatRepo,
		graphRepo:   graphRepo,
		geminiSvc:   geminiSvc,
		rateLimiter: newRateLimiter(20, time.Minute), // 20 messages per minute
	}
}

// CreateThread creates a new chat thread for a graph
func (s *chatService) CreateThread(ctx context.Context, graphID, userID string) (*models.ChatThread, error) {
	// Verify user is a member of the graph
	isMember, err := s.graphRepo.IsMember(ctx, graphID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify graph membership: %w", err)
	}
	if !isMember {
		return nil, ErrNotGraphMember
	}

	// Create thread
	now := time.Now()
	thread := &models.ChatThread{
		ID:        uuid.New().String(),
		GraphID:   graphID,
		UserID:    userID,
		Summary:   nil,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Validate thread
	if err := thread.Validate(); err != nil {
		return nil, fmt.Errorf("invalid thread: %w", err)
	}

	// Save to database
	if err := s.chatRepo.CreateThread(ctx, thread); err != nil {
		return nil, fmt.Errorf("failed to create thread: %w", err)
	}

	return thread, nil
}

// GetThread retrieves a chat thread with access control
func (s *chatService) GetThread(ctx context.Context, threadID, userID string) (*models.ChatThread, error) {
	// Get thread from database
	thread, err := s.chatRepo.GetThreadByID(ctx, threadID)
	if err != nil {
		return nil, ErrChatThreadNotFound
	}

	// Verify user has access to the graph
	isMember, err := s.graphRepo.IsMember(ctx, thread.GraphID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify graph membership: %w", err)
	}
	if !isMember {
		return nil, ErrChatUnauthorized
	}

	return thread, nil
}

// ListThreads lists all threads for a graph with filtering
func (s *chatService) ListThreads(ctx context.Context, graphID, userID string) ([]*models.ChatThread, error) {
	// Verify user is a member of the graph
	isMember, err := s.graphRepo.IsMember(ctx, graphID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify graph membership: %w", err)
	}
	if !isMember {
		return nil, ErrNotGraphMember
	}

	// Get all threads for the graph
	threads, err := s.chatRepo.ListThreadsByGraphID(ctx, graphID)
	if err != nil {
		return nil, fmt.Errorf("failed to list threads: %w", err)
	}

	return threads, nil
}

// GetMessages retrieves messages for a thread with pagination
func (s *chatService) GetMessages(ctx context.Context, threadID string, limit, offset int) ([]*models.ChatMessage, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 50
	}

	// Get messages from database
	messages, err := s.chatRepo.GetMessagesByThreadID(ctx, threadID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return messages, nil
}

// SaveMessage saves a message with validation and sanitization
func (s *chatService) SaveMessage(ctx context.Context, message *models.ChatMessage) error {
	// Validate message
	if err := message.Validate(); err != nil {
		return err
	}

	// Sanitize content (escape HTML)
	message.Content = sanitizeContent(message.Content)

	// Save to database
	if err := s.chatRepo.CreateMessage(ctx, message); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	return nil
}

// GenerateResponse generates an AI response for a user message
func (s *chatService) GenerateResponse(ctx context.Context, threadID, userID, userMessage string, responseChan chan<- string) error {
	// Check rate limit
	if !s.rateLimiter.Allow(userID) {
		return ErrRateLimitExceeded
	}

	// Validate message content
	if strings.TrimSpace(userMessage) == "" {
		return ErrInvalidMessageContent
	}
	if len(userMessage) > 4000 {
		return ErrMessageTooLong
	}

	// Get thread and verify access
	thread, err := s.GetThread(ctx, threadID, userID)
	if err != nil {
		return err
	}

	// Verify user is a member of the graph
	isMember, err := s.graphRepo.IsMember(ctx, thread.GraphID, userID)
	if err != nil {
		return fmt.Errorf("failed to verify graph membership: %w", err)
	}
	if !isMember {
		return ErrChatUnauthorized
	}

	// Get graph for metadata filtering
	graph, err := s.graphRepo.GetByID(ctx, thread.GraphID)
	if err != nil {
		return fmt.Errorf("failed to get graph: %w", err)
	}

	// Save user message
	userMsg := &models.ChatMessage{
		ID:        uuid.New().String(),
		ThreadID:  threadID,
		Role:      "user",
		Content:   userMessage,
		CreatedAt: time.Now(),
	}
	if err := s.SaveMessage(ctx, userMsg); err != nil {
		return fmt.Errorf("failed to save user message: %w", err)
	}

	// Update thread summary if this is the first message
	if thread.Summary == nil {
		thread.GenerateSummary(userMessage)
		thread.UpdatedAt = time.Now()
		if err := s.chatRepo.UpdateThread(ctx, thread); err != nil {
			// Log error but continue
			fmt.Printf("Warning: failed to update thread summary: %v\n", err)
		}
	}

	// Generate AI response using Gemini service
	assistantMsg := &models.ChatMessage{
		ID:        uuid.New().String(),
		ThreadID:  threadID,
		Role:      "assistant",
		Content:   "",
		CreatedAt: time.Now(),
	}

	// Create a channel to collect the full response
	fullResponseChan := make(chan string, 100)
	var fullResponse strings.Builder

	// Start goroutine to collect response and forward to client
	go func() {
		for chunk := range fullResponseChan {
			fullResponse.WriteString(chunk)
			// Forward chunk to client
			select {
			case responseChan <- chunk:
			case <-ctx.Done():
				return
			}
		}

		// Save assistant message after streaming completes
		assistantMsg.Content = fullResponse.String()
		if err := s.SaveMessage(context.Background(), assistantMsg); err != nil {
			fmt.Printf("Error: failed to save assistant message: %v\n", err)
		}

		// Close the response channel to signal completion
		close(responseChan)
	}()

	// Call Gemini service for streaming response with metadata filtering
	// Use empty storeID to let service use the shared store
	if err := s.geminiSvc.GenerateStreamingResponse(ctx, "", graph.ID, "topeic.com", "1.1", userMessage, fullResponseChan); err != nil {
		close(fullResponseChan)
		return fmt.Errorf("failed to generate AI response: %w", err)
	}

	return nil
}

// SaveUserMessage saves a user message with validation and rate limiting
func (s *chatService) SaveUserMessage(ctx context.Context, threadID, userID, content string) (*models.ChatMessage, error) {
	// Check rate limit
	if !s.rateLimiter.Allow(userID) {
		return nil, ErrRateLimitExceeded
	}

	// Validate message content
	if strings.TrimSpace(content) == "" {
		return nil, ErrInvalidMessageContent
	}
	if len(content) > 4000 {
		return nil, ErrMessageTooLong
	}

	// Get thread and verify access
	thread, err := s.GetThread(ctx, threadID, userID)
	if err != nil {
		return nil, err
	}

	// Create user message
	userMsg := &models.ChatMessage{
		ID:        uuid.New().String(),
		ThreadID:  threadID,
		Role:      "user",
		Content:   content,
		CreatedAt: time.Now(),
	}

	// Save message
	if err := s.SaveMessage(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// Update thread summary if this is the first message
	if thread.Summary == nil {
		thread.GenerateSummary(content)
		thread.UpdatedAt = time.Now()
		if err := s.chatRepo.UpdateThread(ctx, thread); err != nil {
			// Log error but continue
			fmt.Printf("Warning: failed to update thread summary: %v\n", err)
		}
	}

	return userMsg, nil
}

// GenerateResponseForMessage generates an AI response for a specific user message
func (s *chatService) GenerateResponseForMessage(
	ctx context.Context,
	threadID string,
	userMessageID string,
	graphID string,
	responseChan chan<- string,
) (string, error) {
	// Get the user message
	userMsg, err := s.chatRepo.GetMessageByID(ctx, userMessageID)
	if err != nil {
		return "", fmt.Errorf("failed to get user message: %w", err)
	}

	// Verify message belongs to the thread
	if userMsg.ThreadID != threadID {
		return "", fmt.Errorf("message does not belong to thread")
	}

	// Verify message is from user
	if userMsg.Role != "user" {
		return "", fmt.Errorf("message is not from user")
	}

	// Get graph for metadata filtering
	graph, err := s.graphRepo.GetByID(ctx, graphID)
	if err != nil {
		return "", fmt.Errorf("failed to get graph: %w", err)
	}

	// Create assistant message
	assistantMsg := &models.ChatMessage{
		ID:        uuid.New().String(),
		ThreadID:  threadID,
		Role:      "assistant",
		Content:   "",
		CreatedAt: time.Now(),
	}

	// Create a channel to collect the full response
	fullResponseChan := make(chan string, 100)
	var fullResponse strings.Builder
	var streamErr error

	// Start goroutine to collect response and forward to client
	done := make(chan struct{})
	go func() {
		defer close(done)
		for chunk := range fullResponseChan {
			fullResponse.WriteString(chunk)
			// Forward chunk to client
			select {
			case responseChan <- chunk:
			case <-ctx.Done():
				streamErr = ctx.Err()
				return
			}
		}
	}()

	// Call Gemini service for streaming response with metadata filtering
	// Use empty storeID to let service use the shared store
	geminiErr := s.geminiSvc.GenerateStreamingResponse(ctx, "", graph.ID, "topeic.com", "1.1", userMsg.Content, fullResponseChan)

	// Close the channel to signal completion to the goroutine
	close(fullResponseChan)

	// Wait for goroutine to finish forwarding all chunks
	<-done

	// Check for Gemini errors first
	if geminiErr != nil {
		return "", fmt.Errorf("failed to generate AI response: %w", geminiErr)
	}

	// Check for streaming errors
	if streamErr != nil {
		return "", streamErr
	}

	// Save assistant message after streaming completes
	assistantMsg.Content = fullResponse.String()
	if err := s.SaveMessage(context.Background(), assistantMsg); err != nil {
		// Log error but DON'T fail - streaming was successful
		// The user already received the response, failing now would send both chunks AND error
		fmt.Printf("Error: failed to save assistant message: %v\n", err)
		// Return the message ID anyway so the client knows streaming completed
		// The message just won't be persisted in the database
	}

	return assistantMsg.ID, nil
}

// sanitizeContent sanitizes message content by escaping HTML
func sanitizeContent(content string) string {
	// Escape HTML to prevent XSS
	return html.EscapeString(content)
}

// rateLimiter implements a simple in-memory rate limiter
type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// newRateLimiter creates a new rate limiter
func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Start cleanup goroutine to prevent memory leaks
	go rl.cleanup()

	return rl
}

// Allow checks if a request is allowed for the given user
func (rl *rateLimiter) Allow(userID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Get user's request history
	requests := rl.requests[userID]

	// Filter out requests outside the window
	validRequests := make([]time.Time, 0)
	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if limit is exceeded
	if len(validRequests) >= rl.limit {
		rl.requests[userID] = validRequests
		return false
	}

	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[userID] = validRequests

	return true
}

// cleanup periodically removes old entries to prevent memory leaks
func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		windowStart := now.Add(-rl.window)

		for userID, requests := range rl.requests {
			// Filter out old requests
			validRequests := make([]time.Time, 0)
			for _, reqTime := range requests {
				if reqTime.After(windowStart) {
					validRequests = append(validRequests, reqTime)
				}
			}

			// Remove user if no valid requests
			if len(validRequests) == 0 {
				delete(rl.requests, userID)
			} else {
				rl.requests[userID] = validRequests
			}
		}
		rl.mu.Unlock()
	}
}
