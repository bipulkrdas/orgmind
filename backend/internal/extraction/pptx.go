package extraction

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// PptxExtractor handles .pptx document extraction
type PptxExtractor struct{}

// NewPptxExtractor creates a new .pptx extractor
func NewPptxExtractor() *PptxExtractor {
	return &PptxExtractor{}
}

// Extract extracts text from .pptx files
func (e *PptxExtractor) Extract(ctx context.Context, data []byte) (string, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Validate file size
	if len(data) < 100 {
		return "", fmt.Errorf("%w: file too small to be a valid .pptx document", ErrCorruptedFile)
	}

	// Check for ZIP magic number (pptx is a ZIP archive)
	if !bytes.HasPrefix(data, []byte("PK")) {
		return "", fmt.Errorf("%w: invalid .pptx header - file may be corrupted or not a .pptx", ErrCorruptedFile)
	}

	// Create a reader from the byte slice
	reader := bytes.NewReader(data)

	// Open the .pptx file as a ZIP archive
	zipReader, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		// Provide descriptive error messages
		errMsg := err.Error()
		if strings.Contains(errMsg, "encrypted") || strings.Contains(errMsg, "password") {
			return "", fmt.Errorf("%w: document is password-protected", ErrPasswordProtected)
		}
		if strings.Contains(errMsg, "zip") || strings.Contains(errMsg, "corrupt") {
			return "", fmt.Errorf("%w: document structure is corrupted", ErrCorruptedFile)
		}
		return "", fmt.Errorf("%w: failed to parse .pptx - %v", ErrCorruptedFile, err)
	}

	// Check for context cancellation before processing
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	var result strings.Builder
	slideCount := 0

	// Extract text from all slides
	for _, file := range zipReader.File {
		// Check for context cancellation between files
		select {
		case <-ctx.Done():
			// If we've extracted some text before timeout, return it
			if result.Len() > 0 {
				return normalizeWhitespace(result.String()), fmt.Errorf("%w: extracted %d slides before timeout", ctx.Err(), slideCount)
			}
			return "", ctx.Err()
		default:
		}

		// Process slide files (ppt/slides/slideX.xml)
		if strings.HasPrefix(file.Name, "ppt/slides/slide") && strings.HasSuffix(file.Name, ".xml") {
			slideCount++

			// Open the slide file
			rc, err := file.Open()
			if err != nil {
				continue // Skip slides that can't be opened
			}

			// Read the slide content
			slideData, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue // Skip slides that can't be read
			}

			// Extract text from the slide XML
			slideText := extractTextFromPptxXML(slideData)
			if slideText != "" {
				result.WriteString(fmt.Sprintf("Slide %d:\n", slideCount))
				result.WriteString(slideText)
				result.WriteString("\n\n")
			}
		}

		// Process notes files (ppt/notesSlides/notesSlideX.xml)
		if strings.HasPrefix(file.Name, "ppt/notesSlides/notesSlide") && strings.HasSuffix(file.Name, ".xml") {
			// Open the notes file
			rc, err := file.Open()
			if err != nil {
				continue // Skip notes that can't be opened
			}

			// Read the notes content
			notesData, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue // Skip notes that can't be read
			}

			// Extract text from the notes XML
			notesText := extractTextFromPptxXML(notesData)
			if notesText != "" {
				result.WriteString("Notes:\n")
				result.WriteString(notesText)
				result.WriteString("\n\n")
			}
		}
	}

	// Get the extracted text
	text := result.String()

	// If no text was extracted, return empty string (valid for empty presentations)
	if text == "" {
		return "", nil
	}

	// Normalize whitespace while preserving paragraph breaks
	text = normalizeWhitespace(text)

	return text, nil
}

// extractTextFromPptxXML extracts text content from PowerPoint XML
func extractTextFromPptxXML(data []byte) string {
	// Simple XML structure for text extraction
	type TextElement struct {
		Text string `xml:",chardata"`
	}

	var result strings.Builder
	decoder := xml.NewDecoder(bytes.NewReader(data))

	// Parse XML and extract text from <a:t> elements (text runs)
	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch element := token.(type) {
		case xml.StartElement:
			// Look for text elements (a:t in PowerPoint XML)
			if element.Name.Local == "t" {
				var text TextElement
				if err := decoder.DecodeElement(&text, &element); err == nil {
					if text.Text != "" {
						result.WriteString(text.Text)
						result.WriteString(" ")
					}
				}
			}
		}
	}

	return strings.TrimSpace(result.String())
}
