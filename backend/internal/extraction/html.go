package extraction

import (
	"context"
	"strings"

	"golang.org/x/net/html"
)

// HTMLExtractor handles HTML files
type HTMLExtractor struct{}

// NewHTMLExtractor creates a new HTML extractor
func NewHTMLExtractor() *HTMLExtractor {
	return &HTMLExtractor{}
}

// Extract extracts text from HTML files
func (e *HTMLExtractor) Extract(ctx context.Context, data []byte) (string, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Parse HTML
	doc, err := html.Parse(strings.NewReader(string(data)))
	if err != nil {
		return "", err
	}

	// Extract text from HTML nodes
	var result strings.Builder
	extractText(doc, &result, ctx)

	// Normalize whitespace
	text := normalizeWhitespace(result.String())

	return text, nil
}

// extractText recursively extracts text from HTML nodes
func extractText(n *html.Node, result *strings.Builder, ctx context.Context) {
	// Check for context cancellation periodically
	select {
	case <-ctx.Done():
		return
	default:
	}

	// Skip script and style tags
	if n.Type == html.ElementNode {
		switch n.Data {
		case "script", "style", "noscript":
			return
		case "br":
			result.WriteString("\n")
			return
		case "p", "div", "h1", "h2", "h3", "h4", "h5", "h6":
			// Add line breaks for block elements
			if result.Len() > 0 && !strings.HasSuffix(result.String(), "\n") {
				result.WriteString("\n")
			}
		case "li":
			// Add bullet for list items
			result.WriteString("• ")
		}
	}

	// Extract text nodes
	if n.Type == html.TextNode {
		text := n.Data
		// Decode HTML entities (already done by html.Parse)
		// Trim excessive whitespace but preserve single spaces
		text = strings.TrimSpace(text)
		if text != "" {
			result.WriteString(text)
			result.WriteString(" ")
		}
	}

	// Recursively process child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractText(c, result, ctx)
	}

	// Add line breaks after block elements
	if n.Type == html.ElementNode {
		switch n.Data {
		case "p", "div", "h1", "h2", "h3", "h4", "h5", "h6", "li", "tr":
			result.WriteString("\n")
		}
	}
}
