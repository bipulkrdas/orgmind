package extraction

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/ledongthuc/pdf"
)

// PDFExtractor handles PDF document extraction
type PDFExtractor struct{}

// NewPDFExtractor creates a new PDF extractor
func NewPDFExtractor() *PDFExtractor {
	return &PDFExtractor{}
}

// Extract extracts text from PDF files
func (e *PDFExtractor) Extract(ctx context.Context, data []byte) (string, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Validate PDF header
	if len(data) < 5 {
		return "", fmt.Errorf("%w: file too small to be a valid PDF", ErrCorruptedFile)
	}

	// Check for PDF magic number (%PDF-)
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		return "", fmt.Errorf("%w: invalid PDF header - file may be corrupted or not a PDF", ErrCorruptedFile)
	}

	// Create a reader from the byte slice
	reader := bytes.NewReader(data)

	// Open the PDF with error recovery
	pdfReader, err := pdf.NewReader(reader, int64(len(data)))
	if err != nil {
		// Provide descriptive error messages based on error type
		errMsg := err.Error()
		if strings.Contains(errMsg, "encrypted") || strings.Contains(errMsg, "password") {
			return "", fmt.Errorf("%w: PDF is password-protected", ErrPasswordProtected)
		}
		if strings.Contains(errMsg, "xref") || strings.Contains(errMsg, "trailer") {
			return "", fmt.Errorf("%w: PDF structure is corrupted (invalid xref table or trailer)", ErrCorruptedFile)
		}
		return "", fmt.Errorf("%w: failed to parse PDF - %v", ErrCorruptedFile, err)
	}

	// Get number of pages
	numPages := pdfReader.NumPage()
	if numPages == 0 {
		return "", nil // Empty PDF is valid, just return empty string
	}

	var result strings.Builder
	var extractionErrors []string

	// Extract text from each page
	for pageNum := 1; pageNum <= numPages; pageNum++ {
		// Check for context cancellation between pages
		select {
		case <-ctx.Done():
			// If we've extracted some text before timeout, return it
			if result.Len() > 0 {
				return normalizeWhitespace(result.String()), fmt.Errorf("%w: extracted %d of %d pages before timeout", ctx.Err(), pageNum-1, numPages)
			}
			return "", ctx.Err()
		default:
		}

		// Safely get the page
		page := pdfReader.Page(pageNum)
		if page.V.IsNull() {
			extractionErrors = append(extractionErrors, fmt.Sprintf("page %d is null", pageNum))
			continue
		}

		// Extract text from the page with error handling
		pageText, err := extractPageText(page)
		if err != nil {
			// Record error but continue with other pages
			extractionErrors = append(extractionErrors, fmt.Sprintf("page %d: %v", pageNum, err))
			continue
		}

		// Add page text with paragraph breaks
		if pageText != "" {
			result.WriteString(pageText)

			// Add double newline between pages to preserve page breaks
			if pageNum < numPages {
				result.WriteString("\n\n")
			}
		}
	}

	// Get the extracted text
	text := result.String()

	// If no text was extracted at all, return an error
	if text == "" && len(extractionErrors) > 0 {
		return "", fmt.Errorf("%w: failed to extract text from any page - errors: %v", ErrExtractionFailed, extractionErrors)
	}

	// Normalize whitespace while preserving paragraph breaks
	text = normalizeWhitespace(text)

	return text, nil
}

// extractPageText extracts text content from a PDF page
func extractPageText(page pdf.Page) (string, error) {
	// Try to get text by row (more structured approach)
	rows, err := page.GetTextByRow()
	if err != nil {
		// If GetTextByRow fails, try the basic content approach
		return extractPageTextBasic(page)
	}

	var result strings.Builder
	for _, row := range rows {
		for _, word := range row.Content {
			result.WriteString(word.S)
			result.WriteString(" ")
		}
		// Add newline after each row to preserve line structure
		if result.Len() > 0 {
			result.WriteString("\n")
		}
	}

	return result.String(), nil
}

// extractPageTextBasic is a fallback method for text extraction
func extractPageTextBasic(page pdf.Page) (string, error) {
	// Get the page content
	content := page.Content()

	var textBuilder strings.Builder

	// Parse the content stream
	for _, text := range content.Text {
		// Get the text string
		textStr := text.S
		if textStr == "" {
			continue
		}

		// Add text with space separator
		textBuilder.WriteString(textStr)
		textBuilder.WriteString(" ")
	}

	return textBuilder.String(), nil
}
