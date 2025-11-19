package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/bipulkrdas/orgmind/backend/internal/models"
	"github.com/jmoiron/sqlx"
)

type geminiStoreRepository struct {
	db *sqlx.DB
	qb squirrel.StatementBuilderType
}

// NewGeminiStoreRepository creates a new Gemini store repository
func NewGeminiStoreRepository(db *sqlx.DB) GeminiStoreRepository {
	return &geminiStoreRepository{
		db: db,
		qb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Create creates a new Gemini File Search store record
// ID, CreatedAt, and UpdatedAt are set by database defaults
func (r *geminiStoreRepository) Create(ctx context.Context, store *models.GeminiFileSearchStore) error {
	query, args, err := r.qb.
		Insert("gemini_filesearch_stores").
		Columns("store_name", "store_id").
		Values(store.StoreName, store.StoreID).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to create gemini store: %w", err)
	}

	return nil
}

// GetByStoreName retrieves a Gemini File Search store by store name
func (r *geminiStoreRepository) GetByStoreName(ctx context.Context, storeName string) (*models.GeminiFileSearchStore, error) {
	query, args, err := r.qb.
		Select("id", "store_name", "store_id", "created_at", "updated_at").
		From("gemini_filesearch_stores").
		Where(squirrel.Eq{"store_name": storeName}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var store models.GeminiFileSearchStore
	err = r.db.GetContext(ctx, &store, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get gemini store: %w", err)
	}

	return &store, nil
}

// Update updates an existing Gemini File Search store record
func (r *geminiStoreRepository) Update(ctx context.Context, store *models.GeminiFileSearchStore) error {
	query, args, err := r.qb.
		Update("gemini_filesearch_stores").
		Set("store_id", store.StoreID).
		Set("updated_at", store.UpdatedAt).
		Where(squirrel.Eq{"id": store.ID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update gemini store: %w", err)
	}

	return nil
}
