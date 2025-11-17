package service

import (
	"context"
	"fmt"
	"time"

	"github.com/bipulkrdas/orgmind/backend/internal/repository"
	"github.com/bipulkrdas/orgmind/backend/pkg/utils"
)

// processingService implements ProcessingService interface
type processingService struct {
	documentRepo repository.DocumentRepository
	zepService   ZepService
}

// NewProcessingService creates a new instance of ProcessingService
func NewProcessingService(documentRepo repository.DocumentRepository, zepService ZepService) ProcessingService {
	return &processingService{
		documentRepo: documentRepo,
		zepService:   zepService,
	}
}

// ProcessDocument orchestrates the document processing workflow:
// 1. Clean the text content
// 2. Chunk the document into manageable pieces
// 3. Send chunks to Zep for knowledge graph creation
// 4. Update document status in database
func (s *processingService) ProcessDocument(ctx context.Context, userID, graphID, documentID, content string) error {
	// Step 1: Clean the text content
	cleanedContent := utils.CleanText(content)
	if cleanedContent == "" {
		return fmt.Errorf("content is empty after cleaning")
	}

	// Step 2: Chunk the document
	chunks := utils.ChunkDocument(cleanedContent)
	if len(chunks) == 0 {
		return fmt.Errorf("no chunks created from document")
	}

	// Step 3: Send chunks to Zep for memory creation
	metadata := map[string]any{
		"documentId": documentID,
		"userId":     userID,
		"chunkCount": len(chunks),
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}

	err := s.zepService.AddMemory(ctx, graphID, chunks, metadata)
	if err != nil {
		// Update document status to failed
		if updateErr := s.updateDocumentStatus(ctx, documentID, "failed"); updateErr != nil {
			return fmt.Errorf("failed to add memory to Zep: %w, and failed to update document status: %v", err, updateErr)
		}
		return fmt.Errorf("failed to add memory to Zep: %w", err)
	}

	// Step 4: Update document status to completed
	err = s.updateDocumentStatus(ctx, documentID, "completed")
	if err != nil {
		return fmt.Errorf("failed to update document status: %w", err)
	}

	return nil
}

// updateDocumentStatus updates the status of a document in the database
func (s *processingService) updateDocumentStatus(ctx context.Context, documentID, status string) error {
	doc, err := s.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	doc.Status = status
	doc.UpdatedAt = time.Now().UTC()

	err = s.documentRepo.Update(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	return nil
}
