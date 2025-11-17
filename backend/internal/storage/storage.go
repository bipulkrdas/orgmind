package storage

import (
	"context"
	"io"
)

// StorageService defines the interface for document storage operations
type StorageService interface {
	// Upload uploads content to storage and returns the storage key
	Upload(ctx context.Context, userID string, documentID string, filename string, content io.Reader, contentType string) (string, error)
	
	// Download retrieves content from storage
	Download(ctx context.Context, storageKey string) (io.ReadCloser, error)
	
	// Delete removes content from storage
	Delete(ctx context.Context, storageKey string) error
	
	// GetURL returns a presigned URL for accessing the content
	GetURL(ctx context.Context, storageKey string, expirationMinutes int) (string, error)
}
