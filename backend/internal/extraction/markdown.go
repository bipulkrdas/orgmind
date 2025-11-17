package extraction

import (
	"context"
	"regexp"
	"strings"
)

// MarkdownExtractor handles markdown files
type MarkdownExtractor struct{}

// NewMarkdownExtractor creates a new markdown extractor
func NewMarkdownExtractor() *MarkdownExtractor {
	return &MarkdownExtractor{}
}

// Extract extracts text from markdown files while preserving structure
func (e *MarkdownExtractor) Extract(ctx context.Context, data []byte) (string, error) {
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

	// Normalize line endings
	text = normalizeLineEndings(text)

	// Process markdown while preserving structure
	text = processMarkdown(text)

	// Normalize whitespace while preserving paragraph breaks
	text = normalizeWhitespace(text)

	return text, nil
}

// processMarkdown processes markdown syntax while preserving semantic structure
func processMarkdown(text string) string {
	lines := strings.Split(text, "\n")
	var result strings.Builder

	inCodeBlock := false
	codeBlockDelimiter := ""

	for _, line := range lines {
		// Handle code blocks
		if strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~") {
			if !inCodeBlock {
				// Starting code block
				inCodeBlock = true
				codeBlockDelimiter = line[:3]
				// Add marker for code block start
				result.WriteString("\n[CODE BLOCK]\n")
				continue
			} else if strings.HasPrefix(line, codeBlockDelimiter) {
				// Ending code block
				inCodeBlock = false
				codeBlockDelimiter = ""
				result.WriteString("[END CODE BLOCK]\n\n")
				continue
			}
		}

		// Inside code block - preserve content with minimal formatting
		if inCodeBlock {
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		// Process markdown syntax outside code blocks
		line = processMarkdownLine(line)
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}

// processMarkdownLine processes a single line of markdown
func processMarkdownLine(line string) string {
	// Preserve headings by converting to natural language
	if strings.HasPrefix(line, "#") {
		// Count heading level
		level := 0
		for i := 0; i < len(line) && line[i] == '#'; i++ {
			level++
		}
		if level <= 6 && len(line) > level && line[level] == ' ' {
			// Extract heading text and preserve with extra line break for emphasis
			heading := strings.TrimSpace(line[level:])
			return "\n" + heading + "\n"
		}
	}

	// Convert list items to natural text
	line = convertListItems(line)

	// Remove inline code backticks but preserve content
	line = removeInlineCode(line)

	// Remove emphasis markers but preserve text
	line = removeEmphasis(line)

	// Convert links to readable format [text](url) -> text (url)
	line = convertLinks(line)

	// Remove image syntax but keep alt text
	line = removeImages(line)

	return line
}

// convertListItems converts markdown list items to plain text
func convertListItems(line string) string {
	trimmed := strings.TrimLeft(line, " \t")
	indent := len(line) - len(trimmed)

	// Unordered lists
	if strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "+ ") {
		return strings.Repeat(" ", indent) + "• " + trimmed[2:]
	}

	// Ordered lists
	re := regexp.MustCompile(`^(\d+)\.\s+(.*)$`)
	if matches := re.FindStringSubmatch(trimmed); matches != nil {
		return strings.Repeat(" ", indent) + matches[1] + ". " + matches[2]
	}

	return line
}

// removeInlineCode removes backticks but preserves code content
func removeInlineCode(line string) string {
	// Match single backticks
	re := regexp.MustCompile("`([^`]+)`")
	return re.ReplaceAllString(line, "$1")
}

// removeEmphasis removes emphasis markers but preserves text
func removeEmphasis(line string) string {
	// Remove bold (**text** or __text__)
	line = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(line, "$1")
	line = regexp.MustCompile(`__([^_]+)__`).ReplaceAllString(line, "$1")

	// Remove italic (*text* or _text_)
	line = regexp.MustCompile(`\*([^*]+)\*`).ReplaceAllString(line, "$1")
	line = regexp.MustCompile(`_([^_]+)_`).ReplaceAllString(line, "$1")

	// Remove strikethrough (~~text~~)
	line = regexp.MustCompile(`~~([^~]+)~~`).ReplaceAllString(line, "$1")

	return line
}

// convertLinks converts markdown links to readable format
func convertLinks(line string) string {
	// Convert [text](url) to text (url)
	re := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	return re.ReplaceAllString(line, "$1 ($2)")
}

// removeImages removes image syntax but keeps alt text
func removeImages(line string) string {
	// Convert ![alt](url) to [Image: alt]
	re := regexp.MustCompile(`!\[([^\]]*)\]\([^)]+\)`)
	return re.ReplaceAllString(line, "[Image: $1]")
}
