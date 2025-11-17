package handler

import (
	"errors"
	"net/http"

	"github.com/bipulkrdas/orgmind/backend/internal/middleware"
	"github.com/bipulkrdas/orgmind/backend/internal/models"
	"github.com/bipulkrdas/orgmind/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// GraphHandler handles graph-related HTTP requests
type GraphHandler struct {
	graphService    service.GraphService
	documentService service.DocumentService
	zepService      service.ZepService
}

// NewGraphHandler creates a new instance of GraphHandler
func NewGraphHandler(graphService service.GraphService, documentService service.DocumentService, zepService service.ZepService) *GraphHandler {
	return &GraphHandler{
		graphService:    graphService,
		documentService: documentService,
		zepService:      zepService,
	}
}

// GraphResponse represents a graph in API responses
type GraphResponse struct {
	ID            string  `json:"id"`
	CreatorID     string  `json:"creatorId"`
	ZepGraphID    string  `json:"zepGraphId"`
	Name          string  `json:"name"`
	Description   *string `json:"description,omitempty"`
	DocumentCount int     `json:"documentCount"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
}

// GraphMembershipResponse represents a graph membership in API responses
type GraphMembershipResponse struct {
	ID        string `json:"id"`
	GraphID   string `json:"graphId"`
	UserID    string `json:"userId"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt"`
}

// CreateGraph handles POST /api/graphs
func (h *GraphHandler) CreateGraph(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	var req models.CreateGraphRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Create graph
	graph, err := h.graphService.Create(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create graph", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, GraphResponse{
		ID:            graph.ID,
		CreatorID:     graph.CreatorID,
		ZepGraphID:    graph.ZepGraphID,
		Name:          graph.Name,
		Description:   graph.Description,
		DocumentCount: graph.DocumentCount,
		CreatedAt:     graph.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     graph.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// ListGraphs handles GET /api/graphs
func (h *GraphHandler) ListGraphs(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// List all graphs the user is a member of
	graphs, err := h.graphService.ListByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list graphs", "details": err.Error()})
		return
	}

	// Convert to response format
	response := make([]GraphResponse, len(graphs))
	for i, graph := range graphs {
		response[i] = GraphResponse{
			ID:            graph.ID,
			CreatorID:     graph.CreatorID,
			ZepGraphID:    graph.ZepGraphID,
			Name:          graph.Name,
			Description:   graph.Description,
			DocumentCount: graph.DocumentCount,
			CreatedAt:     graph.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     graph.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	c.JSON(http.StatusOK, gin.H{"graphs": response})
}

// GetGraph handles GET /api/graphs/:id
func (h *GraphHandler) GetGraph(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Get graph ID from URL parameter
	graphID := c.Param("id")
	if graphID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Graph ID is required"})
		return
	}

	// Get graph with membership verification
	graph, err := h.graphService.GetByID(c.Request.Context(), graphID, userID)
	if err != nil {
		if errors.Is(err, service.ErrGraphNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Graph not found"})
			return
		}
		if errors.Is(err, service.ErrNotGraphMember) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to this graph"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get graph", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, GraphResponse{
		ID:            graph.ID,
		CreatorID:     graph.CreatorID,
		ZepGraphID:    graph.ZepGraphID,
		Name:          graph.Name,
		Description:   graph.Description,
		DocumentCount: graph.DocumentCount,
		CreatedAt:     graph.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     graph.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// UpdateGraph handles PUT /api/graphs/:id
func (h *GraphHandler) UpdateGraph(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Get graph ID from URL parameter
	graphID := c.Param("id")
	if graphID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Graph ID is required"})
		return
	}

	var req models.UpdateGraphRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Update graph (creator verification happens in service)
	graph, err := h.graphService.Update(c.Request.Context(), graphID, userID, &req)
	if err != nil {
		if errors.Is(err, service.ErrGraphNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Graph not found"})
			return
		}
		if errors.Is(err, service.ErrNotGraphCreator) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only the graph creator can update this graph"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update graph", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, GraphResponse{
		ID:            graph.ID,
		CreatorID:     graph.CreatorID,
		ZepGraphID:    graph.ZepGraphID,
		Name:          graph.Name,
		Description:   graph.Description,
		DocumentCount: graph.DocumentCount,
		CreatedAt:     graph.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     graph.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// DeleteGraph handles DELETE /api/graphs/:id
func (h *GraphHandler) DeleteGraph(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Get graph ID from URL parameter
	graphID := c.Param("id")
	if graphID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Graph ID is required"})
		return
	}

	// Delete graph (creator verification happens in service)
	err := h.graphService.Delete(c.Request.Context(), graphID, userID)
	if err != nil {
		if errors.Is(err, service.ErrGraphNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Graph not found"})
			return
		}
		if errors.Is(err, service.ErrNotGraphCreator) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only the graph creator can delete this graph"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete graph", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Graph deleted successfully"})
}

// AddMember handles POST /api/graphs/:id/members
func (h *GraphHandler) AddMember(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Get graph ID from URL parameter
	graphID := c.Param("id")
	if graphID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Graph ID is required"})
		return
	}

	var req models.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Add member (creator verification happens in service)
	err := h.graphService.AddMember(c.Request.Context(), graphID, userID, &req)
	if err != nil {
		if errors.Is(err, service.ErrGraphNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Graph not found"})
			return
		}
		if errors.Is(err, service.ErrNotGraphCreator) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only the graph creator can add members"})
			return
		}
		if errors.Is(err, service.ErrMemberAlreadyExists) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User is already a member of this graph"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Member added successfully"})
}

// RemoveMember handles DELETE /api/graphs/:id/members/:userId
func (h *GraphHandler) RemoveMember(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Get graph ID from URL parameter
	graphID := c.Param("id")
	if graphID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Graph ID is required"})
		return
	}

	// Get member user ID from URL parameter
	memberUserID := c.Param("userId")
	if memberUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// Remove member (creator verification happens in service)
	err := h.graphService.RemoveMember(c.Request.Context(), graphID, userID, memberUserID)
	if err != nil {
		if errors.Is(err, service.ErrGraphNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Graph not found"})
			return
		}
		if errors.Is(err, service.ErrNotGraphCreator) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only the graph creator can remove members"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
}

// ListMembers handles GET /api/graphs/:id/members
func (h *GraphHandler) ListMembers(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Get graph ID from URL parameter
	graphID := c.Param("id")
	if graphID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Graph ID is required"})
		return
	}

	// List members (membership verification happens in service)
	members, err := h.graphService.ListMembers(c.Request.Context(), graphID, userID)
	if err != nil {
		if errors.Is(err, service.ErrGraphNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Graph not found"})
			return
		}
		if errors.Is(err, service.ErrNotGraphMember) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to this graph"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list members", "details": err.Error()})
		return
	}

	// Convert to response format
	response := make([]GraphMembershipResponse, len(members))
	for i, member := range members {
		response[i] = GraphMembershipResponse{
			ID:        member.ID,
			GraphID:   member.GraphID,
			UserID:    member.UserID,
			Role:      member.Role,
			CreatedAt: member.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	c.JSON(http.StatusOK, gin.H{"members": response})
}

// ListGraphDocuments handles GET /api/graphs/:id/documents
func (h *GraphHandler) ListGraphDocuments(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Get graph ID from URL parameter
	graphID := c.Param("id")
	if graphID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Graph ID is required"})
		return
	}

	// Verify membership before listing documents
	_, err := h.graphService.GetByID(c.Request.Context(), graphID, userID)
	if err != nil {
		if errors.Is(err, service.ErrGraphNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Graph not found"})
			return
		}
		if errors.Is(err, service.ErrNotGraphMember) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to this graph"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify graph access", "details": err.Error()})
		return
	}

	// List documents for the graph
	docs, err := h.documentService.ListGraphDocuments(c.Request.Context(), graphID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list documents", "details": err.Error()})
		return
	}

	// Convert to response format
	response := make([]DocumentResponse, len(docs))
	for i, doc := range docs {
		response[i] = DocumentResponse{
			ID:          doc.ID,
			UserID:      doc.UserID,
			GraphID:     doc.GraphID,
			Filename:    doc.Filename,
			ContentType: doc.ContentType,
			StorageKey:  doc.StorageKey,
			SizeBytes:   doc.SizeBytes,
			Source:      doc.Source,
			Status:      doc.Status,
			CreatedAt:   doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   doc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	c.JSON(http.StatusOK, gin.H{"documents": response})
}

// GetGraphVisualization handles GET /api/graphs/:id/visualization
func (h *GraphHandler) GetGraphVisualization(c *gin.Context) {
	// Extract userID from JWT token (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Get graph ID from URL parameter
	graphID := c.Param("id")
	if graphID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Graph ID is required"})
		return
	}

	// Get query parameter (default to "all" if not provided)
	query := c.DefaultQuery("query", "all")

	// Verify membership and get graph details
	graph, err := h.graphService.GetByID(c.Request.Context(), graphID, userID)
	if err != nil {
		if errors.Is(err, service.ErrGraphNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Graph not found"})
			return
		}
		if errors.Is(err, service.ErrNotGraphMember) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to this graph"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify graph access", "details": err.Error()})
		return
	}

	// Get graph visualization data from Zep with query filter
	graphData, err := h.zepService.GetGraph(c.Request.Context(), graph.ZepGraphID, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get graph visualization", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, graphData)
}
