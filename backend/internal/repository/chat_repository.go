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

// chatRepository implements ChatRepository interface
type chatRepository struct {
	db *sqlx.DB
	qb sq.StatementBuilderType
}

// NewChatRepository creates a new instance of ChatRepository
func NewChatRepository(db *sqlx.DB) ChatRepository {
	return &chatRepository{
		db: db,
		qb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// CreateThread inserts a new chat thread into the database
func (r *chatRepository) CreateThread(ctx context.Context, thread *models.ChatThread) error {
	query, args, err := r.qb.
		Insert("chat_threads").
		Columns(
			"id", "graph_id", "user_id", "summary",
			"created_at", "updated_at",
		).
		Values(
			thread.ID, thread.GraphID, thread.UserID, thread.Summary,
			thread.CreatedAt, thread.UpdatedAt,
		).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to create chat thread: %w", err)
	}

	return nil
}

// GetThreadByID retrieves a chat thread by its ID
func (r *chatRepository) GetThreadByID(ctx context.Context, threadID string) (*models.ChatThread, error) {
	query, args, err := r.qb.
		Select(
			"id", "graph_id", "user_id", "summary",
			"created_at", "updated_at",
		).
		From("chat_threads").
		Where(sq.Eq{"id": threadID}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var thread models.ChatThread
	err = r.db.GetContext(ctx, &thread, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("chat thread not found")
		}
		return nil, fmt.Errorf("failed to get chat thread by ID: %w", err)
	}

	return &thread, nil
}

// ListThreadsByGraphID retrieves all chat threads for a specific graph, ordered by most recent activity
func (r *chatRepository) ListThreadsByGraphID(ctx context.Context, graphID string) ([]*models.ChatThread, error) {
	query, args, err := r.qb.
		Select(
			"id", "graph_id", "user_id", "summary",
			"created_at", "updated_at",
		).
		From("chat_threads").
		Where(sq.Eq{"graph_id": graphID}).
		OrderBy("updated_at DESC").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var threads []*models.ChatThread
	err = r.db.SelectContext(ctx, &threads, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list chat threads by graph ID: %w", err)
	}

	return threads, nil
}

// UpdateThread updates an existing chat thread (primarily for summary updates)
func (r *chatRepository) UpdateThread(ctx context.Context, thread *models.ChatThread) error {
	query, args, err := r.qb.
		Update("chat_threads").
		Set("summary", thread.Summary).
		Set("updated_at", thread.UpdatedAt).
		Where(sq.Eq{"id": thread.ID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update chat thread: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("chat thread not found")
	}

	return nil
}

// DeleteThread removes a chat thread from the database (cascade deletes messages)
func (r *chatRepository) DeleteThread(ctx context.Context, threadID string) error {
	query, args, err := r.qb.
		Delete("chat_threads").
		Where(sq.Eq{"id": threadID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete chat thread: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("chat thread not found")
	}

	return nil
}

// CreateMessage inserts a new chat message into the database
func (r *chatRepository) CreateMessage(ctx context.Context, message *models.ChatMessage) error {
	query, args, err := r.qb.
		Insert("chat_messages").
		Columns(
			"id", "thread_id", "role", "content", "created_at",
		).
		Values(
			message.ID, message.ThreadID, message.Role, message.Content, message.CreatedAt,
		).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to create chat message: %w", err)
	}

	return nil
}

// GetMessageByID retrieves a single message by its ID
func (r *chatRepository) GetMessageByID(ctx context.Context, messageID string) (*models.ChatMessage, error) {
	query, args, err := r.qb.
		Select(
			"id", "thread_id", "role", "content", "created_at",
		).
		From("chat_messages").
		Where(sq.Eq{"id": messageID}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var message models.ChatMessage
	err = r.db.GetContext(ctx, &message, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("chat message not found")
		}
		return nil, fmt.Errorf("failed to get chat message by ID: %w", err)
	}

	return &message, nil
}

// GetMessagesByThreadID retrieves all messages for a specific thread with pagination
func (r *chatRepository) GetMessagesByThreadID(ctx context.Context, threadID string, limit, offset int) ([]*models.ChatMessage, error) {
	query, args, err := r.qb.
		Select(
			"id", "thread_id", "role", "content", "created_at",
		).
		From("chat_messages").
		Where(sq.Eq{"thread_id": threadID}).
		OrderBy("created_at ASC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var messages []*models.ChatMessage
	err = r.db.SelectContext(ctx, &messages, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by thread ID: %w", err)
	}

	return messages, nil
}

// DeleteMessagesByThreadID removes all messages for a specific thread
func (r *chatRepository) DeleteMessagesByThreadID(ctx context.Context, threadID string) error {
	query, args, err := r.qb.
		Delete("chat_messages").
		Where(sq.Eq{"thread_id": threadID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete messages by thread ID: %w", err)
	}

	return nil
}
