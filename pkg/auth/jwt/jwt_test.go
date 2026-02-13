package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	cfg := DefaultConfig("test-secret")
	service := NewService(cfg)

	assert.NotNil(t, service)
	assert.Equal(t, cfg, service.config)
}

func TestGenerateAccessToken(t *testing.T) {
	service := NewService(DefaultConfig("test-secret"))

	token, err := service.GenerateAccessToken("user-123", "johndoe", "john@example.com", "user")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the generated token
	claims, err := service.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "johndoe", claims.Username)
	assert.Equal(t, "john@example.com", claims.Email)
	assert.Equal(t, "user", claims.Role)
}

func TestGenerateRefreshToken(t *testing.T) {
	service := NewService(DefaultConfig("test-secret"))

	token, err := service.GenerateRefreshToken("user-123")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the generated token
	claims, err := service.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	// Refresh tokens should not contain user details
	assert.Empty(t, claims.Username)
	assert.Empty(t, claims.Email)
	assert.Empty(t, claims.Role)
}

func TestValidateToken(t *testing.T) {
	service := NewService(DefaultConfig("test-secret"))

	// Generate a valid token
	token, err := service.GenerateAccessToken("user-123", "johndoe", "john@example.com", "admin")
	require.NoError(t, err)

	// Validate it
	claims, err := service.ValidateToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "johndoe", claims.Username)
	assert.Equal(t, "john@example.com", claims.Email)
	assert.Equal(t, "admin", claims.Role)
}

func TestValidateToken_Invalid(t *testing.T) {
	service := NewService(DefaultConfig("test-secret"))

	tests := []struct {
		name  string
		token string
		err   error
	}{
		{
			name:  "empty token",
			token: "",
			err:   ErrInvalidToken,
		},
		{
			name:  "malformed token",
			token: "not.a.valid.token",
			err:   ErrInvalidToken,
		},
		{
			name:  "invalid signature",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			err:   ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateToken(tt.token)
			assert.Error(t, err)
			assert.Nil(t, claims)
			assert.ErrorIs(t, err, tt.err)
		})
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	service1 := NewService(DefaultConfig("secret-1"))
	service2 := NewService(DefaultConfig("secret-2"))

	// Generate token with service1
	token, err := service1.GenerateAccessToken("user-123", "johndoe", "john@example.com", "user")
	require.NoError(t, err)

	// Try to validate with service2 (different secret)
	claims, err := service2.ValidateToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestValidateToken_Expired(t *testing.T) {
	cfg := DefaultConfig("test-secret")
	cfg.AccessTokenExpiry = 1 * time.Millisecond // Very short expiry
	service := NewService(cfg)

	token, err := service.GenerateAccessToken("user-123", "johndoe", "john@example.com", "user")
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	claims, err := service.ValidateToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestGetExpiryTime(t *testing.T) {
	service := NewService(DefaultConfig("test-secret"))

	token, err := service.GenerateAccessToken("user-123", "johndoe", "john@example.com", "user")
	require.NoError(t, err)

	expiryTime, err := service.GetExpiryTime(token)
	assert.NoError(t, err)
	assert.False(t, expiryTime.IsZero())
	assert.True(t, expiryTime.After(time.Now()))
}

func TestGetExpiryTime_InvalidToken(t *testing.T) {
	service := NewService(DefaultConfig("test-secret"))

	expiryTime, err := service.GetExpiryTime("invalid-token")
	assert.Error(t, err)
	assert.True(t, expiryTime.IsZero())
}

func TestIsExpired(t *testing.T) {
	service := NewService(DefaultConfig("test-secret"))

	// Valid token (not expired)
	token, err := service.GenerateAccessToken("user-123", "johndoe", "john@example.com", "user")
	require.NoError(t, err)

	isExpired := service.IsExpired(token)
	assert.False(t, isExpired)

	// Expired token
	cfg := DefaultConfig("test-secret")
	cfg.AccessTokenExpiry = 1 * time.Millisecond
	expiredService := NewService(cfg)

	expiredToken, err := expiredService.GenerateAccessToken("user-123", "johndoe", "john@example.com", "user")
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	isExpired = expiredService.IsExpired(expiredToken)
	assert.True(t, isExpired)
}

func TestIsExpired_InvalidToken(t *testing.T) {
	service := NewService(DefaultConfig("test-secret"))

	isExpired := service.IsExpired("invalid-token")
	assert.True(t, isExpired)
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("my-secret")

	assert.Equal(t, []byte("my-secret"), cfg.SecretKey)
	assert.Equal(t, "martian-stack", cfg.Issuer)
	assert.Equal(t, 15*time.Minute, cfg.AccessTokenExpiry)
	assert.Equal(t, 7*24*time.Hour, cfg.RefreshTokenExpiry)
}

func TestTokenClaims(t *testing.T) {
	service := NewService(DefaultConfig("test-secret"))

	token, err := service.GenerateAccessToken("user-123", "johndoe", "john@example.com", "admin")
	require.NoError(t, err)

	claims, err := service.ValidateToken(token)
	require.NoError(t, err)

	// Verify all claims
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "user-123", claims.Subject)
	assert.Equal(t, "johndoe", claims.Username)
	assert.Equal(t, "john@example.com", claims.Email)
	assert.Equal(t, "admin", claims.Role)
	assert.Equal(t, "martian-stack", claims.Issuer)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.NotBefore)
}

func TestTokenRotation(t *testing.T) {
	service := NewService(DefaultConfig("test-secret"))

	// Generate multiple tokens for same user
	token1, err := service.GenerateAccessToken("user-123", "johndoe", "john@example.com", "user")
	require.NoError(t, err)

	time.Sleep(1 * time.Second) // Ensure different IssuedAt timestamp

	token2, err := service.GenerateAccessToken("user-123", "johndoe", "john@example.com", "user")
	require.NoError(t, err)

	// Tokens should be different due to different IssuedAt
	assert.NotEqual(t, token1, token2)

	// But both should be valid
	claims1, err := service.ValidateToken(token1)
	assert.NoError(t, err)
	assert.Equal(t, "user-123", claims1.UserID)

	claims2, err := service.ValidateToken(token2)
	assert.NoError(t, err)
	assert.Equal(t, "user-123", claims2.UserID)

	// IssuedAt should be different
	assert.True(t, claims2.IssuedAt.Time.After(claims1.IssuedAt.Time))
}

func TestRefreshTokenExpiry(t *testing.T) {
	cfg := DefaultConfig("test-secret")
	cfg.RefreshTokenExpiry = 1 * time.Millisecond
	service := NewService(cfg)

	token, err := service.GenerateRefreshToken("user-123")
	require.NoError(t, err)

	// Wait for expiry
	time.Sleep(10 * time.Millisecond)

	claims, err := service.ValidateToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, ErrExpiredToken)
}
