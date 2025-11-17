package repository

import (
	"context"

	"github.com/bipulkrdas/orgmind/backend/internal/models"
)

// UserRepository defines the interface for user data access operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, userID string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
}

// DocumentRepository defines the interface for document data access operations
type DocumentRepository interface {
	Create(ctx context.Context, doc *models.Document) error
	GetByID(ctx context.Context, docID string) (*models.Document, error)
	ListByUserID(ctx context.Context, userID string) ([]*models.Document, error)
	ListByGraphID(ctx context.Context, graphID string) ([]*models.Document, error)
	Update(ctx context.Context, doc *models.Document) error
	Delete(ctx context.Context, docID string) error
}

// PasswordResetTokenRepository defines the interface for password reset token operations
type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *models.PasswordResetToken) error
	GetByToken(ctx context.Context, token string) (*models.PasswordResetToken, error)
	MarkAsUsed(ctx context.Context, tokenID string) error
}

// GraphRepository defines the interface for graph data access operations
type GraphRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, graph *models.Graph) error
	GetByID(ctx context.Context, graphID string) (*models.Graph, error)
	GetByZepGraphID(ctx context.Context, zepGraphID string) (*models.Graph, error)
	Update(ctx context.Context, graph *models.Graph) error
	Delete(ctx context.Context, graphID string) error

	// Graph listing with membership join
	ListByUserID(ctx context.Context, userID string) ([]*models.Graph, error)

	// Document count management
	UpdateDocumentCount(ctx context.Context, graphID string, delta int) error

	// Membership operations
	CreateMembership(ctx context.Context, membership *models.GraphMembership) error
	DeleteMembership(ctx context.Context, graphID, userID string) error
	GetMembership(ctx context.Context, graphID, userID string) (*models.GraphMembership, error)
	ListMembersByGraphID(ctx context.Context, graphID string) ([]*models.GraphMembership, error)
	IsMember(ctx context.Context, graphID, userID string) (bool, error)
}
