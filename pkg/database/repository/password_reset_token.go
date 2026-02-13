package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"git.martianoids.com/martianoids/martian-stack/pkg/database"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/adapter"
	"github.com/google/uuid"
)

// SQLPasswordResetTokenRepository implements adapter.PasswordResetTokenRepository using SQL database
type SQLPasswordResetTokenRepository struct {
	db database.Database
}

// NewSQLPasswordResetTokenRepository creates a new SQL-based password reset token repository
func NewSQLPasswordResetTokenRepository(db database.Database) *SQLPasswordResetTokenRepository {
	return &SQLPasswordResetTokenRepository{
		db: db,
	}
}

// Create stores a new password reset token
func (r *SQLPasswordResetTokenRepository) Create(token *adapter.PasswordResetToken) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Generate new UUID if not set
	if token.ID == "" {
		token.ID = uuid.NewString()
	}

	// Set created time if not set
	if token.CreatedAt.IsZero() {
		token.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, created_at, used_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	var usedAt interface{}
	if token.UsedAt != nil {
		usedAt = *token.UsedAt
	}

	_, err := r.db.Exec(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
		usedAt,
	)

	return err
}

// GetByTokenHash retrieves a password reset token by its hash
func (r *SQLPasswordResetTokenRepository) GetByTokenHash(tokenHash string) (*adapter.PasswordResetToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, token_hash, expires_at, created_at, used_at
		FROM password_reset_tokens
		WHERE token_hash = ?
	`

	var token adapter.PasswordResetToken
	var usedAt sql.NullTime

	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
		&usedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, adapter.ErrTokenNotFound
		}
		return nil, err
	}

	if usedAt.Valid {
		token.UsedAt = &usedAt.Time
	}

	return &token, nil
}

// MarkAsUsed marks a password reset token as used
func (r *SQLPasswordResetTokenRepository) MarkAsUsed(tokenHash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		UPDATE password_reset_tokens
		SET used_at = ?
		WHERE token_hash = ? AND used_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, time.Now(), tokenHash)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return adapter.ErrTokenNotFound
	}

	return nil
}

// DeleteExpired removes all expired tokens
func (r *SQLPasswordResetTokenRepository) DeleteExpired() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		DELETE FROM password_reset_tokens
		WHERE expires_at < ?
	`

	_, err := r.db.Exec(ctx, query, time.Now())
	return err
}

// DeleteByUserID removes all password reset tokens for a user
func (r *SQLPasswordResetTokenRepository) DeleteByUserID(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		DELETE FROM password_reset_tokens
		WHERE user_id = ?
	`

	_, err := r.db.Exec(ctx, query, userID)
	return err
}
