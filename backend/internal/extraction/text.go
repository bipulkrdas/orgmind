package extraction

import (
	"context"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// PlainTextExtractor handles plain text and markdown files
type PlainTextExtractor struct{}

// NewPlainTextExtractor creates a new plain text extractor
func NewPlainTextExtractor() *PlainTextExtractor {
	return &PlainTextExtractor{}
}

// Extract extracts text from plain text or markdown files
func (e *PlainTextExtractor) Extract(ctx context.Context, data []byte) (string, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Convert to UTF-8 if needed
	text, err := ensureUTF8(data)
	if err != nil {
		return "", err
	}

	// Normalize line endings for consistent processing
	text = normalizeLineEndings(text)

	// Normalize whitespace while preserving paragraph breaks
	text = normalizeWhitespace(text)

	return text, nil
}

// ensureUTF8 converts text to UTF-8 if it's not already
func ensureUTF8(data []byte) (string, error) {
	// Check if already valid UTF-8
	if utf8.Valid(data) {
		return string(data), nil
	}

	// Try to decode as UTF-16
	decoder := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
	result, _, err := transform.Bytes(decoder, data)
	if err == nil && utf8.Valid(result) {
		return string(result), nil
	}

	// If all else fails, replace invalid UTF-8 sequences
	return strings.ToValidUTF8(string(data), "�"), nil
}

// normalizeLineEndings converts all line endings to \n
func normalizeLineEndings(text string) string {
	// Replace Windows line endings
	text = strings.ReplaceAll(text, "\r\n", "\n")
	// Replace old Mac line endings
	text = strings.ReplaceAll(text, "\r", "\n")
	return text
}

// normalizeWhitespace normalizes whitespace while preserving paragraph breaks
func normalizeWhitespace(text string) string {
	// Split into lines
	lines := strings.Split(text, "\n")

	var result strings.Builder
	var prevEmpty bool

	for _, line := range lines {
		// Trim trailing whitespace
		line = strings.TrimRight(line, " \t")

		if line == "" {
			// Preserve one empty line for paragraph breaks
			if !prevEmpty {
				result.WriteString("\n")
				prevEmpty = true
			}
		} else {
			// Normalize internal whitespace
			line = normalizeInternalWhitespace(line)
			result.WriteString(line)
			result.WriteString("\n")
			prevEmpty = false
		}
	}

	return strings.TrimSpace(result.String())
}

// normalizeInternalWhitespace normalizes spaces and tabs within a line
func normalizeInternalWhitespace(line string) string {
	// Replace multiple spaces with single space
	var result strings.Builder
	var prevSpace bool

	for _, r := range line {
		if r == ' ' || r == '\t' {
			if !prevSpace {
				result.WriteRune(' ')
				prevSpace = true
			}
		} else {
			result.WriteRune(r)
			prevSpace = false
		}
	}

	return result.String()
}
