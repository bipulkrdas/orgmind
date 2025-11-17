package extraction

import (
	"context"
	"strings"
	"time"
)

// ExtractionRouter routes extraction requests to format-specific extractors
type ExtractionRouter struct {
	extractors map[string]Extractor
	formats    map[string]FormatInfo
	config     *ExtractionConfig
	queue      *ExtractionQueue
	logger     *ExtractionLogger
	stats      *ExtractionStats
}

// NewExtractionRouter creates a new extraction router
func NewExtractionRouter(config *ExtractionConfig) *ExtractionRouter {
	if config == nil {
		config = DefaultConfig()
	}

	router := &ExtractionRouter{
		extractors: make(map[string]Extractor),
		formats:    make(map[string]FormatInfo),
		config:     config,
		queue:      NewExtractionQueue(config.MaxConcurrent),
		logger:     NewExtractionLogger(true), // Enable logging by default
		stats:      NewExtractionStats(),
	}

	// Register all extractors
	router.registerExtractors()

	return router
}

// registerExtractors registers all format-specific extractors
func (r *ExtractionRouter) registerExtractors() {
	// Text formats
	plainTextExtractor := NewPlainTextExtractor()
	r.Register("text/plain", plainTextExtractor, FormatInfo{
		Name:       "Plain Text",
		Extensions: []string{".txt"},
		MimeType:   "text/plain",
		Extractor:  "PlainTextExtractor",
	})

	// Markdown
	markdownExtractor := NewMarkdownExtractor()
	r.Register("text/markdown", markdownExtractor, FormatInfo{
		Name:       "Markdown",
		Extensions: []string{".md", ".markdown"},
		MimeType:   "text/markdown",
		Extractor:  "MarkdownExtractor",
	})

	// HTML
	htmlExtractor := NewHTMLExtractor()
	r.Register("text/html", htmlExtractor, FormatInfo{
		Name:       "HTML",
		Extensions: []string{".html", ".htm"},
		MimeType:   "text/html",
		Extractor:  "HTMLExtractor",
	})

	// JSON
	jsonExtractor := NewJSONExtractor()
	r.Register("application/json", jsonExtractor, FormatInfo{
		Name:       "JSON",
		Extensions: []string{".json"},
		MimeType:   "application/json",
		Extractor:  "JSONExtractor",
	})

	// CSV
	csvExtractor := NewCSVExtractor()
	r.Register("text/csv", csvExtractor, FormatInfo{
		Name:       "CSV",
		Extensions: []string{".csv"},
		MimeType:   "text/csv",
		Extractor:  "CSVExtractor",
	})

	// PDF
	pdfExtractor := NewPDFExtractor()
	r.Register("application/pdf", pdfExtractor, FormatInfo{
		Name:       "PDF Document",
		Extensions: []string{".pdf"},
		MimeType:   "application/pdf",
		Extractor:  "PDFExtractor",
	})

	// Microsoft Office - Word
	docxExtractor := NewDocxExtractor()
	r.Register("application/vnd.openxmlformats-officedocument.wordprocessingml.document", docxExtractor, FormatInfo{
		Name:       "Word Document",
		Extensions: []string{".docx"},
		MimeType:   "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		Extractor:  "DocxExtractor",
	})

	// Microsoft Office - Excel
	xlsxExtractor := NewXlsxExtractor()
	r.Register("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", xlsxExtractor, FormatInfo{
		Name:       "Excel Spreadsheet",
		Extensions: []string{".xlsx"},
		MimeType:   "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Extractor:  "XlsxExtractor",
	})

	// Microsoft Office - PowerPoint
	pptxExtractor := NewPptxExtractor()
	r.Register("application/vnd.openxmlformats-officedocument.presentationml.presentation", pptxExtractor, FormatInfo{
		Name:       "PowerPoint Presentation",
		Extensions: []string{".pptx"},
		MimeType:   "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		Extractor:  "PptxExtractor",
	})

	// EPUB
	epubExtractor := NewEPUBExtractor()
	r.Register("application/epub+zip", epubExtractor, FormatInfo{
		Name:       "EPUB Document",
		Extensions: []string{".epub"},
		MimeType:   "application/epub+zip",
		Extractor:  "EPUBExtractor",
	})

	// RTF
	rtfExtractor := NewRTFExtractor()
	r.Register("application/rtf", rtfExtractor, FormatInfo{
		Name:       "Rich Text Format",
		Extensions: []string{".rtf"},
		MimeType:   "application/rtf",
		Extractor:  "RTFExtractor",
	})
	// Also register text/rtf as some systems use this MIME type
	r.Register("text/rtf", rtfExtractor, FormatInfo{
		Name:       "Rich Text Format",
		Extensions: []string{".rtf"},
		MimeType:   "text/rtf",
		Extractor:  "RTFExtractor",
	})
}

