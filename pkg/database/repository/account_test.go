package repository

import (
	"testing"
	"time"

	"git.martianoids.com/martianoids/martian-stack/pkg/database/sqlite"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/adapter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *SQLAccountRepository {
	db, err := sqlite.NewInMemory()
	require.NoError(t, err)

	// Create schema
	err = sqlite.CreateAccountsTable(db)
	require.NoError(t, err)

	return NewSQLAccountRepository(db)
}

func createTestAccount(t *testing.T) *adapter.Account {
	acc := &adapter.Account{
		Username: "testuser",
		Name:     "Test User",
		Email:    "test@example.com",
		Enabled:  true,
		Role:     "user",
	}

	err := acc.SetPassword("password123")
	require.NoError(t, err)

	return acc
}

func TestSQLAccountRepository_Create(t *testing.T) {
	repo := setupTestDB(t)
	acc := createTestAccount(t)

	err := repo.Create(acc)
	assert.NoError(t, err)
	assert.NotEmpty(t, acc.ID, "ID should be generated")
	assert.False(t, acc.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.False(t, acc.UpdatedAt.IsZero(), "UpdatedAt should be set")
}

func TestSQLAccountRepository_CreateWithID(t *testing.T) {
	repo := setupTestDB(t)
	acc := createTestAccount(t)
	acc.ID = "some-id"

	err := repo.Create(acc)
	assert.Error(t, err)
	assert.Equal(t, adapter.ErrCannotCreateWithID, err)
}

func TestSQLAccountRepository_CreateWithoutPassword(t *testing.T) {
	repo := setupTestDB(t)
	acc := &adapter.Account{
		Username: "testuser",
		Name:     "Test User",
		Email:    "test@example.com",
		Enabled:  true,
		Role:     "user",
	}

	err := repo.Create(acc)
	assert.Error(t, err)
	assert.Equal(t, adapter.ErrPasswordNotSet, err)
}

func TestSQLAccountRepository_Get(t *testing.T) {
	repo := setupTestDB(t)
	acc := createTestAccount(t)

	err := repo.Create(acc)
	require.NoError(t, err)

	// Get by ID
	retrieved, err := repo.Get(acc.ID)
	assert.NoError(t, err)
	assert.Equal(t, acc.ID, retrieved.ID)
	assert.Equal(t, acc.Username, retrieved.Username)
	assert.Equal(t, acc.Email, retrieved.Email)
	assert.Equal(t, acc.Name, retrieved.Name)
}

func TestSQLAccountRepository_GetNotFound(t *testing.T) {
	repo := setupTestDB(t)

	retrieved, err := repo.Get("non-existent-id")
	assert.Error(t, err)
	assert.Equal(t, adapter.ErrAccountNotFound, err)
	assert.Nil(t, retrieved)
}

func TestSQLAccountRepository_Exists(t *testing.T) {
	repo := setupTestDB(t)
	acc := createTestAccount(t)

	err := repo.Create(acc)
	require.NoError(t, err)

	// Should exist
	exists := repo.Exists(acc.ID)
	assert.True(t, exists)

	// Should not exist
	exists = repo.Exists("non-existent-id")
	assert.False(t, exists)
}

func TestSQLAccountRepository_GetByEmail(t *testing.T) {
	repo := setupTestDB(t)
	acc := createTestAccount(t)

	err := repo.Create(acc)
	require.NoError(t, err)

	// Get by email
	retrieved, err := repo.GetByEmail(acc.Email)
	assert.NoError(t, err)
	assert.Equal(t, acc.ID, retrieved.ID)
	assert.Equal(t, acc.Email, retrieved.Email)
}

func TestSQLAccountRepository_GetByEmailNotFound(t *testing.T) {
	repo := setupTestDB(t)

	retrieved, err := repo.GetByEmail("nonexistent@example.com")
	assert.Error(t, err)
	assert.Equal(t, adapter.ErrAccountNotFound, err)
	assert.Nil(t, retrieved)
}

func TestSQLAccountRepository_GetByUsername(t *testing.T) {
	repo := setupTestDB(t)
	acc := createTestAccount(t)

	err := repo.Create(acc)
	require.NoError(t, err)

	// Get by username
	retrieved, err := repo.GetByUsername(acc.Username)
	assert.NoError(t, err)
	assert.Equal(t, acc.ID, retrieved.ID)
	assert.Equal(t, acc.Username, retrieved.Username)
}

func TestSQLAccountRepository_GetByUsernameNotFound(t *testing.T) {
	repo := setupTestDB(t)

	retrieved, err := repo.GetByUsername("nonexistentuser")
	assert.Error(t, err)
	assert.Equal(t, adapter.ErrAccountNotFound, err)
	assert.Nil(t, retrieved)
}

func TestSQLAccountRepository_Update(t *testing.T) {
	repo := setupTestDB(t)
	acc := createTestAccount(t)

	err := repo.Create(acc)
	require.NoError(t, err)

	// Wait a bit to ensure UpdatedAt changes
	time.Sleep(10 * time.Millisecond)

	// Update account
	acc.Name = "Updated Name"
	acc.Email = "updated@example.com"

	err = repo.Update(acc)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := repo.Get(acc.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", retrieved.Name)
	assert.Equal(t, "updated@example.com", retrieved.Email)
	assert.True(t, retrieved.UpdatedAt.After(retrieved.CreatedAt))
}

func TestSQLAccountRepository_UpdateNotFound(t *testing.T) {
	repo := setupTestDB(t)
	acc := createTestAccount(t)
	acc.ID = "non-existent-id"

	err := repo.Update(acc)
	assert.Error(t, err)
	assert.Equal(t, adapter.ErrAccountNotFound, err)
}

func TestSQLAccountRepository_Delete(t *testing.T) {
	repo := setupTestDB(t)
	acc := createTestAccount(t)

	err := repo.Create(acc)
	require.NoError(t, err)

	// Delete account
	err = repo.Delete(acc.ID)
	assert.NoError(t, err)

	// Verify deletion
	exists := repo.Exists(acc.ID)
	assert.False(t, exists)
}

func TestSQLAccountRepository_DeleteNotFound(t *testing.T) {
	repo := setupTestDB(t)

	err := repo.Delete("non-existent-id")
	assert.Error(t, err)
	assert.Equal(t, adapter.ErrAccountNotFound, err)
}

func TestSQLAccountRepository_DuplicateEmail(t *testing.T) {
	repo := setupTestDB(t)

	// Create first account
	acc1 := createTestAccount(t)
	err := repo.Create(acc1)
	require.NoError(t, err)

	// Try to create second account with same email
	acc2 := &adapter.Account{
		Username: "differentuser",
		Name:     "Different User",
		Email:    acc1.Email, // Same email
		Enabled:  true,
		Role:     "user",
	}
	err = acc2.SetPassword("password123")
	require.NoError(t, err)

	err = repo.Create(acc2)
	assert.Error(t, err)
}

func TestSQLAccountRepository_DuplicateUsername(t *testing.T) {
	repo := setupTestDB(t)

	// Create first account
	acc1 := createTestAccount(t)
	err := repo.Create(acc1)
	require.NoError(t, err)

	// Try to create second account with same username
	acc2 := &adapter.Account{
		Username: acc1.Username, // Same username
		Name:     "Different User",
		Email:    "different@example.com",
		Enabled:  true,
		Role:     "user",
	}
	err = acc2.SetPassword("password123")
	require.NoError(t, err)

	err = repo.Create(acc2)
	assert.Error(t, err)
}

func TestSQLAccountRepository_LastLogin(t *testing.T) {
	repo := setupTestDB(t)
	acc := createTestAccount(t)

	// Create without last login
	err := repo.Create(acc)
	require.NoError(t, err)

	// Verify last login is zero
	retrieved, err := repo.Get(acc.ID)
	assert.NoError(t, err)
	assert.True(t, retrieved.LastLogin.IsZero())

	// Update with last login
	now := time.Now()
	acc.LastLogin = now
	err = repo.Update(acc)
	assert.NoError(t, err)

	// Verify last login is set
	retrieved, err = repo.Get(acc.ID)
	assert.NoError(t, err)
	assert.False(t, retrieved.LastLogin.IsZero())
	assert.WithinDuration(t, now, retrieved.LastLogin, time.Second)
}

func TestSQLAccountRepository_PasswordValidation(t *testing.T) {
	repo := setupTestDB(t)
	acc := createTestAccount(t)
	password := "mySecurePassword123"

	err := acc.SetPassword(password)
	require.NoError(t, err)

	err = repo.Create(acc)
	require.NoError(t, err)

	// Retrieve and validate password
	retrieved, err := repo.Get(acc.ID)
	assert.NoError(t, err)

	// Correct password
	err = retrieved.ValidatePassword(password)
	assert.NoError(t, err)

	// Wrong password
	err = retrieved.ValidatePassword("wrongpassword")
	assert.Error(t, err)
}
