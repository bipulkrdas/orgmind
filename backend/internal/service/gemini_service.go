package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/bipulkrdas/orgmind/backend/internal/models"
	"github.com/bipulkrdas/orgmind/backend/internal/repository"
	"google.golang.org/genai"
)

// Custom error types for Gemini operations
var (
	ErrGeminiStoreCreation = errors.New("failed to create File Search store")
	ErrGeminiStoreNotFound = errors.New("File Search store not found")
	ErrGeminiUploadFailed  = errors.New("failed to upload document to File Search")
	ErrGeminiQueryFailed   = errors.New("failed to query Gemini API")
	ErrGeminiAPIKey        = errors.New("Gemini API key not configured")
)

// geminiService implements the GeminiService interface
type geminiService struct {
	client          *genai.Client
	graphRepo       repository.GraphRepository
	docRepo         repository.DocumentRepository
	geminiStoreRepo repository.GeminiStoreRepository
	storeID         string // Shared File Search store ID
	storeName       string // Store display name
	apiKey          string
	projectID       string
	location        string
}

// NewGeminiService creates a new Gemini service instance
func NewGeminiService(
	apiKey, projectID, location, storeID, storeName string,
	graphRepo repository.GraphRepository,
	docRepo repository.DocumentRepository,
	geminiStoreRepo repository.GeminiStoreRepository,
) (GeminiService, error) {
	if apiKey == "" {
		return nil, ErrGeminiAPIKey
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &geminiService{
		client:          client,
		graphRepo:       graphRepo,
		docRepo:         docRepo,
		geminiStoreRepo: geminiStoreRepo,
		storeID:         storeID,
		storeName:       storeName,
		apiKey:          apiKey,
		projectID:       projectID,
		location:        location,
	}, nil
}

// InitializeStore creates or retrieves the shared File Search store
func (s *geminiService) InitializeStore(ctx context.Context, storeName string) (storeID string, err error) {
	// Log store initialization attempt
	log.Printf("[Gemini] Store Initialization: Checking database for existing store with name '%s'", storeName)

	// Check if store already exists in database
	existingStore, err := s.geminiStoreRepo.GetByStoreName(ctx, storeName)
	if err != nil {
		log.Printf("[Gemini] Store Initialization: ERROR - Failed to query database: %v", err)
		return "", fmt.Errorf("failed to query database for store: %w", err)
	}

	// If store exists in database, use it
	if existingStore != nil {
		log.Printf("[Gemini] Store Initialization: FOUND - Using existing store '%s' with ID: %s", storeName, existingStore.StoreID)
		s.storeID = existingStore.StoreID
		return existingStore.StoreID, nil
	}

	// Store not found in database, create new one in Gemini
	log.Printf("[Gemini] Store Initialization: NOT FOUND - Creating new File Search store with name '%s'", storeName)

	// Create File Search store with retry logic
	var store *genai.FileSearchStore

	for attempt := 1; attempt <= 3; attempt++ {
		log.Printf("[Gemini] Store Initialization: Attempt %d/3 to create store '%s'", attempt, storeName)

		store, err = s.client.FileSearchStores.Create(ctx, &genai.CreateFileSearchStoreConfig{
			DisplayName: storeName,
		})

		if err == nil {
			// Log successful creation with store ID
			storeID = store.Name
			log.Printf("[Gemini] Store Initialization: SUCCESS - Created File Search store '%s' with ID: %s", storeName, storeID)

			// Save to database (ID, CreatedAt, UpdatedAt will be set by database defaults)
			newStore := &models.GeminiFileSearchStore{
				StoreName: storeName,
				StoreID:   storeID,
			}

			if dbErr := s.geminiStoreRepo.Create(ctx, newStore); dbErr != nil {
				log.Printf("[Gemini] Store Initialization: WARNING - Failed to save store to database: %v", dbErr)
				// Don't fail - the store was created successfully in Gemini
			} else {
				log.Printf("[Gemini] Store Initialization: Database record created for store '%s'", storeName)
			}

			s.storeID = storeID
			return storeID, nil
		}

		// Log failures with retry information
		if attempt < 3 {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			log.Printf("[Gemini] Store Initialization: RETRY - Attempt %d/3 failed for store '%s', retrying in %v. Error: %v",
				attempt, storeName, backoff, err)
			time.Sleep(backoff)
		} else {
			// Final attempt failed
			log.Printf("[Gemini] Store Initialization: FAILED - All 3 attempts failed for store '%s'. Final error: %v",
				storeName, err)
		}
	}

	return "", fmt.Errorf("%w: %v", ErrGeminiStoreCreation, err)
}

// UploadDocument uploads a document to a File Search store with metadata
func (s *geminiService) UploadDocument(ctx context.Context, storeID, graphID, graphName, documentID string, content []byte, mimeType string) (string, error) {
	// Use shared store ID from service if storeID parameter is empty
	if storeID == "" {
		storeID = s.storeID
	}

	// Log upload with graph_id, domain, version metadata
	log.Printf("[Gemini] Document Upload: Starting upload for document '%s' to store '%s' | Graph: %s (%s) | Metadata: [graph_id=%s, domain=topeic.com, version=1.1] | Size: %d bytes | Type: %s",
		documentID, storeID, graphID, graphName, graphID, len(content), mimeType)

	var op *genai.UploadToFileSearchStoreOperation
	var err error

	// Retry logic: 3 attempts with exponential backoff
	for attempt := 1; attempt <= 3; attempt++ {
		reader := bytes.NewReader(content)
		op, err = s.client.FileSearchStores.UploadToFileSearchStore(ctx, reader, storeID, &genai.UploadToFileSearchStoreConfig{
			DisplayName: graphName,
			MIMEType:    mimeType,
			CustomMetadata: []*genai.CustomMetadata{
				{Key: "graph_id", StringValue: graphID},
				{Key: "domain", StringValue: "topeic.com"},
				{Key: "version", StringValue: "1.1"},
			},
		})

		if err == nil {
			break
		}

		// Log failures with detailed error
		if attempt < 3 {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			log.Printf("[Gemini] Document Upload: RETRY - Attempt %d/3 failed for document '%s' (graph: %s), retrying in %v. Error: %v",
				attempt, documentID, graphID, backoff, err)
			time.Sleep(backoff)
		} else {
			// Final attempt failed
			log.Printf("[Gemini] Document Upload: FAILED - All 3 attempts failed for document '%s' (graph: %s, graph_id=%s, domain=topeic.com, version=1.1). Final error: %v",
				documentID, graphID, graphID, err)
		}
	}

	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrGeminiUploadFailed, err)
	}

	// Wait for the upload operation to complete
	log.Printf("[Gemini] Document Upload: Waiting for upload operation to complete for document '%s'", documentID)

	maxWaitTime := 2 * time.Minute
	pollInterval := 5 * time.Second
	startTime := time.Now()

	for !op.Done {
		// Check if we've exceeded max wait time
		if time.Since(startTime) > maxWaitTime {
			log.Printf("[Gemini] Document Upload: TIMEOUT - Upload operation exceeded max wait time of %v for document '%s'",
				maxWaitTime, documentID)
			return "", fmt.Errorf("%w: upload operation timeout", ErrGeminiUploadFailed)
		}

		// Wait before polling again
		time.Sleep(pollInterval)
		log.Printf("[Gemini] Document Upload: Polling operation status for document '%s'...", documentID)

		// Get updated operation status
		op, err = s.client.Operations.GetUploadToFileSearchStoreOperation(ctx, op, nil)
		if err != nil {
			log.Printf("[Gemini] Document Upload: ERROR - Failed to get operation status for document '%s': %v",
				documentID, err)
			return "", fmt.Errorf("%w: failed to poll operation status: %v", ErrGeminiUploadFailed, err)
		}
	}

	// Operation completed, check response
	if op.Response == nil || op.Response.DocumentName == "" {
		log.Printf("[Gemini] Document Upload: ERROR - Upload operation completed but no document information returned for document '%s' (graph: %s)",
			documentID, graphID)
		return "", fmt.Errorf("%w: no document information in response", ErrGeminiUploadFailed)
	}

	fileID := op.Response.DocumentName

	// Log successful upload with file ID
	log.Printf("[Gemini] Document Upload: SUCCESS - Uploaded document '%s' with file ID: %s | Metadata: [graph_id=%s, domain=topeic.com, version=1.1]",
		documentID, fileID, graphID)

	// Update document with Gemini file ID
	if err := s.docRepo.UpdateGeminiFileID(ctx, documentID, fileID); err != nil {
		log.Printf("[Gemini] Document Upload: WARNING - Failed to update document '%s' with file ID '%s' in database: %v",
			documentID, fileID, err)
		// Don't fail - upload was successful
	}

	return fileID, nil
}

