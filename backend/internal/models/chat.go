package models

import (
	"fmt"
	"strings"
	"time"
)

// ChatThread represents a conversation session containing multiple messages
type ChatThread struct {
	ID        string    `json:"id" db:"id"`
	GraphID   string    `json:"graphId" db:"graph_id"`
	UserID    string    `json:"userId" db:"user_id"`
	Summary   *string   `json:"summary" db:"summary"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// Validate validates the ChatThread fields
func (ct *ChatThread) Validate() error {
	if ct.ID == "" {
		return fmt.Errorf("thread ID is required")
	}
	if ct.GraphID == "" {
		return fmt.Errorf("graph ID is required")
	}
	if ct.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if ct.Summary != nil && len(*ct.Summary) > 200 {
		return fmt.Errorf("summary must not exceed 200 characters")
	}
	return nil
}

// GenerateSummary creates a summary from the first message (max 100 characters)
func (ct *ChatThread) GenerateSummary(firstMessage string) {
	// Trim whitespace
	message := strings.TrimSpace(firstMessage)

	// Limit to 100 characters
	if len(message) > 100 {
		message = message[:97] + "..."
	}

	ct.Summary = &message
}

// ChatMessage represents a single message in a chat thread
type ChatMessage struct {
	ID        string    `json:"id" db:"id"`
	ThreadID  string    `json:"threadId" db:"thread_id"`
	Role      string    `json:"role" db:"role"` // "user" or "assistant"
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// Validate validates the ChatMessage fields
func (cm *ChatMessage) Validate() error {
	if cm.ID == "" {
		return fmt.Errorf("message ID is required")
	}
	if cm.ThreadID == "" {
		return fmt.Errorf("thread ID is required")
	}
	if err := cm.ValidateRole(); err != nil {
		return err
	}
	if err := cm.ValidateContent(); err != nil {
		return err
	}
	return nil
}

// ValidateRole validates that the role is either "user" or "assistant"
func (cm *ChatMessage) ValidateRole() error {
	if cm.Role != "user" && cm.Role != "assistant" {
		return fmt.Errorf("role must be either 'user' or 'assistant', got '%s'", cm.Role)
	}
	return nil
}

// ValidateContent validates the message content length
func (cm *ChatMessage) ValidateContent() error {
	if cm.Content == "" {
		return fmt.Errorf("message content is required")
	}
	if len(cm.Content) > 4000 {
		return fmt.Errorf("message content must not exceed 4000 characters")
	}
	return nil
}
