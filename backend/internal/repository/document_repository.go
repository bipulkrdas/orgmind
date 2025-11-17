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

// documentRepository implements DocumentRepository interface
type documentRepository struct {
	db *sqlx.DB
	qb sq.StatementBuilderType
}

// NewDocumentRepository creates a new instance of DocumentRepository
func NewDocumentRepository(db *sqlx.DB) DocumentRepository {
	return &documentRepository{
		db: db,
		qb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// Create inserts a new document into the database
func (r *documentRepository) Create(ctx context.Context, doc *models.Document) error {
	query, args, err := r.qb.
		Insert("documents").
		Columns(
			"id", "user_id", "graph_id", "filename", "content_type", "storage_key",
			"size_bytes", "source", "status",
			"created_at", "updated_at",
		).
		Values(
			doc.ID, doc.UserID, doc.GraphID, doc.Filename, doc.ContentType, doc.StorageKey,
			doc.SizeBytes, doc.Source, doc.Status,
			doc.CreatedAt, doc.UpdatedAt,
		).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	return nil
}

// GetByID retrieves a document by its ID
func (r *documentRepository) GetByID(ctx context.Context, docID string) (*models.Document, error) {
	query, args, err := r.qb.
		Select(
			"id", "user_id", "graph_id", "filename", "content_type", "storage_key",
			"size_bytes", "source", "status",
			"created_at", "updated_at",
		).
		From("documents").
		Where(sq.Eq{"id": docID}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var doc models.Document
	err = r.db.GetContext(ctx, &doc, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("document not found")
		}
		return nil, fmt.Errorf("failed to get document by ID: %w", err)
	}

	return &doc, nil
}

// ListByUserID retrieves all documents for a specific user
func (r *documentRepository) ListByUserID(ctx context.Context, userID string) ([]*models.Document, error) {
	query, args, err := r.qb.
		Select(
			"id", "user_id", "graph_id", "filename", "content_type", "storage_key",
			"size_bytes", "source", "status",
			"created_at", "updated_at",
		).
		From("documents").
		Where(sq.Eq{"user_id": userID}).
		OrderBy("created_at DESC").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var docs []*models.Document
	err = r.db.SelectContext(ctx, &docs, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents by user ID: %w", err)
	}

	return docs, nil
}

// ListByGraphID retrieves all documents for a specific graph
func (r *documentRepository) ListByGraphID(ctx context.Context, graphID string) ([]*models.Document, error) {
	query, args, err := r.qb.
		Select(
			"id", "user_id", "graph_id", "filename", "content_type", "storage_key",
			"size_bytes", "source", "status",
			"created_at", "updated_at",
		).
		From("documents").
		Where(sq.Eq{"graph_id": graphID}).
		OrderBy("created_at DESC").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var docs []*models.Document
	err = r.db.SelectContext(ctx, &docs, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents by graph ID: %w", err)
	}

	return docs, nil
}

// Update updates an existing document in the database
func (r *documentRepository) Update(ctx context.Context, doc *models.Document) error {
	query, args, err := r.qb.
		Update("documents").
		Set("graph_id", doc.GraphID).
		Set("filename", doc.Filename).
		Set("content_type", doc.ContentType).
		Set("storage_key", doc.StorageKey).
		Set("size_bytes", doc.SizeBytes).
		Set("source", doc.Source).
		Set("status", doc.Status).
		Set("updated_at", doc.UpdatedAt).
		Where(sq.Eq{"id": doc.ID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("document not found")
	}

	return nil
}

// Delete removes a document from the database
func (r *documentRepository) Delete(ctx context.Context, docID string) error {
	query, args, err := r.qb.
		Delete("documents").
		Where(sq.Eq{"id": docID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("document not found")
	}

	return nil
}
