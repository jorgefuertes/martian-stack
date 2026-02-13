package adapter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSecureToken(t *testing.T) {
	rawToken, tokenHash, err := GenerateSecureToken()
	require.NoError(t, err)
	assert.NotEmpty(t, rawToken, "Raw token should not be empty")
	assert.NotEmpty(t, tokenHash, "Token hash should not be empty")
	assert.NotEqual(t, rawToken, tokenHash, "Raw token and hash should be different")
	assert.Len(t, rawToken, 64, "Raw token should be 64 characters (32 bytes in hex)")
	assert.Len(t, tokenHash, 64, "Token hash should be 64 characters (SHA-256 in hex)")
}

func TestGenerateSecureToken_Uniqueness(t *testing.T) {
	// Generate multiple tokens and ensure they're all unique
	tokens := make(map[string]bool)
	hashes := make(map[string]bool)

	for i := 0; i < 100; i++ {
		rawToken, tokenHash, err := GenerateSecureToken()
		require.NoError(t, err)

		assert.False(t, tokens[rawToken], "Token should be unique")
		assert.False(t, hashes[tokenHash], "Hash should be unique")

		tokens[rawToken] = true
		hashes[tokenHash] = true
	}

	assert.Len(t, tokens, 100, "Should have 100 unique tokens")
	assert.Len(t, hashes, 100, "Should have 100 unique hashes")
}

func TestHashToken(t *testing.T) {
	rawToken, originalHash, err := GenerateSecureToken()
	require.NoError(t, err)

	// Hash the raw token and verify it matches the original hash
	computedHash, err := HashToken(rawToken)
	require.NoError(t, err)
	assert.Equal(t, originalHash, computedHash, "Hashing the raw token should produce the same hash")
}

func TestHashToken_InvalidToken(t *testing.T) {
	// Test with invalid hex string
	_, err := HashToken("not-a-valid-hex-string")
	assert.Error(t, err, "Should error on invalid hex string")
}

func TestHashToken_Consistency(t *testing.T) {
	rawToken, _, err := GenerateSecureToken()
	require.NoError(t, err)

	// Hash the same token multiple times
	hash1, err := HashToken(rawToken)
	require.NoError(t, err)

	hash2, err := HashToken(rawToken)
	require.NoError(t, err)

	hash3, err := HashToken(rawToken)
	require.NoError(t, err)

	// All hashes should be identical
	assert.Equal(t, hash1, hash2)
	assert.Equal(t, hash2, hash3)
}

func TestRefreshToken_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresIn time.Duration
		want      bool
	}{
		{
			name:      "not expired",
			expiresIn: 1 * time.Hour,
			want:      false,
		},
		{
			name:      "expired",
			expiresIn: -1 * time.Hour,
			want:      true,
		},
		{
			name:      "expires in future",
			expiresIn: 24 * time.Hour,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := NewRefreshToken("user-id", "token-hash", tt.expiresIn)
			assert.Equal(t, tt.want, token.IsExpired())
		})
	}
}

func TestRefreshToken_IsRevoked(t *testing.T) {
	token := NewRefreshToken("user-id", "token-hash", 1*time.Hour)

	// Not revoked initially
	assert.False(t, token.IsRevoked())

	// Revoke token
	now := time.Now()
	token.RevokedAt = &now

	// Should be revoked now
	assert.True(t, token.IsRevoked())
}

func TestRefreshToken_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *RefreshToken
		wantValid bool
	}{
		{
			name: "valid token",
			setup: func() *RefreshToken {
				return NewRefreshToken("user-id", "token-hash", 1*time.Hour)
			},
			wantValid: true,
		},
		{
			name: "expired token",
			setup: func() *RefreshToken {
				return NewRefreshToken("user-id", "token-hash", -1*time.Hour)
			},
			wantValid: false,
		},
		{
			name: "revoked token",
			setup: func() *RefreshToken {
				token := NewRefreshToken("user-id", "token-hash", 1*time.Hour)
				now := time.Now()
				token.RevokedAt = &now
				return token
			},
			wantValid: false,
		},
		{
			name: "expired and revoked",
			setup: func() *RefreshToken {
				token := NewRefreshToken("user-id", "token-hash", -1*time.Hour)
				now := time.Now()
				token.RevokedAt = &now
				return token
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setup()
			assert.Equal(t, tt.wantValid, token.IsValid())
		})
	}
}

func TestPasswordResetToken_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresIn time.Duration
		want      bool
	}{
		{
			name:      "not expired",
			expiresIn: 1 * time.Hour,
			want:      false,
		},
		{
			name:      "expired",
			expiresIn: -1 * time.Hour,
			want:      true,
		},
		{
			name:      "expires in future",
			expiresIn: 24 * time.Hour,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := NewPasswordResetToken("user-id", "token-hash", tt.expiresIn)
			assert.Equal(t, tt.want, token.IsExpired())
		})
	}
}

func TestPasswordResetToken_IsUsed(t *testing.T) {
	token := NewPasswordResetToken("user-id", "token-hash", 1*time.Hour)

	// Not used initially
	assert.False(t, token.IsUsed())

	// Mark as used
	now := time.Now()
	token.UsedAt = &now

	// Should be used now
	assert.True(t, token.IsUsed())
}

func TestPasswordResetToken_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *PasswordResetToken
		wantValid bool
	}{
		{
			name: "valid token",
			setup: func() *PasswordResetToken {
				return NewPasswordResetToken("user-id", "token-hash", 1*time.Hour)
			},
			wantValid: true,
		},
		{
			name: "expired token",
			setup: func() *PasswordResetToken {
				return NewPasswordResetToken("user-id", "token-hash", -1*time.Hour)
			},
			wantValid: false,
		},
		{
			name: "used token",
			setup: func() *PasswordResetToken {
				token := NewPasswordResetToken("user-id", "token-hash", 1*time.Hour)
				now := time.Now()
				token.UsedAt = &now
				return token
			},
			wantValid: false,
		},
		{
			name: "expired and used",
			setup: func() *PasswordResetToken {
				token := NewPasswordResetToken("user-id", "token-hash", -1*time.Hour)
				now := time.Now()
				token.UsedAt = &now
				return token
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setup()
			assert.Equal(t, tt.wantValid, token.IsValid())
		})
	}
}

func TestNewRefreshToken(t *testing.T) {
	userID := "user-123"
	tokenHash := "hash-abc"
	expiresIn := 7 * 24 * time.Hour

	token := NewRefreshToken(userID, tokenHash, expiresIn)

	assert.Equal(t, userID, token.UserID)
	assert.Equal(t, tokenHash, token.TokenHash)
	assert.False(t, token.CreatedAt.IsZero())
	assert.False(t, token.ExpiresAt.IsZero())
	assert.True(t, token.ExpiresAt.After(token.CreatedAt))
	assert.WithinDuration(t, time.Now().Add(expiresIn), token.ExpiresAt, 1*time.Second)
}

func TestNewPasswordResetToken(t *testing.T) {
	userID := "user-123"
	tokenHash := "hash-abc"
	expiresIn := 1 * time.Hour

	token := NewPasswordResetToken(userID, tokenHash, expiresIn)

	assert.Equal(t, userID, token.UserID)
	assert.Equal(t, tokenHash, token.TokenHash)
	assert.False(t, token.CreatedAt.IsZero())
	assert.False(t, token.ExpiresAt.IsZero())
	assert.True(t, token.ExpiresAt.After(token.CreatedAt))
	assert.WithinDuration(t, time.Now().Add(expiresIn), token.ExpiresAt, 1*time.Second)
}
