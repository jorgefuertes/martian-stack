package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jorgefuertes/martian-stack/pkg/database"
	"github.com/jorgefuertes/martian-stack/pkg/server/adapter"
)

// SQLRefreshTokenRepository implements adapter.RefreshTokenRepository using SQL database
type SQLRefreshTokenRepository struct {
	db database.Database
}

// NewSQLRefreshTokenRepository creates a new SQL-based refresh token repository
func NewSQLRefreshTokenRepository(db database.Database) *SQLRefreshTokenRepository {
	return &SQLRefreshTokenRepository{
		db: db,
	}
}

// Create stores a new refresh token
func (r *SQLRefreshTokenRepository) Create(token *adapter.RefreshToken) error {
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
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at, revoked_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	var revokedAt interface{}
	if token.RevokedAt != nil {
		revokedAt = *token.RevokedAt
	}

	_, err := r.db.Exec(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
		revokedAt,
	)

	return err
}

// GetByTokenHash retrieves a refresh token by its hash
func (r *SQLRefreshTokenRepository) GetByTokenHash(tokenHash string) (*adapter.RefreshToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, token_hash, expires_at, created_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = ?
	`

	var token adapter.RefreshToken
	var revokedAt sql.NullTime

	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
		&revokedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, adapter.ErrTokenNotFound
		}
		return nil, err
	}

	if revokedAt.Valid {
		token.RevokedAt = &revokedAt.Time
	}

	return &token, nil
}

// GetByUserID retrieves all refresh tokens for a user
func (r *SQLRefreshTokenRepository) GetByUserID(userID string) ([]*adapter.RefreshToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, token_hash, expires_at, created_at, revoked_at
		FROM refresh_tokens
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*adapter.RefreshToken
	for rows.Next() {
		var token adapter.RefreshToken
		var revokedAt sql.NullTime

		err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.TokenHash,
			&token.ExpiresAt,
			&token.CreatedAt,
			&revokedAt,
		)
		if err != nil {
			return nil, err
		}

		if revokedAt.Valid {
			token.RevokedAt = &revokedAt.Time
		}

		tokens = append(tokens, &token)
	}

	return tokens, rows.Err()
}

// Revoke marks a refresh token as revoked
func (r *SQLRefreshTokenRepository) Revoke(tokenHash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		UPDATE refresh_tokens
		SET revoked_at = ?
		WHERE token_hash = ? AND revoked_at IS NULL
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

// RevokeAll revokes all refresh tokens for a user
func (r *SQLRefreshTokenRepository) RevokeAll(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		UPDATE refresh_tokens
		SET revoked_at = ?
		WHERE user_id = ? AND revoked_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, time.Now(), userID)
	return err
}

// DeleteExpired removes all expired tokens
func (r *SQLRefreshTokenRepository) DeleteExpired() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < ?
	`

	_, err := r.db.Exec(ctx, query, time.Now())
	return err
}

// Delete removes a specific token
func (r *SQLRefreshTokenRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `DELETE FROM refresh_tokens WHERE id = ?`

	result, err := r.db.Exec(ctx, query, id)
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
