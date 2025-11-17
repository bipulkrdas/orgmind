package extraction

import (
	"errors"
	"fmt"
)

// Base extraction errors
var (
	ErrUnsupportedFormat = errors.New("unsupported document format")
	ErrCorruptedFile     = errors.New("file appears to be corrupted")
	ErrPasswordProtected = errors.New("document is password protected")
	ErrExtractionFailed  = errors.New("text extraction failed")
	ErrFileTooLarge      = errors.New("file exceeds maximum size")
	ErrExtractionTimeout = errors.New("extraction timeout exceeded")
	ErrInvalidFormat     = errors.New("file format validation failed")
	ErrEmptyFile         = errors.New("file is empty")
	ErrMemoryLimit       = errors.New("extraction exceeded memory limit")
)

// ExtractionError wraps extraction errors with additional context
type ExtractionError struct {
	// Type of error (e.g., "unsupported_format", "corrupted_file")
	Type string

	// Original error
	Err error

	// Content type that was being processed
	ContentType string

	// File size in bytes
	FileSize int64

	// Filename if available
	Filename string

	// User-friendly message
	UserMessage string

	// Technical details for logging
	TechnicalDetails string

	// Whether this error is retryable
	Retryable bool
}

// Error implements the error interface
func (e *ExtractionError) Error() string {
	if e.TechnicalDetails != "" {
		return fmt.Sprintf("%s: %s (details: %s)", e.Type, e.Err.Error(), e.TechnicalDetails)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Err.Error())
}

// Unwrap returns the underlying error
func (e *ExtractionError) Unwrap() error {
	return e.Err
}

// NewExtractionError creates a new extraction error with context
func NewExtractionError(errType string, err error, contentType string, fileSize int64) *ExtractionError {
	return &ExtractionError{
		Type:        errType,
		Err:         err,
		ContentType: contentType,
		FileSize:    fileSize,
		Retryable:   false,
	}
}

// WithFilename adds filename context to the error
func (e *ExtractionError) WithFilename(filename string) *ExtractionError {
	e.Filename = filename
	return e
}

// WithUserMessage sets a user-friendly message
func (e *ExtractionError) WithUserMessage(msg string) *ExtractionError {
	e.UserMessage = msg
	return e
}

// WithTechnicalDetails adds technical details for logging
func (e *ExtractionError) WithTechnicalDetails(details string) *ExtractionError {
	e.TechnicalDetails = details
	return e
}

// WithRetryable marks the error as retryable
func (e *ExtractionError) WithRetryable(retryable bool) *ExtractionError {
	e.Retryable = retryable
	return e
}

// Error type constants for consistent error handling
const (
	ErrTypeUnsupportedFormat = "unsupported_format"
	ErrTypeCorruptedFile     = "corrupted_file"
	ErrTypePasswordProtected = "password_protected"
	ErrTypeExtractionFailed  = "extraction_failed"
	ErrTypeFileTooLarge      = "file_too_large"
	ErrTypeTimeout           = "extraction_timeout"
	ErrTypeInvalidFormat     = "invalid_format"
	ErrTypeEmptyFile         = "empty_file"
	ErrTypeMemoryLimit       = "memory_limit_exceeded"
)

// WrapUnsupportedFormat wraps an unsupported format error with context
func WrapUnsupportedFormat(contentType string, fileSize int64, supportedFormats []string) *ExtractionError {
	err := NewExtractionError(
		ErrTypeUnsupportedFormat,
		ErrUnsupportedFormat,
		contentType,
		fileSize,
	)

	formatList := "PDF, Word (.docx), Excel (.xlsx), PowerPoint (.pptx), Plain Text, Markdown, HTML, JSON, CSV, EPUB, RTF"
	if len(supportedFormats) > 0 {
		formatList = fmt.Sprintf("%v", supportedFormats)
	}

	err.WithUserMessage(fmt.Sprintf(
		"The file format '%s' is not supported. Supported formats: %s",
		contentType,
		formatList,
	))
	err.WithTechnicalDetails(fmt.Sprintf("Content-Type: %s, Size: %d bytes", contentType, fileSize))

	return err
}

// WrapCorruptedFile wraps a corrupted file error with context
func WrapCorruptedFile(contentType string, fileSize int64, filename string, originalErr error) *ExtractionError {
	err := NewExtractionError(
		ErrTypeCorruptedFile,
		ErrCorruptedFile,
		contentType,
		fileSize,
	)

	err.WithFilename(filename)
	err.WithUserMessage(
		"The file appears to be corrupted or invalid. Please check the file and try uploading again.",
	)
	err.WithTechnicalDetails(fmt.Sprintf(
		"File: %s, Content-Type: %s, Size: %d bytes, Original error: %v",
		filename,
		contentType,
		fileSize,
		originalErr,
	))

	return err
}

// WrapPasswordProtected wraps a password-protected error with context
func WrapPasswordProtected(contentType string, fileSize int64, filename string) *ExtractionError {
	err := NewExtractionError(
		ErrTypePasswordProtected,
		ErrPasswordProtected,
		contentType,
		fileSize,
	)

	err.WithFilename(filename)
	err.WithUserMessage(
		"This document is password-protected. Please remove the password and try again.",
	)
	err.WithTechnicalDetails(fmt.Sprintf(
		"File: %s, Content-Type: %s, Size: %d bytes",
		filename,
		contentType,
		fileSize,
	))

	return err
}

