package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bipulkrdas/orgmind/backend/internal/extraction"
	"github.com/bipulkrdas/orgmind/backend/internal/models"
	"github.com/bipulkrdas/orgmind/backend/internal/repository"
	"github.com/bipulkrdas/orgmind/backend/internal/storage"
	"github.com/google/uuid"
)

const (
	MaxFileSize = 50 * 1024 * 1024 // 50MB in bytes
)

// documentService implements DocumentService interface
type documentService struct {
	documentRepo      repository.DocumentRepository
	storageService    storage.StorageService
	processingService ProcessingService
	graphService      GraphService
	extractionService extraction.ExtractionService
}

// NewDocumentService creates a new instance of DocumentService
func NewDocumentService(
	documentRepo repository.DocumentRepository,
	storageService storage.StorageService,
	processingService ProcessingService,
	graphService GraphService,
	extractionService extraction.ExtractionService,
) DocumentService {
	return &documentService{
		documentRepo:      documentRepo,
		storageService:    storageService,
		processingService: processingService,
		graphService:      graphService,
		extractionService: extractionService,
	}
}

// CreateFromEditor handles text content from the editor with Lexical state
func (s *documentService) CreateFromEditor(ctx context.Context, userID, graphID, plainText, lexicalState string) (*models.Document, error) {
	// Validate content
	if plainText == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}
	if lexicalState == "" {
		return nil, fmt.Errorf("lexical state cannot be empty")
	}

	// Verify graph membership before creating document
	gr, err := s.graphService.GetByID(ctx, graphID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify graph membership: %w", err)
	}

	// Generate unique document ID
	documentID := uuid.New().String()

	// Create combined JSON structure for storage
	combinedContent := map[string]interface{}{
		"plainText":    plainText,
		"lexicalState": lexicalState,
		"metadata": map[string]interface{}{
			"version":   "1.0",
			"createdAt": time.Now().UTC().Format(time.RFC3339),
		},
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(combinedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal content: %w", err)
	}

	// Create document metadata
	now := time.Now().UTC()
	contentType := "application/json"
	sizeBytes := int64(len(jsonBytes))

	doc := &models.Document{
		ID:          documentID,
		UserID:      userID,
		GraphID:     &graphID,
		Filename:    nil, // No filename for editor content
		ContentType: &contentType,
		StorageKey:  "", // Will be set after upload
		SizeBytes:   sizeBytes,
		Source:      "editor",
		Status:      "processing",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Upload combined JSON to S3
	storageKey, err := s.storageService.Upload(
		ctx,
		userID,
		documentID,
		"editor-content.json",
		bytes.NewReader(jsonBytes),
		contentType,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upload content to storage: %w", err)
	}

	doc.StorageKey = storageKey

	// Store document metadata in database
	err = s.documentRepo.Create(ctx, doc)
	if err != nil {
		// Attempt to clean up uploaded file
		_ = s.storageService.Delete(ctx, storageKey)
		return nil, fmt.Errorf("failed to create document in database: %w", err)
	}

	// Increment document count for the graph
	if err := s.graphService.IncrementDocumentCount(ctx, graphID); err != nil {
		// Log error but don't fail the document creation
		fmt.Printf("Warning: failed to increment document count for graph %s: %v\n", graphID, err)
	}

	// Process document asynchronously using plain text for Zep
	go func() {
		// Use a new context for background processing
		bgCtx := context.Background()
		if err := s.processingService.ProcessDocument(bgCtx, userID, gr.ZepGraphID, documentID, plainText); err != nil {
			// Log error (in production, use proper logging)
			fmt.Printf("Error processing document %s: %v\n", documentID, err)
		}
	}()

	return doc, nil
}

// CreateFromFile handles multipart file uploads
func (s *documentService) CreateFromFile(ctx context.Context, userID, graphID string, file []byte, filename, contentType string) (*models.Document, error) {
	// Validate file size
	if len(file) > MaxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of 50MB")
	}

	if len(file) == 0 {
		return nil, fmt.Errorf("file cannot be empty")
	}

	// Validate file type (whitelist of supported types)
	if !s.isValidFileType(contentType) {
		return nil, fmt.Errorf("unsupported file type: %s. Supported formats: %v", contentType, s.extractionService.SupportedFormats())
	}

	// Verify graph membership before creating document
	gr, err := s.graphService.GetByID(ctx, graphID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify graph membership: %w", err)
	}

	// Generate unique document ID
	documentID := uuid.New().String()

	// Create document metadata
	now := time.Now().UTC()
	sizeBytes := int64(len(file))

	doc := &models.Document{
		ID:          documentID,
		UserID:      userID,
		GraphID:     &graphID,
		Filename:    &filename,
		ContentType: &contentType,
		StorageKey:  "", // Will be set after upload
		SizeBytes:   sizeBytes,
		Source:      "upload",
		Status:      "processing",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Upload file to S3
	storageKey, err := s.storageService.Upload(
		ctx,
		userID,
		documentID,
		filename,
		bytes.NewReader(file),
		contentType,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to storage: %w", err)
	}

	doc.StorageKey = storageKey

	// Store document metadata in database
	err = s.documentRepo.Create(ctx, doc)
	if err != nil {
		// Attempt to clean up uploaded file
		_ = s.storageService.Delete(ctx, storageKey)
		return nil, fmt.Errorf("failed to create document in database: %w", err)
	}

	// Increment document count for the graph
	if err = s.graphService.IncrementDocumentCount(ctx, graphID); err != nil {
		// Log error but don't fail the document creation
		fmt.Printf("Warning: failed to increment document count for graph %s: %v\n", graphID, err)
	}

	// Extract text content from file using extraction service
	textContent, err := s.extractionService.Extract(ctx, file, contentType)
	if err != nil {
		// Get user-friendly error message
		userMessage := extraction.GetUserFriendlyMessage(err)

		// Update document status to failed with error message
		doc.Status = "failed"
		doc.ErrorMessage = &userMessage
		doc.UpdatedAt = time.Now().UTC()
		_ = s.documentRepo.Update(ctx, doc)

		// Return user-friendly error
		return nil, fmt.Errorf("%s", userMessage)
	}

	// Process document asynchronously (in production, this would be a background job)
	go func() {
		// Use a new context for background processing
		bgCtx := context.Background()
		if err := s.processingService.ProcessDocument(bgCtx, userID, gr.ZepGraphID, documentID, textContent); err != nil {
			// Log error (in production, use proper logging)
			fmt.Printf("Error processing document %s: %v\n", documentID, err)
		}
	}()

	return doc, nil
}

// GetDocument retrieves a document by ID, ensuring the user owns it
func (s *documentService) GetDocument(ctx context.Context, documentID, userID string) (*models.Document, error) {
	doc, err := s.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	// Verify ownership
	if doc.UserID != userID {
		return nil, fmt.Errorf("access denied: document does not belong to user")
	}

	return doc, nil
}

// ListUserDocuments retrieves all documents for a specific user
func (s *documentService) ListUserDocuments(ctx context.Context, userID string) ([]*models.Document, error) {
	docs, err := s.documentRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user documents: %w", err)
	}

	return docs, nil
}