// Register adds a format-specific extractor
func (r *ExtractionRouter) Register(contentType string, extractor Extractor, info FormatInfo) {
	r.extractors[contentType] = extractor
	r.formats[contentType] = info
}

// Extract routes the extraction request to the appropriate extractor
func (r *ExtractionRouter) Extract(ctx context.Context, data []byte, contentType string) (string, error) {
	// Validate file size
	fileSize := int64(len(data))
	if fileSize == 0 {
		return "", WrapEmptyFile(contentType, "")
	}

	if fileSize > r.config.MaxFileSize {
		return "", WrapFileTooLarge(contentType, fileSize, "", r.config.MaxFileSize)
	}

	// Normalize content type (remove parameters like charset)
	contentType = normalizeContentType(contentType)

	// Log extraction start
	r.logger.LogExtractionStart(contentType, fileSize)
	startTime := time.Now()

	// Find extractor
	extractor, exists := r.extractors[contentType]
	if !exists {
		err := WrapUnsupportedFormat(contentType, fileSize, r.SupportedFormats())
		r.logger.LogExtractionFailure(contentType, fileSize, time.Since(startTime), err)
		return "", err
	}

	// Calculate timeout based on file size
	timeout := r.calculateTimeout(fileSize)

	// Create timeout context
	extractCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute extraction with queue management for concurrency control
	text, err := r.queue.Execute(extractCtx, func() (string, error) {
		// Extract text with memory monitoring
		return extractWithMemoryLimit(extractCtx, r.config.MaxMemoryPerFile, func() (string, error) {
			return extractor.Extract(extractCtx, data)
		})
	})

	duration := time.Since(startTime)

	// Record extraction event
	event := ExtractionEvent{
		Timestamp:   startTime,
		ContentType: contentType,
		FileSize:    fileSize,
		Duration:    duration,
		Success:     err == nil,
		TextLength:  len(text),
	}
	if err != nil {
		event.Error = err.Error()
	}
	r.stats.RecordExtraction(event)

	// Log result
	if err != nil {
		if extractCtx.Err() == context.DeadlineExceeded {
			r.logger.LogExtractionTimeout(contentType, fileSize, timeout)
			wrappedErr := WrapExtractionTimeout(contentType, fileSize, "", timeout.String())
			return "", wrappedErr
		}
		r.logger.LogExtractionFailure(contentType, fileSize, duration, err)

		// Wrap the error with context
		wrappedErr := WrapGenericExtractionError(contentType, fileSize, "", err)
		return "", wrappedErr
	}

	r.logger.LogExtractionSuccess(contentType, fileSize, duration, len(text))
	return text, nil
}