// WrapExtractionTimeout wraps a timeout error with context
func WrapExtractionTimeout(contentType string, fileSize int64, filename string, timeoutDuration string) *ExtractionError {
	err := NewExtractionError(
		ErrTypeTimeout,
		ErrExtractionTimeout,
		contentType,
		fileSize,
	)

	err.WithFilename(filename)
	err.WithUserMessage(
		"Text extraction took too long to complete. Please try with a smaller file or a simpler document.",
	)
	err.WithTechnicalDetails(fmt.Sprintf(
		"File: %s, Content-Type: %s, Size: %d bytes, Timeout: %s",
		filename,
		contentType,
		fileSize,
		timeoutDuration,
	))
	err.WithRetryable(true)

	return err
}

// WrapFileTooLarge wraps a file size error with context
func WrapFileTooLarge(contentType string, fileSize int64, filename string, maxSize int64) *ExtractionError {
	err := NewExtractionError(
		ErrTypeFileTooLarge,
		ErrFileTooLarge,
		contentType,
		fileSize,
	)

	err.WithFilename(filename)
	err.WithUserMessage(fmt.Sprintf(
		"File size (%d MB) exceeds the maximum allowed size of %d MB. Please upload a smaller file.",
		fileSize/(1024*1024),
		maxSize/(1024*1024),
	))
	err.WithTechnicalDetails(fmt.Sprintf(
		"File: %s, Content-Type: %s, Size: %d bytes, Max: %d bytes",
		filename,
		contentType,
		fileSize,
		maxSize,
	))

	return err
}

// WrapInvalidFormat wraps a format validation error with context
func WrapInvalidFormat(contentType string, fileSize int64, filename string, reason string) *ExtractionError {
	err := NewExtractionError(
		ErrTypeInvalidFormat,
		ErrInvalidFormat,
		contentType,
		fileSize,
	)

	err.WithFilename(filename)
	err.WithUserMessage(fmt.Sprintf(
		"The file format doesn't match its extension. %s",
		reason,
	))
	err.WithTechnicalDetails(fmt.Sprintf(
		"File: %s, Content-Type: %s, Size: %d bytes, Reason: %s",
		filename,
		contentType,
		fileSize,
		reason,
	))

	return err
}

// WrapEmptyFile wraps an empty file error with context
func WrapEmptyFile(contentType string, filename string) *ExtractionError {
	err := NewExtractionError(
		ErrTypeEmptyFile,
		ErrEmptyFile,
		contentType,
		0,
	)

	err.WithFilename(filename)
	err.WithUserMessage(
		"The file is empty. Please upload a file with content.",
	)
	err.WithTechnicalDetails(fmt.Sprintf(
		"File: %s, Content-Type: %s",
		filename,
		contentType,
	))

	return err
}

// WrapMemoryLimit wraps a memory limit error with context
func WrapMemoryLimit(contentType string, fileSize int64, filename string) *ExtractionError {
	err := NewExtractionError(
		ErrTypeMemoryLimit,
		ErrMemoryLimit,
		contentType,
		fileSize,
	)

	err.WithFilename(filename)
	err.WithUserMessage(
		"The file is too complex to process. Please try with a simpler document or contact support.",
	)
	err.WithTechnicalDetails(fmt.Sprintf(
		"File: %s, Content-Type: %s, Size: %d bytes",
		filename,
		contentType,
		fileSize,
	))
	err.WithRetryable(true)

	return err
}

// WrapGenericExtractionError wraps a generic extraction error with context
func WrapGenericExtractionError(contentType string, fileSize int64, filename string, originalErr error) *ExtractionError {
	err := NewExtractionError(
		ErrTypeExtractionFailed,
		ErrExtractionFailed,
		contentType,
		fileSize,
	)

	err.WithFilename(filename)
	err.WithUserMessage(
		"Failed to extract text from the document. The file may be corrupted or in an unsupported format variant.",
	)
	err.WithTechnicalDetails(fmt.Sprintf(
		"File: %s, Content-Type: %s, Size: %d bytes, Error: %v",
		filename,
		contentType,
		fileSize,
		originalErr,
	))
	err.WithRetryable(true)

	return err
}

// GetUserFriendlyMessage returns a user-friendly error message for any error
func GetUserFriendlyMessage(err error) string {
	// Check if it's already an ExtractionError
	var extractionErr *ExtractionError
	if errors.As(err, &extractionErr) {
		if extractionErr.UserMessage != "" {
			return extractionErr.UserMessage
		}
	}

	// Map standard errors to user-friendly messages
	switch {
	case errors.Is(err, ErrUnsupportedFormat):
		return "This file format is not supported. Please upload a PDF, Word, Excel, PowerPoint, or text document."
	case errors.Is(err, ErrCorruptedFile):
		return "The file appears to be corrupted or invalid. Please check the file and try again."
	case errors.Is(err, ErrPasswordProtected):
		return "This document is password-protected. Please remove the password and try again."
	case errors.Is(err, ErrExtractionTimeout):
		return "Text extraction took too long. Please try with a smaller file."
	case errors.Is(err, ErrFileTooLarge):
		return "File size exceeds the maximum allowed size of 50MB."
	case errors.Is(err, ErrInvalidFormat):
		return "The file format doesn't match its extension. Please check the file."
	case errors.Is(err, ErrEmptyFile):
		return "The file is empty. Please upload a file with content."
	case errors.Is(err, ErrMemoryLimit):
		return "The file is too complex to process. Please try with a simpler document."
	default:
		return "Failed to process the document. Please try again or contact support if the problem persists."
	}
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	var extractionErr *ExtractionError
	if errors.As(err, &extractionErr) {
		return extractionErr.Retryable
	}

	// Timeouts and memory limits are generally retryable
	return errors.Is(err, ErrExtractionTimeout) || errors.Is(err, ErrMemoryLimit)
}
