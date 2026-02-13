package adapter

import "time"

// RefreshTokenRepository defines the interface for refresh token operations
type RefreshTokenRepository interface {
	// Create stores a new refresh token
	Create(token *RefreshToken) error

	// GetByTokenHash retrieves a refresh token by its hash
	GetByTokenHash(tokenHash string) (*RefreshToken, error)

	// GetByUserID retrieves all refresh tokens for a user
	GetByUserID(userID string) ([]*RefreshToken, error)

	// Revoke marks a refresh token as revoked
	Revoke(tokenHash string) error

	// RevokeAll revokes all refresh tokens for a user
	RevokeAll(userID string) error

	// DeleteExpired removes all expired tokens
	DeleteExpired() error

	// Delete removes a specific token
	Delete(id string) error
}

// PasswordResetTokenRepository defines the interface for password reset token operations
type PasswordResetTokenRepository interface {
	// Create stores a new password reset token
	Create(token *PasswordResetToken) error

	// GetByTokenHash retrieves a password reset token by its hash
	GetByTokenHash(tokenHash string) (*PasswordResetToken, error)

	// MarkAsUsed marks a password reset token as used
	MarkAsUsed(tokenHash string) error

	// DeleteExpired removes all expired tokens
	DeleteExpired() error

	// DeleteByUserID removes all password reset tokens for a user
	DeleteByUserID(userID string) error
}

// TokenRepositories combines both token repositories
type TokenRepositories struct {
	RefreshToken      RefreshTokenRepository
	PasswordResetToken PasswordResetTokenRepository
}

// NewRefreshToken creates a new RefreshToken with the given parameters
func NewRefreshToken(userID string, tokenHash string, expiresIn time.Duration) *RefreshToken {
	now := time.Now()
	return &RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(expiresIn),
		CreatedAt: now,
	}
}

// NewPasswordResetToken creates a new PasswordResetToken with the given parameters
func NewPasswordResetToken(userID string, tokenHash string, expiresIn time.Duration) *PasswordResetToken {
	now := time.Now()
	return &PasswordResetToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(expiresIn),
		CreatedAt: now,
	}
}
