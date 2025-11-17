# Design Document

## Overview

This document outlines the design for implementing comprehensive document text extraction in OrgMind. The solution uses well-established Go libraries to extract text from various document formats, with a focus on reliability, performance, and maintainability.

## Architecture

### High-Level Design

```
┌─────────────────┐
│  File Upload    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Document Service│
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────┐
│  Text Extraction Service        │
│  ┌───────────────────────────┐  │
│  │  Format Detector          │  │
│  └───────────┬───────────────┘  │
│              │                   │
│              ▼                   │
│  ┌───────────────────────────┐  │
│  │  Extraction Router        │  │
│  └───────────┬───────────────┘  │
│              │                   │
│      ┌───────┴───────┐          │
│      ▼               ▼          │
│  ┌────────┐    ┌────────┐      │
│  │  PDF   │    │ Office │      │
│  │Extract │    │Extract │      │
│  └────────┘    └────────┘      │
│      ▼               ▼          │
│  ┌────────┐    ┌────────┐      │
│  │  Text  │    │  EPUB  │      │
│  │Extract │    │Extract │      │
│  └────────┘    └────────┘      │
└─────────────────────────────────┘
         │
         ▼
┌─────────────────┐
│  Zep Processing │
└─────────────────┘
```

## Library Research and Selection

### PDF Extraction

**Selected Library: `github.com/ledongthuc/pdf`**
- **Pros**: Pure Go, no CGO dependencies, actively maintained, MIT license
- **Cons**: Limited support for complex PDFs with images
- **Alternative**: `github.com/pdfcpu/pdfcpu` for more advanced features

**Fallback Library: `github.com/unidoc/unipdf` (Commercial)**
- For enterprise features like OCR and advanced PDF parsing
- Requires license for production use

### Microsoft Office Documents

**Selected Library: `github.com/unidoc/unioffice`**
- **Pros**: Pure Go, supports .docx, .xlsx, .pptx, actively maintained
- **Cons**: Commercial license required for production (AGPL for open source)
- **Use Case**: Best for Office 2007+ formats (.docx, .xlsx, .pptx)

**Alternative: `github.com/nguyenthenguyen/docx`**
- **Pros**: MIT license, good for .docx files
- **Cons**: Limited to Word documents only

**For Legacy Formats (.doc, .xls):**
- Use `github.com/richardlehane/mscfb` for parsing OLE2 format
- Or recommend users convert to modern formats

### HTML Parsing

**Selected Library: `golang.org/x/net/html`**
- **Pros**: Official Go package, well-maintained, BSD license
- **Cons**: None significant
- **Use Case**: Extract text from HTML while removing tags

### RTF Documents

**Selected Library: `github.com/EndFirstCorp/peekingReader` + custom parser**
- **Pros**: Lightweight, can build custom RTF parser
- **Alternative**: Convert RTF to plain text using regex patterns

### EPUB Documents

**Selected Library: `github.com/bmaupin/go-epub`**
- **Pros**: Pure Go, MIT license, good EPUB support
- **Cons**: Primarily for creating EPUBs, but can read them
- **Alternative**: `github.com/taylorskalyo/goreader` for reading

### PowerPoint

**Covered by**: `github.com/unidoc/unioffice`
- Extracts text from .pptx slides

### CSV Processing

**Selected Library: `encoding/csv` (standard library)**
- **Pros**: Built-in, no dependencies, well-tested
- **Cons**: None
- **Use Case**: Parse CSV and extract all cell values

## Components and Interfaces

### Text Extraction Service Interface

```go
type TextExtractor interface {
    // Extract text from document bytes
    Extract(data []byte, contentType string) (string, error)
    
    // Check if format is supported
    IsSupported(contentType string) bool
    
    // Get list of supported formats
    SupportedFormats() []string
}
```

### Format-Specific Extractors

```go
type PDFExtractor struct{}
func (e *PDFExtractor) Extract(data []byte) (string, error)

type DocxExtractor struct{}
func (e *DocxExtractor) Extract(data []byte) (string, error)

type XlsxExtractor struct{}
func (e *XlsxExtractor) Extract(data []byte) (string, error)

type HTMLExtractor struct{}
func (e *HTMLExtractor) Extract(data []byte) (string, error)

type EPUBExtractor struct{}
func (e *EPUBExtractor) Extract(data []byte) (string, error)
```

### Extraction Router

```go
type ExtractionRouter struct {
    extractors map[string]Extractor
}

func (r *ExtractionRouter) Route(data []byte, contentType string) (string, error) {
    extractor, exists := r.extractors[contentType]
    if !exists {
        return "", ErrUnsupportedFormat
    }
    return extractor.Extract(data)
}
```