// ListGraphDocuments retrieves all documents for a specific graph
func (s *documentService) ListGraphDocuments(ctx context.Context, graphID string) ([]*models.Document, error) {
	docs, err := s.documentRepo.ListByGraphID(ctx, graphID)
	if err != nil {
		return nil, fmt.Errorf("failed to list graph documents: %w", err)
	}

	return docs, nil
}

// UpdateDocument updates document content and re-processes it
func (s *documentService) UpdateDocument(ctx context.Context, documentID, userID, plainText, lexicalState string) (*models.Document, error) {
	// Validate content
	if plainText == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}
	if lexicalState == "" {
		return nil, fmt.Errorf("lexical state cannot be empty")
	}

	// Get the document
	doc, err := s.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	// Verify user is member of document's graph
	if doc.GraphID == nil {
		return nil, fmt.Errorf("document is not associated with a graph")
	}

	gr, err := s.graphService.GetByID(ctx, *doc.GraphID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify graph membership: %w", err)
	}

	// Create combined JSON structure for storage
	combinedContent := map[string]interface{}{
		"plainText":    plainText,
		"lexicalState": lexicalState,
		"metadata": map[string]interface{}{
			"version":   "1.0",
			"updatedAt": time.Now().UTC().Format(time.RFC3339),
		},
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(combinedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal content: %w", err)
	}

	// Update document metadata
	now := time.Now().UTC()
	contentType := "application/json"
	sizeBytes := int64(len(jsonBytes))

	doc.ContentType = &contentType
	doc.SizeBytes = sizeBytes
	doc.Status = "processing"
	doc.UpdatedAt = now

	// Upload new content to storage
	storageKey, err := s.storageService.Upload(
		ctx,
		userID,
		documentID,
		"editor-content.json",
		bytes.NewReader(jsonBytes),
		contentType,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upload content to storage: %w", err)
	}

	// Delete old storage file if different
	if doc.StorageKey != storageKey && doc.StorageKey != "" {
		_ = s.storageService.Delete(ctx, doc.StorageKey)
	}

	doc.StorageKey = storageKey

	// Update document in database
	err = s.documentRepo.Update(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to update document in database: %w", err)
	}

	// Re-process document asynchronously using plain text for Zep
	go func() {
		bgCtx := context.Background()
		if err := s.processingService.ProcessDocument(bgCtx, userID, gr.ZepGraphID, documentID, plainText); err != nil {
			fmt.Printf("Error processing document %s: %v\n", documentID, err)
		}
	}()

	return doc, nil
}

