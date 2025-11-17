package extraction

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// JSONExtractor handles JSON files
type JSONExtractor struct{}

// NewJSONExtractor creates a new JSON extractor
func NewJSONExtractor() *JSONExtractor {
	return &JSONExtractor{}
}

// Extract extracts text from JSON files
func (e *JSONExtractor) Extract(ctx context.Context, data []byte) (string, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Parse JSON
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	// Extract text from JSON structure
	var result strings.Builder
	extractJSONText(jsonData, &result, "", 0, ctx)

	// Normalize whitespace
	text := normalizeWhitespace(result.String())

	return text, nil
}

// extractJSONText recursively extracts text from JSON structures
func extractJSONText(data interface{}, result *strings.Builder, key string, depth int, ctx context.Context) {
	// Check for context cancellation periodically
	select {
	case <-ctx.Done():
		return
	default:
	}

	// Limit depth to prevent stack overflow
	if depth > 20 {
		return
	}

	indent := strings.Repeat("  ", depth)

	switch v := data.(type) {
	case map[string]interface{}:
		// Object - extract key-value pairs
		for k, val := range v {
			// Write key as context
			if depth > 0 {
				result.WriteString(indent)
			}
			result.WriteString(k)
			result.WriteString(": ")

			// Handle different value types
			switch valTyped := val.(type) {
			case string:
				result.WriteString(valTyped)
				result.WriteString("\n")
			case map[string]interface{}, []interface{}:
				result.WriteString("\n")
				extractJSONText(val, result, k, depth+1, ctx)
			default:
				result.WriteString(fmt.Sprintf("%v", val))
				result.WriteString("\n")
			}
		}

	case []interface{}:
		// Array - extract items
		for i, item := range v {
			// Add array index as context
			if depth > 0 {
				result.WriteString(indent)
			}
			result.WriteString(fmt.Sprintf("[%d]: ", i))

			// Handle different item types
			switch itemTyped := item.(type) {
			case string:
				result.WriteString(itemTyped)
				result.WriteString("\n")
			case map[string]interface{}, []interface{}:
				result.WriteString("\n")
				extractJSONText(item, result, fmt.Sprintf("[%d]", i), depth+1, ctx)
			default:
				result.WriteString(fmt.Sprintf("%v", item))
				result.WriteString("\n")
			}
		}

	case string:
		// String value
		if key != "" {
			result.WriteString(indent)
			result.WriteString(key)
			result.WriteString(": ")
		}
		result.WriteString(v)
		result.WriteString("\n")

	case float64, int, bool:
		// Primitive values
		if key != "" {
			result.WriteString(indent)
			result.WriteString(key)
			result.WriteString(": ")
		}
		result.WriteString(fmt.Sprintf("%v", v))
		result.WriteString("\n")

	case nil:
		// Null value - skip
		return

	default:
		// Unknown type - convert to string
		if key != "" {
			result.WriteString(indent)
			result.WriteString(key)
			result.WriteString(": ")
		}
		result.WriteString(fmt.Sprintf("%v", v))
		result.WriteString("\n")
	}
}
