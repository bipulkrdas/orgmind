package service

import (
	"context"
	"fmt"
	"time"

	"github.com/bipulkrdas/orgmind/backend/internal/models"
	"github.com/bipulkrdas/orgmind/backend/internal/repository"
	"github.com/google/uuid"
)

// Custom errors for graph operations
var (
	ErrGraphNotFound       = fmt.Errorf("graph not found")
	ErrGraphUnauthorized   = fmt.Errorf("you don't have access to this graph")
	ErrNotGraphCreator     = fmt.Errorf("only the graph creator can perform this action")
	ErrNotGraphMember      = fmt.Errorf("you are not a member of this graph")
	ErrMemberAlreadyExists = fmt.Errorf("user is already a member of this graph")
	ErrZepGraphCreation    = fmt.Errorf("failed to create graph in Zep Cloud")
	ErrZepGraphDeletion    = fmt.Errorf("failed to delete graph from Zep Cloud")
)

// graphService implements the GraphService interface
type graphService struct {
	graphRepo repository.GraphRepository
	zepSvc    ZepService
}

// NewGraphService creates a new graph service instance
func NewGraphService(graphRepo repository.GraphRepository, zepSvc ZepService) GraphService {
	return &graphService{
		graphRepo: graphRepo,
		zepSvc:    zepSvc,
	}
}

// verifyMembership checks if user is a member of the graph
func (s *graphService) verifyMembership(ctx context.Context, graphID, userID string) (*models.Graph, error) {
	graph, err := s.graphRepo.GetByID(ctx, graphID)
	if err != nil {
		return nil, ErrGraphNotFound
	}

	isMember, err := s.graphRepo.IsMember(ctx, graphID, userID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, ErrNotGraphMember
	}

	return graph, nil
}

// verifyCreator checks if user is the creator of the graph
func (s *graphService) verifyCreator(ctx context.Context, graphID, userID string) (*models.Graph, error) {
	graph, err := s.graphRepo.GetByID(ctx, graphID)
	if err != nil {
		return nil, ErrGraphNotFound
	}

	if graph.CreatorID != userID {
		return nil, ErrNotGraphCreator
	}

	return graph, nil
}

// Create creates a new graph in Zep Cloud, saves to DB, and creates owner membership
func (s *graphService) Create(ctx context.Context, creatorID string, req *models.CreateGraphRequest) (*models.Graph, error) {
	// Generate a unique graph ID
	graphID := uuid.New().String()

	// Step 1: Create graph in Zep Cloud FIRST (critical dependency)
	// If Zep creation fails, we don't create database records
	zepGraphID := fmt.Sprintf("graph-%s", graphID)

	actualZepGraphID, err := s.zepSvc.CreateGraph(ctx, zepGraphID, req.Name, req.Description)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrZepGraphCreation, err)
	}

	// Use the Zep-assigned graph ID if different from our generated one
	if actualZepGraphID != "" {
		zepGraphID = actualZepGraphID
	}

	// Step 2: Create graph record in database
	now := time.Now()
	graph := &models.Graph{
		ID:            graphID,
		CreatorID:     creatorID,
		ZepGraphID:    zepGraphID,
		Name:          req.Name,
		Description:   req.Description,
		DocumentCount: 0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.graphRepo.Create(ctx, graph); err != nil {
		// Rollback: delete from Zep if database creation fails
		_ = s.zepSvc.DeleteGraph(ctx, zepGraphID)
		return nil, fmt.Errorf("failed to create graph in database: %w", err)
	}

	// Step 3: Create owner membership for the creator
	membership := &models.GraphMembership{
		ID:        uuid.New().String(),
		GraphID:   graphID,
		UserID:    creatorID,
		Role:      "owner",
		CreatedAt: now,
	}

	if err := s.graphRepo.CreateMembership(ctx, membership); err != nil {
		// Rollback: delete both database record and Zep graph
		_ = s.graphRepo.Delete(ctx, graphID)
		_ = s.zepSvc.DeleteGraph(ctx, zepGraphID)
		return nil, fmt.Errorf("failed to create owner membership: %w", err)
	}

	return graph, nil
}

// GetByID retrieves a graph by ID with membership verification
func (s *graphService) GetByID(ctx context.Context, graphID, userID string) (*models.Graph, error) {
	return s.verifyMembership(ctx, graphID, userID)
}