// GetDocumentContent retrieves the actual content of a document from storage
func (s *documentService) GetDocumentContent(ctx context.Context, documentID, userID string) (map[string]interface{}, error) {
	// Get the document
	doc, err := s.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	// Verify user is member of document's graph
	if doc.GraphID == nil {
		return nil, fmt.Errorf("document is not associated with a graph")
	}

	_, err = s.graphService.GetByID(ctx, *doc.GraphID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify graph membership: %w", err)
	}

	// Download content from storage
	reader, err := s.storageService.Download(ctx, doc.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to download content from storage: %w", err)
	}
	defer reader.Close()

	// Read content
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(reader); err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	// Handle different content types
	if doc.ContentType != nil && *doc.ContentType == "application/json" {
		// Parse JSON content (editor documents with Lexical state)
		var content map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &content); err != nil {
			return nil, fmt.Errorf("failed to parse JSON content: %w", err)
		}
		return content, nil
	} else {
		// Legacy plain text content
		return map[string]interface{}{
			"plainText": buf.String(),
			"metadata": map[string]interface{}{
				"version": "legacy",
				"type":    "plain-text",
			},
		}, nil
	}
}

// DeleteDocument deletes a document and its associated data
func (s *documentService) DeleteDocument(ctx context.Context, documentID, userID string) error {
	// Get the document
	doc, err := s.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	// Verify user is member of document's graph
	if doc.GraphID == nil {
		return fmt.Errorf("document is not associated with a graph")
	}

	_, err = s.graphService.GetByID(ctx, *doc.GraphID, userID)
	if err != nil {
		return fmt.Errorf("failed to verify graph membership: %w", err)
	}

	// Delete from storage
	if doc.StorageKey != "" {
		if err := s.storageService.Delete(ctx, doc.StorageKey); err != nil {
			// Log error but continue with database deletion
			fmt.Printf("Warning: failed to delete storage file %s: %v\n", doc.StorageKey, err)
		}
	}

	// Delete from database
	if err := s.documentRepo.Delete(ctx, documentID); err != nil {
		return fmt.Errorf("failed to delete document from database: %w", err)
	}

	// Decrement document count for the graph
	if err := s.graphService.DecrementDocumentCount(ctx, *doc.GraphID); err != nil {
		// Log error but don't fail the deletion
		fmt.Printf("Warning: failed to decrement document count for graph %s: %v\n", *doc.GraphID, err)
	}

	return nil
}

// isValidFileType checks if the content type is supported by the extraction service
func (s *documentService) isValidFileType(contentType string) bool {
	return s.extractionService.IsSupported(contentType)
}
