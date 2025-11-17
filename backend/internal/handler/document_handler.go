package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/bipulkrdas/orgmind/backend/internal/middleware"
	"github.com/bipulkrdas/orgmind/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// DocumentHandler handles document-related HTTP requests
type DocumentHandler struct {
	documentService service.DocumentService
}

// NewDocumentHandler creates a new instance of DocumentHandler
func NewDocumentHandler(documentService service.DocumentService) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
	}
}

// EditorSubmitRequest represents the request body for editor content submission
type EditorSubmitRequest struct {
	Content      string `json:"content" binding:"required"`      // Plain text for Zep processing
	LexicalState string `json:"lexicalState" binding:"required"` // Lexical JSON for editor restoration
	GraphID      string `json:"graphId" binding:"required"`
}

// UpdateDocumentRequest represents the request body for updating document content
type UpdateDocumentRequest struct {
	Content      string `json:"content" binding:"required"`      // Plain text for Zep processing
	LexicalState string `json:"lexicalState" binding:"required"` // Lexical JSON for editor restoration
}

// DocumentResponse represents a document in API responses
type DocumentResponse struct {
	ID           string  `json:"id"`
	UserID       string  `json:"userId"`
	GraphID      *string `json:"graphId,omitempty"`
	Filename     *string `json:"filename,omitempty"`
	ContentType  *string `json:"contentType,omitempty"`
	StorageKey   string  `json:"storageKey"`
	SizeBytes    int64   `json:"sizeBytes"`
	Source       string  `json:"source"`
	Status       string  `json:"status"`
	ErrorMessage *string `json:"errorMessage,omitempty"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

// SubmitEditorContent handles POST /api/documents/editor
func (h *DocumentHandler) SubmitEditorContent(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	var req EditorSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Create document from editor content (with both plain text and Lexical state)
	doc, err := h.documentService.CreateFromEditor(c.Request.Context(), userID, req.GraphID, req.Content, req.LexicalState)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create document", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, DocumentResponse{
		ID:           doc.ID,
		UserID:       doc.UserID,
		GraphID:      doc.GraphID,
		Filename:     doc.Filename,
		ContentType:  doc.ContentType,
		StorageKey:   doc.StorageKey,
		SizeBytes:    doc.SizeBytes,
		Source:       doc.Source,
		Status:       doc.Status,
		ErrorMessage: doc.ErrorMessage,
		CreatedAt:    doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    doc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// UploadFile handles POST /api/documents/upload
func (h *DocumentHandler) UploadFile(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Parse multipart form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read file from request", "details": err.Error()})
		return
	}
	defer file.Close()

	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file content", "details": err.Error()})
		return
	}

	// Get content type from header or detect it
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(fileBytes)
	}

	// Get graphId from form field
	graphID := c.PostForm("graphId")
	if graphID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "graphId is required"})
		return
	}

	// Create document from file
	doc, err := h.documentService.CreateFromFile(c.Request.Context(), userID, graphID, fileBytes, header.Filename, contentType)
	if err != nil {
		// Provide more specific error responses based on error type
		errMsg := err.Error()
		if strings.Contains(errMsg, "unsupported") || strings.Contains(errMsg, "not supported") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Unsupported file format",
				"message": errMsg,
			})
		} else if strings.Contains(errMsg, "file size exceeds") || strings.Contains(errMsg, "maximum allowed size") {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":   "File too large",
				"message": errMsg,
			})
		} else if strings.Contains(errMsg, "password-protected") || strings.Contains(errMsg, "password") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Password-protected documents not supported",
				"message": errMsg,
			})
		} else if strings.Contains(errMsg, "corrupted") || strings.Contains(errMsg, "invalid") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid or corrupted file",
				"message": errMsg,
			})
		} else if strings.Contains(errMsg, "took too long") || strings.Contains(errMsg, "timeout") {
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error":   "Extraction timeout",
				"message": errMsg,
			})
		} else if strings.Contains(errMsg, "empty") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Empty file",
				"message": errMsg,
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Failed to process document",
				"message": errMsg,
			})
		}
		return
	}

	c.JSON(http.StatusCreated, DocumentResponse{
		ID:           doc.ID,
		UserID:       doc.UserID,
		GraphID:      doc.GraphID,
		Filename:     doc.Filename,
		ContentType:  doc.ContentType,
		StorageKey:   doc.StorageKey,
		SizeBytes:    doc.SizeBytes,
		Source:       doc.Source,
		Status:       doc.Status,
		ErrorMessage: doc.ErrorMessage,
		CreatedAt:    doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    doc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// ListDocuments handles GET /api/documents
func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Get all documents for the user
	docs, err := h.documentService.ListUserDocuments(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list documents", "details": err.Error()})
		return
	}

	// Convert to response format
	response := make([]DocumentResponse, len(docs))
	for i, doc := range docs {
		response[i] = DocumentResponse{
			ID:           doc.ID,
			UserID:       doc.UserID,
			GraphID:      doc.GraphID,
			Filename:     doc.Filename,
			ContentType:  doc.ContentType,
			StorageKey:   doc.StorageKey,
			SizeBytes:    doc.SizeBytes,
			Source:       doc.Source,
			Status:       doc.Status,
			ErrorMessage: doc.ErrorMessage,
			CreatedAt:    doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    doc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	c.JSON(http.StatusOK, gin.H{"documents": response})
}

// GetDocument handles GET /api/documents/:id
func (h *DocumentHandler) GetDocument(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Get document ID from URL parameter
	documentID := c.Param("id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	// Get document
	doc, err := h.documentService.GetDocument(c.Request.Context(), documentID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DocumentResponse{
		ID:           doc.ID,
		UserID:       doc.UserID,
		GraphID:      doc.GraphID,
		Filename:     doc.Filename,
		ContentType:  doc.ContentType,
		StorageKey:   doc.StorageKey,
		SizeBytes:    doc.SizeBytes,
		Source:       doc.Source,
		Status:       doc.Status,
		ErrorMessage: doc.ErrorMessage,
		CreatedAt:    doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    doc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// UpdateDocument handles PUT /api/documents/:id
func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Get document ID from URL parameter
	documentID := c.Param("id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	// Parse request body
	var req UpdateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Update document (with both plain text and Lexical state)
	doc, err := h.documentService.UpdateDocument(c.Request.Context(), documentID, userID, req.Content, req.LexicalState)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update document", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DocumentResponse{
		ID:           doc.ID,
		UserID:       doc.UserID,
		GraphID:      doc.GraphID,
		Filename:     doc.Filename,
		ContentType:  doc.ContentType,
		StorageKey:   doc.StorageKey,
		SizeBytes:    doc.SizeBytes,
		Source:       doc.Source,
		Status:       doc.Status,
		ErrorMessage: doc.ErrorMessage,
		CreatedAt:    doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    doc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// DeleteDocument handles DELETE /api/documents/:id
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Get document ID from URL parameter
	documentID := c.Param("id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	// Delete document
	err := h.documentService.DeleteDocument(c.Request.Context(), documentID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete document", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document deleted successfully"})
}

// GetDocumentContent handles GET /api/documents/:id/content
func (h *DocumentHandler) GetDocumentContent(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Get document ID from URL parameter
	documentID := c.Param("id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	// Get document content
	content, err := h.documentService.GetDocumentContent(c.Request.Context(), documentID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get document content", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, content)
}
