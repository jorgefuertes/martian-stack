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

func setupPasswordResetTokenTestDB(t *testing.T) (*SQLPasswordResetTokenRepository, *SQLAccountRepository) {
	db, err := sqlite.NewInMemory()
	require.NoError(t, err)

	// Run migrations
	migrator := migration.New(db)
	migrator.RegisterMultiple(migrations.All())
	err = migrator.Up(context.Background())
	require.NoError(t, err)

	return NewSQLPasswordResetTokenRepository(db), NewSQLAccountRepository(db)
}

func TestPasswordResetTokenRepository_Create(t *testing.T) {
	tokenRepo, accountRepo := setupPasswordResetTokenTestDB(t)
	user := createTestUser(t, accountRepo)

	_, tokenHash, err := adapter.GenerateSecureToken()
	require.NoError(t, err)
	require.NotEmpty(t, tokenHash)

	token := adapter.NewPasswordResetToken(user.ID, tokenHash, 1*time.Hour)

	err = tokenRepo.Create(token)
	assert.NoError(t, err)
	assert.NotEmpty(t, token.ID, "ID should be generated")
	assert.False(t, token.CreatedAt.IsZero(), "CreatedAt should be set")
}

func TestPasswordResetTokenRepository_GetByTokenHash(t *testing.T) {
	tokenRepo, accountRepo := setupPasswordResetTokenTestDB(t)
	user := createTestUser(t, accountRepo)

	_, tokenHash, err := adapter.GenerateSecureToken()
	require.NoError(t, err)

	token := adapter.NewPasswordResetToken(user.ID, tokenHash, 1*time.Hour)
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

func TestPasswordResetTokenRepository_GetByTokenHash_NotFound(t *testing.T) {
	tokenRepo, _ := setupPasswordResetTokenTestDB(t)

	retrieved, err := tokenRepo.GetByTokenHash("nonexistent")
	assert.Error(t, err)
	assert.Equal(t, adapter.ErrTokenNotFound, err)
	assert.Nil(t, retrieved)
}

func TestPasswordResetTokenRepository_MarkAsUsed(t *testing.T) {
	tokenRepo, accountRepo := setupPasswordResetTokenTestDB(t)
	user := createTestUser(t, accountRepo)

	_, tokenHash, err := adapter.GenerateSecureToken()
	require.NoError(t, err)

	token := adapter.NewPasswordResetToken(user.ID, tokenHash, 1*time.Hour)
	err = tokenRepo.Create(token)
	require.NoError(t, err)

	// Mark token as used
	err = tokenRepo.MarkAsUsed(tokenHash)
	assert.NoError(t, err)

	// Verify token is marked as used
	retrieved, err := tokenRepo.GetByTokenHash(tokenHash)
	assert.NoError(t, err)
	assert.True(t, retrieved.IsUsed())
	assert.False(t, retrieved.IsValid())
}

func TestPasswordResetTokenRepository_DeleteExpired(t *testing.T) {
	tokenRepo, accountRepo := setupPasswordResetTokenTestDB(t)
	user := createTestUser(t, accountRepo)

	// Create an expired token
	_, expiredHash, err := adapter.GenerateSecureToken()
	require.NoError(t, err)

	expiredToken := adapter.NewPasswordResetToken(user.ID, expiredHash, -1*time.Hour)
	err = tokenRepo.Create(expiredToken)
	require.NoError(t, err)

	// Create a valid token
	_, validHash, err := adapter.GenerateSecureToken()
	require.NoError(t, err)

	validToken := adapter.NewPasswordResetToken(user.ID, validHash, 1*time.Hour)
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

func TestPasswordResetTokenRepository_DeleteByUserID(t *testing.T) {
	tokenRepo, accountRepo := setupPasswordResetTokenTestDB(t)
	user := createTestUser(t, accountRepo)

	// Create multiple tokens for the same user
	var tokenHashes []string
	for i := 0; i < 3; i++ {
		_, tokenHash, err := adapter.GenerateSecureToken()
		require.NoError(t, err)
		tokenHashes = append(tokenHashes, tokenHash)

		token := adapter.NewPasswordResetToken(user.ID, tokenHash, 1*time.Hour)
		err = tokenRepo.Create(token)
		require.NoError(t, err)
	}

	// Delete all tokens for user
	err := tokenRepo.DeleteByUserID(user.ID)
	assert.NoError(t, err)

	// Verify all tokens are deleted
	for _, tokenHash := range tokenHashes {
		_, err := tokenRepo.GetByTokenHash(tokenHash)
		assert.Error(t, err)
		assert.Equal(t, adapter.ErrTokenNotFound, err)
	}
}

func TestPasswordResetToken_IsValid(t *testing.T) {
	// Valid token
	token := adapter.NewPasswordResetToken("user-id", "hash", 1*time.Hour)
	assert.True(t, token.IsValid())
	assert.False(t, token.IsExpired())
	assert.False(t, token.IsUsed())

	// Expired token
	expiredToken := adapter.NewPasswordResetToken("user-id", "hash", -1*time.Hour)
	assert.False(t, expiredToken.IsValid())
	assert.True(t, expiredToken.IsExpired())

	// Used token
	usedTime := time.Now()
	usedToken := adapter.NewPasswordResetToken("user-id", "hash", 1*time.Hour)
	usedToken.UsedAt = &usedTime
	assert.False(t, usedToken.IsValid())
	assert.True(t, usedToken.IsUsed())
}