## Data Models

### Extraction Result

```go
type ExtractionResult struct {
    Text      string
    PageCount int  // For PDFs
    WordCount int
    Metadata  map[string]string
    Error     error
}
```

### Supported Format Registry

```go
var SupportedFormats = map[string]FormatInfo{
    "application/pdf": {
        Name: "PDF Document",
        Extensions: []string{".pdf"},
        Extractor: "PDFExtractor",
    },
    "application/vnd.openxmlformats-officedocument.wordprocessingml.document": {
        Name: "Word Document",
        Extensions: []string{".docx"},
        Extractor: "DocxExtractor",
    },
    // ... more formats
}
```

## Error Handling

### Error Types

```go
var (
    ErrUnsupportedFormat = errors.New("unsupported document format")
    ErrCorruptedFile     = errors.New("file appears to be corrupted")
    ErrPasswordProtected = errors.New("document is password protected")
    ErrExtractionFailed  = errors.New("text extraction failed")
    ErrFileTooLarge      = errors.New("file exceeds maximum size")
)
```

### Error Recovery Strategy

1. **Primary Extraction Attempt**: Use format-specific library
2. **Fallback Strategy**: Try alternative library or method
3. **Graceful Degradation**: Return partial content if available
4. **User Notification**: Update document status with clear error message

## Testing Strategy

### Unit Tests

- Test each extractor with sample documents
- Test error handling for corrupted files
- Test edge cases (empty files, very large files)
- Mock file uploads for integration tests

### Test Documents

Create test suite with:
- Valid documents of each supported format
- Corrupted documents
- Password-protected documents
- Documents with special characters
- Multi-page/multi-sheet documents

### Performance Tests

- Benchmark extraction speed for various file sizes
- Test memory usage during concurrent extractions
- Test timeout handling for large documents

## Implementation Phases

### Phase 1: Core Text Formats (Week 1)
- Plain text, Markdown, HTML, JSON, CSV
- Basic error handling
- Unit tests

### Phase 2: PDF Support (Week 2)
- PDF text extraction using ledongthuc/pdf
- Multi-page handling
- Error handling for corrupted PDFs

### Phase 3: Office Documents (Week 3)
- .docx extraction
- .xlsx extraction
- .pptx extraction
- License compliance check

### Phase 4: Additional Formats (Week 4)
- EPUB support
- RTF support
- Enhanced error messages
- Performance optimization

## Security Considerations

1. **File Size Limits**: Enforce 50MB maximum to prevent DoS
2. **Memory Limits**: Set extraction timeout and memory caps
3. **Malicious Files**: Validate file headers match content types
4. **Dependency Security**: Regular security audits of libraries
5. **Sandboxing**: Consider running extraction in isolated environment

## Performance Optimization

1. **Streaming**: Process large files in chunks where possible
2. **Caching**: Cache extraction results for duplicate files
3. **Concurrency**: Use goroutines for parallel extraction
4. **Resource Limits**: Implement extraction timeouts
5. **Lazy Loading**: Only extract when needed for processing

## Monitoring and Logging

### Metrics to Track

- Extraction success/failure rates by format
- Average extraction time by format
- Memory usage during extraction
- Error types and frequencies

### Logging Strategy

```go
log.Info("Starting extraction", 
    "format", contentType,
    "size", len(data),
    "documentID", docID)

log.Error("Extraction failed",
    "format", contentType,
    "error", err,
    "documentID", docID)
```

## Deployment Considerations

### Dependencies

Add to `go.mod`:
```
github.com/ledongthuc/pdf v0.0.0-20220302134840-0c2507a12d80
github.com/nguyenthenguyen/docx v0.0.0-20220721043308-1903da0ef37d
github.com/xuri/excelize/v2 v2.8.0
github.com/bmaupin/go-epub v1.1.0
golang.org/x/net v0.19.0
```

### Docker Image

Ensure base image includes necessary dependencies (none required for pure Go libraries)

### Configuration

```go
type ExtractionConfig struct {
    MaxFileSize      int64
    ExtractionTimeout time.Duration
    EnableOCR        bool
    MaxConcurrent    int
}
```

## Future Enhancements

1. **OCR Support**: Add image text extraction using tesseract
2. **Language Detection**: Identify document language
3. **Metadata Extraction**: Extract author, creation date, etc.
4. **Format Conversion**: Convert between formats
5. **Preview Generation**: Create text previews for UI
6. **Batch Processing**: Process multiple documents efficiently
