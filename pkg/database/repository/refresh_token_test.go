package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jorgefuertes/martian-stack/pkg/database/migration"
	"github.com/jorgefuertes/martian-stack/pkg/database/migration/migrations"
	"github.com/jorgefuertes/martian-stack/pkg/database/sqlite"
	"github.com/jorgefuertes/martian-stack/pkg/server/adapter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTokenTestDB(t *testing.T) (*SQLRefreshTokenRepository, *SQLAccountRepository) {
	db, err := sqlite.NewInMemory()
	require.NoError(t, err)

	// Run migrations
	migrator := migration.New(db)
	migrator.RegisterMultiple(migrations.All())
	err = migrator.Up(context.Background())
	require.NoError(t, err)

	return NewSQLRefreshTokenRepository(db), NewSQLAccountRepository(db)
}

func createTestUser(t *testing.T, repo *SQLAccountRepository) *adapter.Account {
	acc := &adapter.Account{
		Username: "testuser",
		Name:     "Test User",
		Email:    "test@example.com",
		Enabled:  true,
		Role:     "user",
	}

	err := acc.SetPassword("password123")
	require.NoError(t, err)

	err = repo.Create(acc)
	require.NoError(t, err)

	return acc
}

func TestRefreshTokenRepository_Create(t *testing.T) {
	tokenRepo, accountRepo := setupTokenTestDB(t)
	user := createTestUser(t, accountRepo)

	_, tokenHash, err := adapter.GenerateSecureToken()
	require.NoError(t, err)
	require.NotEmpty(t, tokenHash)

	token := adapter.NewRefreshToken(user.ID, tokenHash, 7*24*time.Hour)

	err = tokenRepo.Create(token)
	assert.NoError(t, err)
	assert.NotEmpty(t, token.ID, "ID should be generated")
	assert.False(t, token.CreatedAt.IsZero(), "CreatedAt should be set")
}

func TestRefreshTokenRepository_GetByTokenHash(t *testing.T) {
	tokenRepo, accountRepo := setupTokenTestDB(t)
	user := createTestUser(t, accountRepo)

	_, tokenHash, err := adapter.GenerateSecureToken()
	require.NoError(t, err)

	token := adapter.NewRefreshToken(user.ID, tokenHash, 7*24*time.Hour)
	err = tokenRepo.Create(token)
	require.NoError(t, err)

	// Get by token hash
	retrieved, err := tokenRepo.GetByTokenHash(tokenHash)
	assert.NoError(t, err)
	assert.Equal(t, token.ID, retrieved.ID)
	assert.Equal(t, token.UserID, retrieved.UserID)
	assert.Equal(t, token.TokenHash, retrieved.TokenHash)
	assert.True(t, retrieved.IsValid())
}

func TestRefreshTokenRepository_GetByTokenHash_NotFound(t *testing.T) {
	tokenRepo, _ := setupTokenTestDB(t)

	retrieved, err := tokenRepo.GetByTokenHash("nonexistent")
	assert.Error(t, err)
	assert.Equal(t, adapter.ErrTokenNotFound, err)
	assert.Nil(t, retrieved)
}

func TestRefreshTokenRepository_GetByUserID(t *testing.T) {
	tokenRepo, accountRepo := setupTokenTestDB(t)
	user := createTestUser(t, accountRepo)

	// Create multiple tokens for the same user
	for i := 0; i < 3; i++ {
		_, tokenHash, err := adapter.GenerateSecureToken()
		require.NoError(t, err)

		token := adapter.NewRefreshToken(user.ID, tokenHash, 7*24*time.Hour)
		err = tokenRepo.Create(token)
		require.NoError(t, err)
	}

	// Get all tokens for user
	tokens, err := tokenRepo.GetByUserID(user.ID)
	assert.NoError(t, err)
	assert.Len(t, tokens, 3)
}

