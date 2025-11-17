package extraction

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
)

// FileSignature represents a file format signature (magic number)
type FileSignature struct {
	Offset    int
	Signature []byte
	MimeType  string
}

// Common file signatures for supported formats
var fileSignatures = []FileSignature{
	// PDF
	{Offset: 0, Signature: []byte("%PDF-"), MimeType: "application/pdf"},

	// ZIP-based formats (DOCX, XLSX, PPTX, EPUB)
	{Offset: 0, Signature: []byte("PK\x03\x04"), MimeType: "application/zip"},

	// RTF
	{Offset: 0, Signature: []byte("{\\rtf"), MimeType: "application/rtf"},

	// HTML
	{Offset: 0, Signature: []byte("<!DOCTYPE html"), MimeType: "text/html"},
	{Offset: 0, Signature: []byte("<!doctype html"), MimeType: "text/html"},
	{Offset: 0, Signature: []byte("<html"), MimeType: "text/html"},
	{Offset: 0, Signature: []byte("<HTML"), MimeType: "text/html"},

	// JSON
	{Offset: 0, Signature: []byte("{"), MimeType: "application/json"},
	{Offset: 0, Signature: []byte("["), MimeType: "application/json"},
}

// ValidateFormat validates that the file extension matches the actual file content
func ValidateFormat(data []byte, filename string, declaredContentType string) error {
	// Get file extension
	ext := strings.ToLower(filepath.Ext(filename))

	// Detect actual content type from file header
	detectedContentType := DetectContentType(data)

	// Normalize content types for comparison
	declaredContentType = normalizeContentType(declaredContentType)
	detectedContentType = normalizeContentType(detectedContentType)

	// Special handling for ZIP-based formats
	if detectedContentType == "application/zip" {
		return validateZipBasedFormat(data, ext, declaredContentType)
	}

	// For non-ZIP formats, check if detected type matches declared type
	if detectedContentType != "" && declaredContentType != "" {
		if !isCompatibleContentType(detectedContentType, declaredContentType) {
			return fmt.Errorf("file extension %s does not match content type (expected %s, detected %s)", ext, declaredContentType, detectedContentType)
		}
	}

	// Validate extension matches declared content type
	if !isValidExtensionForContentType(ext, declaredContentType) {
		return fmt.Errorf("file extension %s is not valid for content type %s", ext, declaredContentType)
	}

	return nil
}

// DetectContentType detects the content type from file header (magic number)
func DetectContentType(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// Check against known signatures
	for _, sig := range fileSignatures {
		if len(data) >= sig.Offset+len(sig.Signature) {
			if bytes.Equal(data[sig.Offset:sig.Offset+len(sig.Signature)], sig.Signature) {
				return sig.MimeType
			}
		}
	}

	// Check if it's plain text (all printable ASCII or UTF-8)
	if isPlainText(data) {
		return "text/plain"
	}

	return ""
}

// validateZipBasedFormat validates ZIP-based formats (DOCX, XLSX, PPTX, EPUB)
func validateZipBasedFormat(data []byte, ext string, declaredContentType string) error {
	// Map extensions to their expected content types
	zipBasedFormats := map[string]string{
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".epub": "application/epub+zip",
	}

	expectedContentType, isZipBased := zipBasedFormats[ext]

	if isZipBased {
		// If extension suggests a ZIP-based format, verify it's actually a ZIP
		if !bytes.HasPrefix(data, []byte("PK\x03\x04")) {
			return fmt.Errorf("file extension %s suggests a ZIP-based format, but file is not a valid ZIP archive", ext)
		}

		// Check if declared content type matches the extension
		if declaredContentType != "" && declaredContentType != expectedContentType {
			return fmt.Errorf("file extension %s does not match declared content type %s", ext, declaredContentType)
		}

		return nil
	}

	// If it's a ZIP but extension doesn't match any known ZIP-based format
	if declaredContentType == "application/zip" {
		return fmt.Errorf("file appears to be a ZIP archive but has unexpected extension %s", ext)
	}

	return nil
}

// isCompatibleContentType checks if two content types are compatible
func isCompatibleContentType(detected, declared string) bool {
	// Exact match
	if detected == declared {
		return true
	}

	// RTF can be either application/rtf or text/rtf
	if (detected == "application/rtf" || detected == "text/rtf") &&
		(declared == "application/rtf" || declared == "text/rtf") {
		return true
	}

	// HTML variations
	if (detected == "text/html" || detected == "application/xhtml+xml") &&
		(declared == "text/html" || declared == "application/xhtml+xml") {
		return true
	}

	// JSON can be application/json or text/json
	if (detected == "application/json" || detected == "text/json") &&
		(declared == "application/json" || declared == "text/json") {
		return true
	}

	return false
}

// isValidExtensionForContentType checks if an extension is valid for a content type
func isValidExtensionForContentType(ext, contentType string) bool {
	validExtensions := map[string][]string{
		"application/pdf": {".pdf"},
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   {".docx"},
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         {".xlsx"},
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": {".pptx"},
		"application/epub+zip": {".epub"},
		"application/rtf":      {".rtf"},
		"text/rtf":             {".rtf"},
		"text/plain":           {".txt", ".text"},
		"text/markdown":        {".md", ".markdown"},
		"text/html":            {".html", ".htm"},
		"application/json":     {".json"},
		"text/csv":             {".csv"},
	}

	extensions, exists := validExtensions[contentType]
	if !exists {
		// Unknown content type, allow any extension
		return true
	}

	for _, validExt := range extensions {
		if ext == validExt {
			return true
		}
	}

	return false
}

// isPlainText checks if data appears to be plain text
func isPlainText(data []byte) bool {
	// Sample first 512 bytes
	sampleSize := 512
	if len(data) < sampleSize {
		sampleSize = len(data)
	}

	sample := data[:sampleSize]

	// Count printable characters
	printable := 0
	for _, b := range sample {
		// Printable ASCII or common whitespace
		if (b >= 32 && b <= 126) || b == '\t' || b == '\n' || b == '\r' {
			printable++
		} else if b >= 128 {
			// Could be UTF-8, allow it
			printable++
		}
	}

	// If more than 95% is printable, consider it text
	return float64(printable)/float64(len(sample)) > 0.95
}

// GetExpectedContentType returns the expected content type for a file extension
func GetExpectedContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	extensionToContentType := map[string]string{
		".pdf":      "application/pdf",
		".docx":     "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xlsx":     "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".pptx":     "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".epub":     "application/epub+zip",
		".rtf":      "application/rtf",
		".txt":      "text/plain",
		".text":     "text/plain",
		".md":       "text/markdown",
		".markdown": "text/markdown",
		".html":     "text/html",
		".htm":      "text/html",
		".json":     "application/json",
		".csv":      "text/csv",
	}

	if contentType, exists := extensionToContentType[ext]; exists {
		return contentType
	}

	return ""
}
