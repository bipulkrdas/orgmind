package extraction

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/nguyenthenguyen/docx"
)

// DocxExtractor handles .docx document extraction
type DocxExtractor struct{}

// NewDocxExtractor creates a new .docx extractor
func NewDocxExtractor() *DocxExtractor {
	return &DocxExtractor{}
}

// Extract extracts text from .docx files
func (e *DocxExtractor) Extract(ctx context.Context, data []byte) (string, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Validate file size
	if len(data) < 100 {
		return "", fmt.Errorf("%w: file too small to be a valid .docx document", ErrCorruptedFile)
	}

	// Check for ZIP magic number (docx is a ZIP archive)
	if !bytes.HasPrefix(data, []byte("PK")) {
		return "", fmt.Errorf("%w: invalid .docx header - file may be corrupted or not a .docx", ErrCorruptedFile)
	}

	// Create a reader from the byte slice
	reader := bytes.NewReader(data)

	// Read the .docx file
	doc, err := docx.ReadDocxFromMemory(reader, int64(len(data)))
	if err != nil {
		// Provide descriptive error messages
		errMsg := err.Error()
		if strings.Contains(errMsg, "encrypted") || strings.Contains(errMsg, "password") {
			return "", fmt.Errorf("%w: document is password-protected", ErrPasswordProtected)
		}
		if strings.Contains(errMsg, "zip") || strings.Contains(errMsg, "corrupt") {
			return "", fmt.Errorf("%w: document structure is corrupted", ErrCorruptedFile)
		}
		return "", fmt.Errorf("%w: failed to parse .docx - %v", ErrCorruptedFile, err)
	}

	// Check for context cancellation before processing
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	var result strings.Builder

	// Extract body text
	bodyText := doc.Editable().GetContent()
	if bodyText != "" {
		result.WriteString(bodyText)
	}

	// Get the extracted text
	text := result.String()

	// If no text was extracted, return empty string (valid for empty documents)
	if text == "" {
		return "", nil
	}

	// Normalize whitespace while preserving paragraph breaks
	text = normalizeWhitespace(text)

	return text, nil
}
