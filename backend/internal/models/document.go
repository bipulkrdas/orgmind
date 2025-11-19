package models

import "time"

// Document represents a document in the system
type Document struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"userId" db:"user_id"`
	GraphID      *string   `json:"graphId" db:"graph_id"`
	Filename     *string   `json:"filename" db:"filename"`
	ContentType  *string   `json:"contentType" db:"content_type"`
	StorageKey   string    `json:"storageKey" db:"storage_key"`
	SizeBytes    int64     `json:"sizeBytes" db:"size_bytes"`
	Source       string    `json:"source" db:"source"` // "editor" or "upload"
	Status       string    `json:"status" db:"status"`
	ErrorMessage *string   `json:"errorMessage,omitempty" db:"error_message"`
	GeminiFileID *string   `json:"geminiFileId,omitempty" db:"gemini_file_id"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}
