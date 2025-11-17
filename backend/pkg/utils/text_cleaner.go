package utils

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	// Regex patterns for cleaning
	excessiveWhitespaceRegex = regexp.MustCompile(`[ \t]+`)
	excessiveNewlinesRegex   = regexp.MustCompile(`\n{3,}`)
	formattingArtifactsRegex = regexp.MustCompile(`[\x00-\x08\x0B-\x0C\x0E-\x1F\x7F]`) // Control characters except \t, \n, \r
)

// CleanText removes excessive whitespace, formatting artifacts, and normalizes line endings
// while preserving document structure and readability.
func CleanText(text string) string {
	if text == "" {
		return ""
	}

	// Step 1: Normalize line endings (convert \r\n and \r to \n)
	text = normalizeLineEndings(text)

	// Step 2: Remove control characters and formatting artifacts
	text = formattingArtifactsRegex.ReplaceAllString(text, "")

	// Step 3: Replace multiple spaces/tabs with single space
	text = excessiveWhitespaceRegex.ReplaceAllString(text, " ")

	// Step 4: Trim whitespace from each line while preserving structure
	text = trimLinesWhitespace(text)

	// Step 5: Replace excessive newlines (3+ consecutive) with double newline
	text = excessiveNewlinesRegex.ReplaceAllString(text, "\n\n")

	// Step 6: Trim leading and trailing whitespace from entire text
	text = strings.TrimSpace(text)

	return text
}

// normalizeLineEndings converts all line endings to \n
func normalizeLineEndings(text string) string {
	// Replace \r\n with \n
	text = strings.ReplaceAll(text, "\r\n", "\n")
	// Replace remaining \r with \n
	text = strings.ReplaceAll(text, "\r", "\n")
	return text
}

// trimLinesWhitespace trims leading and trailing whitespace from each line
func trimLinesWhitespace(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRightFunc(line, unicode.IsSpace)
	}
	return strings.Join(lines, "\n")
}