// ListByUserID returns all graphs the user is a member of
func (s *graphService) ListByUserID(ctx context.Context, userID string) ([]*models.Graph, error) {
	graphs, err := s.graphRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list graphs: %w", err)
	}

	return graphs, nil
}

// Update updates graph metadata (creator only)
func (s *graphService) Update(ctx context.Context, graphID, userID string, req *models.UpdateGraphRequest) (*models.Graph, error) {
	// Verify user is the creator
	graph, err := s.verifyCreator(ctx, graphID, userID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		graph.Name = *req.Name
	}
	if req.Description != nil {
		graph.Description = req.Description
	}
	graph.UpdatedAt = time.Now()

	// Save to database
	if err := s.graphRepo.Update(ctx, graph); err != nil {
		return nil, fmt.Errorf("failed to update graph: %w", err)
	}

	return graph, nil
}

// Delete deletes a graph and all associated data (creator only)
func (s *graphService) Delete(ctx context.Context, graphID, userID string) error {
	// Verify user is the creator
	graph, err := s.verifyCreator(ctx, graphID, userID)
	if err != nil {
		return err
	}

	// Step 1: Delete from Zep Cloud first
	if err := s.zepSvc.DeleteGraph(ctx, graph.ZepGraphID); err != nil {
		// Log the error but continue with database deletion
		// The Zep graph might already be deleted or not exist
		fmt.Printf("Warning: failed to delete graph from Zep (continuing with database deletion): %v\n", err)
	}

	// Step 2: Delete from database (cascade deletes memberships and documents)
	if err := s.graphRepo.Delete(ctx, graph.ID); err != nil {
		return fmt.Errorf("failed to delete graph from database: %w", err)
	}

	return nil
}

// AddMember adds a member to a graph (creator only)
func (s *graphService) AddMember(ctx context.Context, graphID, creatorID string, req *models.AddMemberRequest) error {
	// Verify user is the creator
	_, err := s.verifyCreator(ctx, graphID, creatorID)
	if err != nil {
		return err
	}

	// Check if user is already a member
	isMember, err := s.graphRepo.IsMember(ctx, graphID, req.UserID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if isMember {
		return ErrMemberAlreadyExists
	}

	// Set default role if not provided
	role := req.Role
	if role == "" {
		role = "member"
	}

	// Create membership
	membership := &models.GraphMembership{
		ID:        uuid.New().String(),
		GraphID:   graphID,
		UserID:    req.UserID,
		Role:      role,
		CreatedAt: time.Now(),
	}

	if err := s.graphRepo.CreateMembership(ctx, membership); err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	return nil
}

// RemoveMember removes a member from a graph (creator only)
func (s *graphService) RemoveMember(ctx context.Context, graphID, creatorID, memberUserID string) error {
	// Verify user is the creator
	_, err := s.verifyCreator(ctx, graphID, creatorID)
	if err != nil {
		return err
	}

	// Prevent creator from removing themselves
	if creatorID == memberUserID {
		return fmt.Errorf("creator cannot remove themselves from the graph")
	}

	// Delete membership
	if err := s.graphRepo.DeleteMembership(ctx, graphID, memberUserID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	return nil
}

// ListMembers lists all members of a graph (requires membership)
func (s *graphService) ListMembers(ctx context.Context, graphID, userID string) ([]*models.GraphMembership, error) {
	// Verify user is a member
	_, err := s.verifyMembership(ctx, graphID, userID)
	if err != nil {
		return nil, err
	}

	// Get all members
	members, err := s.graphRepo.ListMembersByGraphID(ctx, graphID)
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}

	return members, nil
}

// IsMember checks if user is a member of a graph
func (s *graphService) IsMember(ctx context.Context, graphID, userID string) (bool, error) {
	isMember, err := s.graphRepo.IsMember(ctx, graphID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check membership: %w", err)
	}

	return isMember, nil
}

// IncrementDocumentCount increments the document count for a graph
func (s *graphService) IncrementDocumentCount(ctx context.Context, graphID string) error {
	if err := s.graphRepo.UpdateDocumentCount(ctx, graphID, 1); err != nil {
		return fmt.Errorf("failed to increment document count: %w", err)
	}

	return nil
}

// DecrementDocumentCount decrements the document count for a graph
func (s *graphService) DecrementDocumentCount(ctx context.Context, graphID string) error {
	if err := s.graphRepo.UpdateDocumentCount(ctx, graphID, -1); err != nil {
		return fmt.Errorf("failed to decrement document count: %w", err)
	}

	return nil
}