func TestRefreshTokenRepository_Revoke(t *testing.T) {
	tokenRepo, accountRepo := setupTokenTestDB(t)
	user := createTestUser(t, accountRepo)

	_, tokenHash, err := adapter.GenerateSecureToken()
	require.NoError(t, err)

	token := adapter.NewRefreshToken(user.ID, tokenHash, 7*24*time.Hour)
	err = tokenRepo.Create(token)
	require.NoError(t, err)

	// Revoke token
	err = tokenRepo.Revoke(tokenHash)
	assert.NoError(t, err)

	// Verify token is revoked
	retrieved, err := tokenRepo.GetByTokenHash(tokenHash)
	assert.NoError(t, err)
	assert.True(t, retrieved.IsRevoked())
	assert.False(t, retrieved.IsValid())
}

func TestRefreshTokenRepository_RevokeAll(t *testing.T) {
	tokenRepo, accountRepo := setupTokenTestDB(t)
	user := createTestUser(t, accountRepo)

	// Create multiple tokens
	var tokenHashes []string
	for i := 0; i < 3; i++ {
		_, tokenHash, err := adapter.GenerateSecureToken()
		require.NoError(t, err)
		tokenHashes = append(tokenHashes, tokenHash)

		token := adapter.NewRefreshToken(user.ID, tokenHash, 7*24*time.Hour)
		err = tokenRepo.Create(token)
		require.NoError(t, err)
	}

	// Revoke all tokens
	err := tokenRepo.RevokeAll(user.ID)
	assert.NoError(t, err)

	// Verify all tokens are revoked
	for _, tokenHash := range tokenHashes {
		retrieved, err := tokenRepo.GetByTokenHash(tokenHash)
		assert.NoError(t, err)
		assert.True(t, retrieved.IsRevoked())
	}
}

func TestRefreshTokenRepository_DeleteExpired(t *testing.T) {
	tokenRepo, accountRepo := setupTokenTestDB(t)
	user := createTestUser(t, accountRepo)

	// Create an expired token
	_, expiredHash, err := adapter.GenerateSecureToken()
	require.NoError(t, err)

	expiredToken := adapter.NewRefreshToken(user.ID, expiredHash, -1*time.Hour) // Already expired
	err = tokenRepo.Create(expiredToken)
	require.NoError(t, err)

	// Create a valid token
	_, validHash, err := adapter.GenerateSecureToken()
	require.NoError(t, err)

	validToken := adapter.NewRefreshToken(user.ID, validHash, 7*24*time.Hour)
	err = tokenRepo.Create(validToken)
	require.NoError(t, err)

	// Delete expired tokens
	err = tokenRepo.DeleteExpired()
	assert.NoError(t, err)

	// Verify expired token is deleted
	_, err = tokenRepo.GetByTokenHash(expiredHash)
	assert.Error(t, err)
	assert.Equal(t, adapter.ErrTokenNotFound, err)

	// Verify valid token still exists
	retrieved, err := tokenRepo.GetByTokenHash(validHash)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
}

func TestRefreshTokenRepository_Delete(t *testing.T) {
	tokenRepo, accountRepo := setupTokenTestDB(t)
	user := createTestUser(t, accountRepo)

	_, tokenHash, err := adapter.GenerateSecureToken()
	require.NoError(t, err)

	token := adapter.NewRefreshToken(user.ID, tokenHash, 7*24*time.Hour)
	err = tokenRepo.Create(token)
	require.NoError(t, err)

	// Delete token
	err = tokenRepo.Delete(token.ID)
	assert.NoError(t, err)

	// Verify token is deleted
	_, err = tokenRepo.GetByTokenHash(tokenHash)
	assert.Error(t, err)
	assert.Equal(t, adapter.ErrTokenNotFound, err)
}

func TestRefreshToken_IsValid(t *testing.T) {
	// Valid token
	token := adapter.NewRefreshToken("user-id", "hash", 7*24*time.Hour)
	assert.True(t, token.IsValid())
	assert.False(t, token.IsExpired())
	assert.False(t, token.IsRevoked())

	// Expired token
	expiredToken := adapter.NewRefreshToken("user-id", "hash", -1*time.Hour)
	assert.False(t, expiredToken.IsValid())
	assert.True(t, expiredToken.IsExpired())

	// Revoked token
	revokedTime := time.Now()
	revokedToken := adapter.NewRefreshToken("user-id", "hash", 7*24*time.Hour)
	revokedToken.RevokedAt = &revokedTime
	assert.False(t, revokedToken.IsValid())
	assert.True(t, revokedToken.IsRevoked())
}
