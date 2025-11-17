package service

import (
	"context"

	"github.com/bipulkrdas/orgmind/backend/internal/models"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	SignUp(ctx context.Context, email, password, firstName, lastName string) (*models.User, string, error)
	SignIn(ctx context.Context, email, password string) (string, error)
	InitiateOAuth(ctx context.Context, provider string) (string, error)
	HandleOAuthCallback(ctx context.Context, provider, code string) (string, error)
	ResetPassword(ctx context.Context, email string) error
	UpdatePassword(ctx context.Context, token, newPassword string) error
}

// ProcessingService defines the interface for document processing operations
type ProcessingService interface {
	ProcessDocument(ctx context.Context, userID, graphID, documentID, content string) error
}

// ZepService defines the interface for Zep Cloud integration
type ZepService interface {
	// Create a new graph in Zep Cloud
	CreateGraph(ctx context.Context, graphID, name string, description *string) (string, error)

	// Delete a graph from Zep Cloud
	DeleteGraph(ctx context.Context, zepGraphID string) error

	// Add memory to a specific graph
	AddMemory(ctx context.Context, graphID string, chunks []string, metadata map[string]any) error

	// Get graph data for visualization with optional query filter
	GetGraph(ctx context.Context, graphID, query string) (*models.GraphData, error)

	// Search memory in a specific graph
	SearchMemory(ctx context.Context, graphID, query string) ([]models.MemoryResult, error)
}

// DocumentService defines the interface for document operations
type DocumentService interface {
	CreateFromEditor(ctx context.Context, userID, graphID, plainText, lexicalState string) (*models.Document, error)
	CreateFromFile(ctx context.Context, userID, graphID string, file []byte, filename, contentType string) (*models.Document, error)
	GetDocument(ctx context.Context, documentID, userID string) (*models.Document, error)
	GetDocumentContent(ctx context.Context, documentID, userID string) (map[string]interface{}, error)
	ListUserDocuments(ctx context.Context, userID string) ([]*models.Document, error)
	ListGraphDocuments(ctx context.Context, graphID string) ([]*models.Document, error)
	UpdateDocument(ctx context.Context, documentID, userID, plainText, lexicalState string) (*models.Document, error)
	DeleteDocument(ctx context.Context, documentID, userID string) error
}

// GraphService defines the interface for graph operations
type GraphService interface {
	// Create a new graph for a user (creator becomes owner)
	Create(ctx context.Context, creatorID string, req *models.CreateGraphRequest) (*models.Graph, error)

	// Get a graph by ID with membership verification
	GetByID(ctx context.Context, graphID, userID string) (*models.Graph, error)

	// List all graphs the user is a member of
	ListByUserID(ctx context.Context, userID string) ([]*models.Graph, error)

	// Update graph metadata (creator only)
	Update(ctx context.Context, graphID, userID string, req *models.UpdateGraphRequest) (*models.Graph, error)

	// Delete a graph and all associated data (creator only)
	Delete(ctx context.Context, graphID, userID string) error

	// Add a member to a graph (creator only)
	AddMember(ctx context.Context, graphID, creatorID string, req *models.AddMemberRequest) error

	// Remove a member from a graph (creator only)
	RemoveMember(ctx context.Context, graphID, creatorID, memberUserID string) error

	// List all members of a graph
	ListMembers(ctx context.Context, graphID, userID string) ([]*models.GraphMembership, error)

	// Check if user is a member of a graph
	IsMember(ctx context.Context, graphID, userID string) (bool, error)

	// Increment document count for a graph
	IncrementDocumentCount(ctx context.Context, graphID string) error

	// Decrement document count for a graph
	DecrementDocumentCount(ctx context.Context, graphID string) error
}
