package extraction

import (
	"context"
	"time"
)

// ExtractionService defines the interface for extracting text from documents
type ExtractionService interface {
	// Extract text from document bytes with context for timeout support
	Extract(ctx context.Context, data []byte, contentType string) (string, error)

	// Check if format is supported
	IsSupported(contentType string) bool

	// Get list of supported formats
	SupportedFormats() []string
}

// Extractor is a format-specific extractor
type Extractor interface {
	// Extract text from document bytes
	Extract(ctx context.Context, data []byte) (string, error)
}

// ExtractionConfig holds configuration for text extraction
type ExtractionConfig struct {
	MaxFileSize       int64
	ExtractionTimeout time.Duration
	MaxConcurrent     int
	MaxMemoryPerFile  int64 // Maximum memory usage per file extraction
}

// DefaultConfig returns default extraction configuration
func DefaultConfig() *ExtractionConfig {
	return &ExtractionConfig{
		MaxFileSize:       50 * 1024 * 1024, // 50MB
		ExtractionTimeout: 30 * time.Second,
		MaxConcurrent:     10,
		MaxMemoryPerFile:  100 * 1024 * 1024, // 100MB per file
	}
}

// FormatInfo contains metadata about a supported format
type FormatInfo struct {
	Name       string
	Extensions []string
	MimeType   string
	Extractor  string
}