// GenerateStreamingResponse generates a streaming AI response using File Search with metadata filtering
func (s *geminiService) GenerateStreamingResponse(ctx context.Context, storeID, graphID, domain, version, query string, responseChan chan<- string) error {
	// NOTE: Do NOT close responseChan here - let the caller manage channel lifecycle
	// The caller needs to know when streaming completes vs when an error occurs

	// Use shared store ID from service if storeID parameter is empty
	if storeID == "" {
		storeID = s.storeID
	}

	// Log query execution with graph_id
	log.Printf("[Gemini] Query Filtering: Starting query execution | Store: %s | Graph ID: %s | Domain: %s | Version: %s | Query: %.100s...",
		storeID, graphID, domain, version, query)

	// Build metadata filter expression
	// Escape special characters in values to prevent injection
	escapedGraphID := escapeFilterValue(graphID)
	escapedDomain := escapeFilterValue(domain)
	escapedVersion := escapeFilterValue(version)

	// Construct filter: (chunk.custom_metadata.graph_id = "{graphID}" AND chunk.custom_metadata.domain = "{domain}" AND chunk.custom_metadata.version = "{version}")
	metadataFilter := fmt.Sprintf(
		`(chunk.custom_metadata.graph_id = "%s" AND chunk.custom_metadata.domain = "%s" AND chunk.custom_metadata.version = "%s")`,
		escapedGraphID, escapedDomain, escapedVersion,
	)

	// Validate filter syntax (basic check)
	if err := validateFilterSyntax(metadataFilter); err != nil {
		log.Printf("[Gemini] Query Filtering: ERROR - Invalid metadata filter syntax for graph '%s': %v", graphID, err)
		return fmt.Errorf("invalid metadata filter: %w", err)
	}

	// Log metadata filter expression used
	log.Printf("[Gemini] Query Filtering: Using metadata filter expression: %s", metadataFilter)

	// Create the prompt with File Search tool
	prompt := fmt.Sprintf("Based on the documents in the knowledge graph, please answer the following question: %s", query)
	contents := []*genai.Content{
		{
			Role: genai.RoleUser,
			Parts: []*genai.Part{
				genai.NewPartFromText(prompt),
			},
		},
	}

	// Configure with File Search tool and metadata filter
	config := &genai.GenerateContentConfig{
		Tools: []*genai.Tool{
			{
				FileSearch: &genai.FileSearch{
					FileSearchStoreNames: []string{storeID},
					MetadataFilter:       metadataFilter,
				},
			},
		},
	}

	log.Printf("[Gemini] Query Filtering: Initiating streaming response for graph '%s'", graphID)

	// Generate streaming response
	responseIter := s.client.Models.GenerateContentStream(ctx, "gemini-2.5-flash", contents, config)

	// Process the stream
	chunkCount := 0
	var lastErr error
	for resp, err := range responseIter {
		if err != nil {
			// Store the error but continue - the iterator might return an error at the end
			lastErr = err
			log.Printf("[Gemini] Query Filtering: Iterator returned error for graph '%s' after %d chunks: %v",
				graphID, chunkCount, err)
			// Don't return immediately - check if we got any chunks
			break
		}

		// Extract text from response
		for _, cand := range resp.Candidates {
			if cand.Content != nil {
				for _, part := range cand.Content.Parts {
					if part.Text != "" {
						chunk := part.Text
						chunkCount++
						select {
						case responseChan <- chunk:
							// Chunk sent successfully
						case <-ctx.Done():
							log.Printf("[Gemini] Query Filtering: CANCELLED - Context cancelled during streaming for graph '%s' after %d chunks",
								graphID, chunkCount)
							return ctx.Err()
						}
					}
				}
			}
		}
	}

	// Check if we got any chunks - if yes, consider it a success even if there was an error at the end
	if chunkCount > 0 {
		// Log streaming completion
		log.Printf("[Gemini] Query Filtering: SUCCESS - Streaming complete for graph '%s' | Total chunks: %d | Filter: [graph_id=%s, domain=%s, version=%s]",
			graphID, chunkCount, graphID, domain, version)

		// If there was an error but we got chunks, log it but don't fail
		if lastErr != nil {
			log.Printf("[Gemini] Query Filtering: NOTE - Iterator returned error after successful streaming (this is normal): %v", lastErr)
		}

		return nil
	}

	// No chunks received - this is a real error
	if lastErr != nil {
		log.Printf("[Gemini] Query Filtering: ERROR - No chunks received and iterator returned error for graph '%s': %v",
			graphID, lastErr)
		return fmt.Errorf("%w: %v", ErrGeminiQueryFailed, lastErr)
	}

	// No chunks and no error - empty response
	log.Printf("[Gemini] Query Filtering: WARNING - No chunks received for graph '%s' (empty response)", graphID)
	return nil
}

// escapeFilterValue escapes special characters in metadata filter values
func escapeFilterValue(value string) string {
	// Escape double quotes and backslashes
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return value
}

// validateFilterSyntax performs basic validation of the metadata filter syntax
func validateFilterSyntax(filter string) error {
	// Check for balanced parentheses
	openCount := strings.Count(filter, "(")
	closeCount := strings.Count(filter, ")")
	if openCount != closeCount {
		return fmt.Errorf("unbalanced parentheses in filter expression")
	}

	// Check for required components
	if !strings.Contains(filter, "chunk.custom_metadata.") {
		return fmt.Errorf("filter must contain chunk.custom_metadata prefix")
	}

	// Check for AND operators
	if !strings.Contains(filter, " AND ") {
		return fmt.Errorf("filter must contain AND operators")
	}

	// Check that filter is not empty
	if strings.TrimSpace(filter) == "" {
		return fmt.Errorf("filter cannot be empty")
	}

	return nil
}
