package extraction

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// CSVExtractor handles CSV files
type CSVExtractor struct{}

// NewCSVExtractor creates a new CSV extractor
func NewCSVExtractor() *CSVExtractor {
	return &CSVExtractor{}
}

// Extract extracts text from CSV files
func (e *CSVExtractor) Extract(ctx context.Context, data []byte) (string, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Try to detect delimiter
	delimiter := detectCSVDelimiter(data)

	// Create CSV reader
	reader := csv.NewReader(bytes.NewReader(data))
	reader.Comma = delimiter
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true // Be lenient with quotes

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return "", fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) == 0 {
		return "", nil
	}

	// Extract text from CSV
	var result strings.Builder

	// Check if first row is a header (heuristic: all strings, no numbers)
	hasHeader := isLikelyHeader(records[0])

	// Process records
	for i, record := range records {
		// Check for context cancellation periodically
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		if i == 0 && hasHeader {
			// Format header row
			result.WriteString("Headers: ")
			result.WriteString(strings.Join(record, ", "))
			result.WriteString("\n\n")
			continue
		}

		// Format data rows
		if hasHeader && i > 0 {
			// Format as key-value pairs using header
			result.WriteString(fmt.Sprintf("Row %d:\n", i))
			for j, cell := range record {
				if j < len(records[0]) {
					header := records[0][j]
					if header == "" {
						header = fmt.Sprintf("Column%d", j+1)
					}
					result.WriteString(fmt.Sprintf("  %s: %s\n", header, cell))
				}
			}
			result.WriteString("\n")
		} else {
			// Format as simple row
			result.WriteString(fmt.Sprintf("Row %d: ", i+1))
			result.WriteString(strings.Join(record, ", "))
			result.WriteString("\n")
		}
	}

	return strings.TrimSpace(result.String()), nil
}

// detectCSVDelimiter attempts to detect the CSV delimiter
func detectCSVDelimiter(data []byte) rune {
	// Read first few lines to detect delimiter
	reader := bytes.NewReader(data)
	scanner := io.LimitReader(reader, 1024) // Check first 1KB

	sample := make([]byte, 1024)
	n, _ := scanner.Read(sample)
	sampleStr := string(sample[:n])

	// Count common delimiters
	delimiters := []rune{',', ';', '\t', '|'}
	maxCount := 0
	bestDelimiter := ','

	for _, delim := range delimiters {
		count := strings.Count(sampleStr, string(delim))
		if count > maxCount {
			maxCount = count
			bestDelimiter = delim
		}
	}

	return bestDelimiter
}

// isLikelyHeader checks if a row is likely a header row
func isLikelyHeader(row []string) bool {
	if len(row) == 0 {
		return false
	}

	// Heuristic: headers typically don't contain only numbers
	numericCount := 0
	for _, cell := range row {
		cell = strings.TrimSpace(cell)
		if cell == "" {
			continue
		}

		// Check if cell is purely numeric
		isNumeric := true
		for _, r := range cell {
			if !((r >= '0' && r <= '9') || r == '.' || r == '-' || r == '+') {
				isNumeric = false
				break
			}
		}

		if isNumeric {
			numericCount++
		}
	}

	// If more than half the cells are numeric, probably not a header
	return numericCount < len(row)/2
}
