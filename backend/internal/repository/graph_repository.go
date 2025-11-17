package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bipulkrdas/orgmind/backend/internal/models"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

// graphRepository implements GraphRepository interface
type graphRepository struct {
	db *sqlx.DB
	qb sq.StatementBuilderType
}

// NewGraphRepository creates a new instance of GraphRepository
func NewGraphRepository(db *sqlx.DB) GraphRepository {
	return &graphRepository{
		db: db,
		qb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// Create inserts a new graph record into the database
func (r *graphRepository) Create(ctx context.Context, graph *models.Graph) error {
	query, args, err := r.qb.
		Insert("graphs").
		Columns(
			"id", "creator_id", "zep_graph_id", "name", "description",
			"document_count", "created_at", "updated_at",
		).
		Values(
			graph.ID, graph.CreatorID, graph.ZepGraphID, graph.Name, graph.Description,
			graph.DocumentCount, graph.CreatedAt, graph.UpdatedAt,
		).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to create graph: %w", err)
	}

	return nil
}

// GetByID retrieves a graph by its ID
func (r *graphRepository) GetByID(ctx context.Context, graphID string) (*models.Graph, error) {
	query, args, err := r.qb.
		Select(
			"id", "creator_id", "zep_graph_id", "name", "description",
			"document_count", "created_at", "updated_at",
		).
		From("graphs").
		Where(sq.Eq{"id": graphID}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var graph models.Graph
	err = r.db.GetContext(ctx, &graph, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("graph not found")
		}
		return nil, fmt.Errorf("failed to get graph by ID: %w", err)
	}

	return &graph, nil
}

// GetByZepGraphID retrieves a graph by its Zep graph ID
func (r *graphRepository) GetByZepGraphID(ctx context.Context, zepGraphID string) (*models.Graph, error) {
	query, args, err := r.qb.
		Select(
			"id", "creator_id", "zep_graph_id", "name", "description",
			"document_count", "created_at", "updated_at",
		).
		From("graphs").
		Where(sq.Eq{"zep_graph_id": zepGraphID}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var graph models.Graph
	err = r.db.GetContext(ctx, &graph, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("graph not found")
		}
		return nil, fmt.Errorf("failed to get graph by Zep graph ID: %w", err)
	}

	return &graph, nil
}

// Update updates an existing graph's name and description
func (r *graphRepository) Update(ctx context.Context, graph *models.Graph) error {
	query, args, err := r.qb.
		Update("graphs").
		Set("name", graph.Name).
		Set("description", graph.Description).
		Set("updated_at", graph.UpdatedAt).
		Where(sq.Eq{"id": graph.ID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update graph: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("graph not found")
	}

	return nil
}

// Delete removes a graph from the database (cascade deletes memberships and documents)
func (r *graphRepository) Delete(ctx context.Context, graphID string) error {
	query, args, err := r.qb.
		Delete("graphs").
		Where(sq.Eq{"id": graphID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete graph: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("graph not found")
	}

	return nil
}

// CreateMembership inserts a new membership record
func (r *graphRepository) CreateMembership(ctx context.Context, membership *models.GraphMembership) error {
	query, args, err := r.qb.
		Insert("graph_memberships").
		Columns(
			"id", "graph_id", "user_id", "role", "created_at",
		).
		Values(
			membership.ID, membership.GraphID, membership.UserID, membership.Role, membership.CreatedAt,
		).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to create membership: %w", err)
	}

	return nil
}

// DeleteMembership removes a member from a graph
func (r *graphRepository) DeleteMembership(ctx context.Context, graphID, userID string) error {
	query, args, err := r.qb.
		Delete("graph_memberships").
		Where(sq.Eq{"graph_id": graphID, "user_id": userID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete membership: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("membership not found")
	}

	return nil
}

// GetMembership fetches a membership by graph_id and user_id
func (r *graphRepository) GetMembership(ctx context.Context, graphID, userID string) (*models.GraphMembership, error) {
	query, args, err := r.qb.
		Select(
			"id", "graph_id", "user_id", "role", "created_at",
		).
		From("graph_memberships").
		Where(sq.Eq{"graph_id": graphID, "user_id": userID}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var membership models.GraphMembership
	err = r.db.GetContext(ctx, &membership, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("membership not found")
		}
		return nil, fmt.Errorf("failed to get membership: %w", err)
	}

	return &membership, nil
}

// ListMembersByGraphID gets all members of a graph
func (r *graphRepository) ListMembersByGraphID(ctx context.Context, graphID string) ([]*models.GraphMembership, error) {
	query, args, err := r.qb.
		Select(
			"id", "graph_id", "user_id", "role", "created_at",
		).
		From("graph_memberships").
		Where(sq.Eq{"graph_id": graphID}).
		OrderBy("created_at ASC").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var memberships []*models.GraphMembership
	err = r.db.SelectContext(ctx, &memberships, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list members by graph ID: %w", err)
	}

	return memberships, nil
}

// IsMember checks if a user is a member of a graph
func (r *graphRepository) IsMember(ctx context.Context, graphID, userID string) (bool, error) {
	query, args, err := r.qb.
		Select("COUNT(*)").
		From("graph_memberships").
		Where(sq.Eq{"graph_id": graphID, "user_id": userID}).
		ToSql()

	if err != nil {
		return false, fmt.Errorf("failed to build select query: %w", err)
	}

	var count int
	err = r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return false, fmt.Errorf("failed to check membership: %w", err)
	}

	return count > 0, nil
}

// ListByUserID returns all graphs where the user is a member (via graph_memberships join)
func (r *graphRepository) ListByUserID(ctx context.Context, userID string) ([]*models.Graph, error) {
	query, args, err := r.qb.
		Select(
			"g.id", "g.creator_id", "g.zep_graph_id", "g.name", "g.description",
			"g.document_count", "g.created_at", "g.updated_at",
		).
		From("graphs g").
		Join("graph_memberships gm ON g.id = gm.graph_id").
		Where(sq.Eq{"gm.user_id": userID}).
		OrderBy("g.created_at DESC").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var graphs []*models.Graph
	err = r.db.SelectContext(ctx, &graphs, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list graphs by user ID: %w", err)
	}

	return graphs, nil
}

// UpdateDocumentCount atomically increments or decrements the document count
func (r *graphRepository) UpdateDocumentCount(ctx context.Context, graphID string, delta int) error {
	// Use raw SQL for atomic UPDATE with delta
	query := `
		UPDATE graphs
		SET document_count = document_count + $1,
		    updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, delta, graphID)
	if err != nil {
		return fmt.Errorf("failed to update document count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("graph not found")
	}

	return nil
}
