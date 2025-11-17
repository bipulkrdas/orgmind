package extraction

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// XlsxExtractor handles .xlsx document extraction
type XlsxExtractor struct{}

// NewXlsxExtractor creates a new .xlsx extractor
func NewXlsxExtractor() *XlsxExtractor {
	return &XlsxExtractor{}
}

// Extract extracts text from .xlsx files
func (e *XlsxExtractor) Extract(ctx context.Context, data []byte) (string, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Validate file size
	if len(data) < 100 {
		return "", fmt.Errorf("%w: file too small to be a valid .xlsx document", ErrCorruptedFile)
	}

	// Check for ZIP magic number (xlsx is a ZIP archive)
	if !bytes.HasPrefix(data, []byte("PK")) {
		return "", fmt.Errorf("%w: invalid .xlsx header - file may be corrupted or not a .xlsx", ErrCorruptedFile)
	}

	// Create a reader from the byte slice
	reader := bytes.NewReader(data)

	// Open the .xlsx file
	file, err := excelize.OpenReader(reader)
	if err != nil {
		// Provide descriptive error messages
		errMsg := err.Error()
		if strings.Contains(errMsg, "encrypted") || strings.Contains(errMsg, "password") {
			return "", fmt.Errorf("%w: document is password-protected", ErrPasswordProtected)
		}
		if strings.Contains(errMsg, "zip") || strings.Contains(errMsg, "corrupt") {
			return "", fmt.Errorf("%w: document structure is corrupted", ErrCorruptedFile)
		}
		return "", fmt.Errorf("%w: failed to parse .xlsx - %v", ErrCorruptedFile, err)
	}
	defer file.Close()

	// Check for context cancellation before processing
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	var result strings.Builder

	// Get all sheet names
	sheetNames := file.GetSheetList()
	if len(sheetNames) == 0 {
		return "", nil // Empty workbook is valid
	}

	// Extract text from all sheets
	for sheetIndex, sheetName := range sheetNames {
		// Check for context cancellation between sheets
		select {
		case <-ctx.Done():
			// If we've extracted some text before timeout, return it
			if result.Len() > 0 {
				return normalizeWhitespace(result.String()), fmt.Errorf("%w: extracted %d of %d sheets before timeout", ctx.Err(), sheetIndex, len(sheetNames))
			}
			return "", ctx.Err()
		default:
		}

		// Add sheet name as a header for context
		if len(sheetNames) > 1 {
			result.WriteString(fmt.Sprintf("Sheet: %s\n", sheetName))
		}

		// Get all rows from the sheet
		rows, err := file.GetRows(sheetName)
		if err != nil {
			// Skip sheets that can't be read but continue with others
			continue
		}

		// Extract text from each row
		for _, row := range rows {
			// Check if row has any non-empty cells
			hasContent := false
			var rowText strings.Builder

			for cellIndex, cellValue := range row {
				// Trim whitespace from cell value
				cellValue = strings.TrimSpace(cellValue)

				if cellValue != "" {
					hasContent = true

					// Add cell value with delimiter
					if cellIndex > 0 {
						rowText.WriteString(" | ")
					}
					rowText.WriteString(cellValue)
				}
			}

			// Add row text if it has content
			if hasContent {
				result.WriteString(rowText.String())
				result.WriteString("\n")
			}
		}

		// Add spacing between sheets
		if sheetIndex < len(sheetNames)-1 {
			result.WriteString("\n")
		}
	}

	// Get the extracted text
	text := result.String()

	// If no text was extracted, return empty string (valid for empty workbooks)
	if text == "" {
		return "", nil
	}

	// Normalize whitespace while preserving line breaks
	text = normalizeWhitespace(text)

	return text, nil
}
