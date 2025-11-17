# Implementation Plan

- [x] 1. Set up text extraction infrastructure
  - Create `backend/internal/extraction` package structure
  - Define TextExtractor interface
  - Implement extraction router with format detection
  - _Requirements: 7.4_

- [x] 2. Implement basic text format extractors
  - [x] 2.1 Implement plain text extractor
    - Handle UTF-8 and other encodings
    - Preserve line breaks and basic formatting
    - _Requirements: 3.1_

  - [x] 2.2 Implement Markdown extractor
    - Extract text while preserving structure
    - Handle code blocks appropriately
    - _Requirements: 3.2_

  - [x] 2.3 Implement HTML extractor using golang.org/x/net/html
    - Parse HTML and extract text nodes
    - Remove script and style tags
    - Handle special characters and entities
    - _Requirements: 3.3_

  - [x] 2.4 Implement JSON extractor
    - Extract string values from JSON
    - Format nested structures readably
    - _Requirements: 3.4_

  - [x] 2.5 Implement CSV extractor using encoding/csv
    - Parse CSV and extract all cell values
    - Handle different delimiters
    - _Requirements: 3.5_

- [x] 3. Implement PDF text extraction
  - [x] 3.1 Add github.com/ledongthuc/pdf dependency
    - Update go.mod with PDF library
    - Verify license compatibility (MIT)
    - _Requirements: 7.1, 7.2_

  - [x] 3.2 Implement PDF extractor
    - Extract text from all pages sequentially
    - Handle multi-page documents
    - Preserve paragraph breaks
    - _Requirements: 1.1, 1.2, 1.5_

  - [x] 3.3 Add PDF error handling
    - Handle corrupted PDF files gracefully
    - Return descriptive error messages
    - Implement extraction timeout
    - _Requirements: 1.4, 5.3_

- [x] 4. Implement Microsoft Office document extraction
  - [x] 4.1 Add Office document dependencies
    - Add github.com/nguyenthenguyen/docx for Word documents
    - Add github.com/xuri/excelize/v2 for Excel documents
    - Verify licenses (MIT)
    - _Requirements: 7.1, 7.2_

  - [x] 4.2 Implement .docx extractor
    - Extract body text, headers, and footers
    - Handle tables and lists
    - Preserve basic document structure
    - _Requirements: 2.1, 2.4_

  - [x] 4.3 Implement .xlsx extractor
    - Extract text from all sheets
    - Format cell values with appropriate delimiters
    - Handle formulas by extracting calculated values
    - _Requirements: 2.3, 2.4_

  - [x] 4.4 Implement .pptx extractor
    - Extract text from all slides
    - Include slide notes if present
    - Maintain slide order
    - _Requirements: 4.3_

  - [x] 4.5 Add Office document error handling
    - Handle password-protected documents
    - Return appropriate error messages
    - Handle corrupted Office files
    - _Requirements: 2.5, 6.2_

- [x] 5. Implement additional format support
  - [x] 5.1 Add EPUB support
    - Add github.com/bmaupin/go-epub dependency
    - Extract text from all chapters
    - Maintain chapter order
    - _Requirements: 4.2_

  - [x] 5.2 Implement RTF extractor
    - Parse RTF format and extract plain text
    - Handle basic RTF formatting codes
    - _Requirements: 4.1_

  - [x] 5.3 Add format validation
    - Validate file extension matches content type
    - Check file headers for format verification
    - _Requirements: 4.5_

- [x] 6. Integrate extraction service with document service
  - [x] 6.1 Update document_service.go
    - Replace extractTextContent function with new extraction service
    - Add extraction service dependency injection
    - Update error handling to use new error types
    - _Requirements: 6.1, 6.2_

  - [x] 6.2 Update isValidFileType function
    - Add new supported formats to whitelist
    - Include modern document formats
    - _Requirements: 4.4, 4.5_

  - [x] 6.3 Update file upload handler
    - Add better error messages for unsupported formats
    - Return list of supported formats on error
    - _Requirements: 6.3, 6.4_

- [x] 7. Implement performance and reliability features
  - [x] 7.1 Add extraction timeouts
    - Implement context-based timeouts for extraction
    - Set 5-second timeout for files under 10MB
    - Scale timeout based on file size
    - _Requirements: 5.1, 5.4_

  - [x] 7.2 Add memory management
    - Implement memory limits for extraction
    - Use streaming where possible for large files
    - _Requirements: 5.4_

  - [x] 7.3 Implement concurrent extraction safety
    - Ensure thread-safe extraction operations
    - Add extraction queue if needed
    - _Requirements: 5.5_

  - [x] 7.4 Add comprehensive logging
    - Log extraction attempts with format and size
    - Log extraction failures with detailed errors
    - Track extraction performance metrics
    - _Requirements: 5.2, 7.5_

- [x] 8. Error handling and user feedback
  - [x] 8.1 Define extraction error types
    - Create specific error types for each failure mode
    - Implement error wrapping for context
    - _Requirements: 6.2, 6.5_

  - [x] 8.2 Update document status handling
    - Set status to "failed" on extraction errors
    - Store error message in database
    - _Requirements: 6.1_

  - [x] 8.3 Implement user-friendly error messages
    - Map technical errors to user-friendly messages
    - Include actionable guidance in error messages
    - _Requirements: 6.2, 6.3_

- [ ]* 9. Testing and validation
  - [ ]* 9.1 Create test document suite
    - Collect sample documents for each supported format
    - Include edge cases (corrupted, empty, large files)
    - Add password-protected documents for error testing

  - [ ]* 9.2 Write unit tests for each extractor
    - Test successful extraction for each format
    - Test error handling for corrupted files
    - Test edge cases and boundary conditions

  - [ ]* 9.3 Write integration tests
    - Test end-to-end document upload and extraction
    - Test concurrent extraction scenarios
    - Test timeout and memory limit enforcement

  - [ ]* 9.4 Performance testing
    - Benchmark extraction speed for various file sizes
    - Test memory usage during extraction
    - Verify timeout handling works correctly

- [ ]* 10. Documentation and deployment
  - [ ]* 10.1 Update API documentation
    - Document supported file formats
    - Document error responses
    - Add examples for each format

  - [ ]* 10.2 Update user documentation
    - List all supported file formats
    - Explain file size limits
    - Provide troubleshooting guide

  - [ ]* 10.3 Update deployment configuration
    - Add new dependencies to Dockerfile
    - Update go.mod and go.sum
    - Verify no CGO dependencies for easy deployment

  - [ ]* 10.4 Create monitoring dashboard
    - Track extraction success rates by format
    - Monitor extraction performance
    - Alert on high failure rates