// ExtractWithValidation extracts text with format validation
func (r *ExtractionRouter) ExtractWithValidation(ctx context.Context, data []byte, contentType, filename string) (string, error) {
	// Validate file size
	fileSize := int64(len(data))
	if fileSize == 0 {
		return "", WrapEmptyFile(contentType, filename)
	}

	if fileSize > r.config.MaxFileSize {
		return "", WrapFileTooLarge(contentType, fileSize, filename, r.config.MaxFileSize)
	}

	// Normalize content type (remove parameters like charset)
	contentType = normalizeContentType(contentType)

	// Log extraction start
	r.logger.LogExtractionStart(contentType, fileSize)
	startTime := time.Now()

	// Validate format if filename is provided
	if filename != "" {
		if err := ValidateFormat(data, filename, contentType); err != nil {
			duration := time.Since(startTime)
			r.logger.LogExtractionFailure(contentType, fileSize, duration, err)
			wrappedErr := WrapInvalidFormat(contentType, fileSize, filename, err.Error())
			return "", wrappedErr
		}
	}

	// Find extractor
	extractor, exists := r.extractors[contentType]
	if !exists {
		err := WrapUnsupportedFormat(contentType, fileSize, r.SupportedFormats())
		r.logger.LogExtractionFailure(contentType, fileSize, time.Since(startTime), err)
		return "", err
	}

	// Calculate timeout based on file size
	timeout := r.calculateTimeout(fileSize)

	// Create timeout context
	extractCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute extraction with queue management for concurrency control
	text, err := r.queue.Execute(extractCtx, func() (string, error) {
		// Extract text with memory monitoring
		return extractWithMemoryLimit(extractCtx, r.config.MaxMemoryPerFile, func() (string, error) {
			return extractor.Extract(extractCtx, data)
		})
	})

	duration := time.Since(startTime)

	// Record extraction event
	event := ExtractionEvent{
		Timestamp:   startTime,
		ContentType: contentType,
		FileSize:    fileSize,
		Duration:    duration,
		Success:     err == nil,
		TextLength:  len(text),
	}
	if err != nil {
		event.Error = err.Error()
	}
	r.stats.RecordExtraction(event)

	// Log result
	if err != nil {
		if extractCtx.Err() == context.DeadlineExceeded {
			r.logger.LogExtractionTimeout(contentType, fileSize, timeout)
			wrappedErr := WrapExtractionTimeout(contentType, fileSize, filename, timeout.String())
			return "", wrappedErr
		}
		r.logger.LogExtractionFailure(contentType, fileSize, duration, err)

		// Wrap the error with context
		wrappedErr := WrapGenericExtractionError(contentType, fileSize, filename, err)
		return "", wrappedErr
	}

	r.logger.LogExtractionSuccess(contentType, fileSize, duration, len(text))
	return text, nil
}

// IsSupported checks if a content type is supported
func (r *ExtractionRouter) IsSupported(contentType string) bool {
	contentType = normalizeContentType(contentType)
	_, exists := r.extractors[contentType]
	return exists
}

// SupportedFormats returns a list of supported content types
func (r *ExtractionRouter) SupportedFormats() []string {
	formats := make([]string, 0, len(r.formats))
	for _, info := range r.formats {
		formats = append(formats, info.Name)
	}
	return formats
}

// GetFormatInfo returns information about a supported format
func (r *ExtractionRouter) GetFormatInfo(contentType string) (FormatInfo, bool) {
	contentType = normalizeContentType(contentType)
	info, exists := r.formats[contentType]
	return info, exists
}

// GetMetrics returns current extraction metrics
func (r *ExtractionRouter) GetMetrics() ExtractionMetrics {
	metrics := r.queue.GetMetrics()
	r.logger.LogExtractionMetrics(metrics)
	return metrics
}

// GetStats returns extraction statistics
func (r *ExtractionRouter) GetStats() ExtractionStats {
	return r.stats.GetStats()
}

// SetLoggingEnabled enables or disables logging
func (r *ExtractionRouter) SetLoggingEnabled(enabled bool) {
	r.logger.enabled = enabled
}

// calculateTimeout determines the extraction timeout based on file size
// Requirements: 5.1 - Set 5-second timeout for files under 10MB, scale for larger files
func (r *ExtractionRouter) calculateTimeout(fileSize int64) time.Duration {
	const (
		tenMB           = 10 * 1024 * 1024
		baseTimeout     = 5 * time.Second
		additionalPerMB = 500 * time.Millisecond
	)

	// For files under 10MB, use base timeout of 5 seconds
	if fileSize <= tenMB {
		return baseTimeout
	}

	// For larger files, add 500ms per MB above 10MB
	additionalMB := (fileSize - tenMB) / (1024 * 1024)
	scaledTimeout := baseTimeout + time.Duration(additionalMB)*additionalPerMB

	// Cap at the configured maximum timeout
	if scaledTimeout > r.config.ExtractionTimeout {
		return r.config.ExtractionTimeout
	}

	return scaledTimeout
}

// normalizeContentType removes parameters from content type
func normalizeContentType(contentType string) string {
	// Remove charset and other parameters
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = contentType[:idx]
	}
	return strings.TrimSpace(strings.ToLower(contentType))
}
