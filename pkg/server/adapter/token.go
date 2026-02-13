package adapter

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

var (
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenExpired  = errors.New("token expired")
	ErrTokenRevoked  = errors.New("token revoked")
	ErrTokenUsed     = errors.New("token already used")
)

// RefreshToken represents a refresh token in the database
type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}

// IsExpired checks if the refresh token has expired
func (t *RefreshToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsRevoked checks if the refresh token has been revoked
func (t *RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

// IsValid checks if the token is valid (not expired, not revoked)
func (t *RefreshToken) IsValid() bool {
	return !t.IsExpired() && !t.IsRevoked()
}

// PasswordResetToken represents a password reset token in the database
type PasswordResetToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	UsedAt    *time.Time
}

// IsExpired checks if the password reset token has expired
func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsUsed checks if the password reset token has been used
func (t *PasswordResetToken) IsUsed() bool {
	return t.UsedAt != nil
}

// IsValid checks if the token is valid (not expired, not used)
func (t *PasswordResetToken) IsValid() bool {
	return !t.IsExpired() && !t.IsUsed()
}

// GenerateSecureToken generates a cryptographically secure random token
// Returns the raw token (to send to user) and the hashed token (to store in DB)
func GenerateSecureToken() (string, string, error) {
	// Generate 32 random bytes (256 bits)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}

	// Convert to hex string (this is what we send to the user)
	rawToken := hex.EncodeToString(tokenBytes)

	// Hash the token (this is what we store in the database)
	hash := sha256.Sum256(tokenBytes)
	tokenHash := hex.EncodeToString(hash[:])

	return rawToken, tokenHash, nil
}

// HashToken hashes a raw token for database storage
func HashToken(rawToken string) (string, error) {
	tokenBytes, err := hex.DecodeString(rawToken)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(tokenBytes)
	return hex.EncodeToString(hash[:]), nil
}
