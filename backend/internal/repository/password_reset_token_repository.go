package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bipulkrdas/orgmind/backend/internal/models"
	"github.com/jmoiron/sqlx"
)

// passwordResetTokenRepository implements PasswordResetTokenRepository interface
type passwordResetTokenRepository struct {
	db *sqlx.DB
}

// NewPasswordResetTokenRepository creates a new instance of PasswordResetTokenRepository
func NewPasswordResetTokenRepository(db *sqlx.DB) PasswordResetTokenRepository {
	return &passwordResetTokenRepository{db: db}
}

// Create inserts a new password reset token into the database
func (r *passwordResetTokenRepository) Create(ctx context.Context, token *models.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (
			id, user_id, token, expires_at, used, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		token.ID,
		token.UserID,
		token.Token,
		token.ExpiresAt,
		token.Used,
		token.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create password reset token: %w", err)
	}

	return nil
}

// GetByToken retrieves a password reset token by its token string
func (r *passwordResetTokenRepository) GetByToken(ctx context.Context, tokenStr string) (*models.PasswordResetToken, error) {
	query := `
		SELECT 
			id, user_id, token, expires_at, used, created_at
		FROM password_reset_tokens
		WHERE token = $1
	`

	var token models.PasswordResetToken
	err := r.db.GetContext(ctx, &token, query, tokenStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("token not found")
		}
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	return &token, nil
}

// MarkAsUsed marks a password reset token as used
func (r *passwordResetTokenRepository) MarkAsUsed(ctx context.Context, tokenID string) error {
	query := `
		UPDATE password_reset_tokens
		SET used = true
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("token not found")
	}

	return nil
}
