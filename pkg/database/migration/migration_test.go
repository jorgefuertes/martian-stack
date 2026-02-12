package migration

import (
	"context"
	"testing"
	"time"

	"git.martianoids.com/martianoids/martian-stack/pkg/database/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestMigrator(t *testing.T) *Migrator {
	db, err := sqlite.NewInMemory()
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	return New(db)
}

func TestMigrator_Init(t *testing.T) {
	migrator := setupTestMigrator(t)
	ctx := context.Background()

	err := migrator.Init(ctx)
	assert.NoError(t, err)

	// Verify table exists
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations'`
	var tableName string
	err = migrator.db.QueryRow(ctx, query).Scan(&tableName)
	assert.NoError(t, err)
	assert.Equal(t, "schema_migrations", tableName)
}

func TestMigrator_Register(t *testing.T) {
	migrator := setupTestMigrator(t)

	migration := Migration{
		Version: 20260101000001,
		Name:    "test_migration",
		Up:      "CREATE TABLE test (id INTEGER PRIMARY KEY)",
		Down:    "DROP TABLE test",
	}

	migrator.Register(migration)
	assert.Len(t, migrator.migrations, 1)
	assert.Equal(t, migration.Version, migrator.migrations[0].Version)
}

func TestMigrator_RegisterMultiple(t *testing.T) {
	migrator := setupTestMigrator(t)

	migrations := []Migration{
		{Version: 1, Name: "first", Up: "SELECT 1", Down: "SELECT 1"},
		{Version: 2, Name: "second", Up: "SELECT 2", Down: "SELECT 2"},
	}

	migrator.RegisterMultiple(migrations)
	assert.Len(t, migrator.migrations, 2)
}

func TestMigrator_Up(t *testing.T) {
	migrator := setupTestMigrator(t)
	ctx := context.Background()

	// Register migrations
	migrations := []Migration{
		{
			Version: 20260101000001,
			Name:    "create_users",
			Up:      "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)",
			Down:    "DROP TABLE users",
		},
		{
			Version: 20260101000002,
			Name:    "create_posts",
			Up:      "CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT)",
			Down:    "DROP TABLE posts",
		},
	}
	migrator.RegisterMultiple(migrations)

	// Run migrations
	err := migrator.Up(ctx)
	assert.NoError(t, err)

	// Verify migrations were applied
	applied, err := migrator.getAppliedMigrations(ctx)
	assert.NoError(t, err)
	assert.Len(t, applied, 2)
	assert.Contains(t, applied, int64(20260101000001))
	assert.Contains(t, applied, int64(20260101000002))

	// Verify tables exist
	var tableName string
	err = migrator.db.QueryRow(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableName)
	assert.NoError(t, err)

	err = migrator.db.QueryRow(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name='posts'").Scan(&tableName)
	assert.NoError(t, err)
}

func TestMigrator_UpIdempotent(t *testing.T) {
	migrator := setupTestMigrator(t)
	ctx := context.Background()

	migration := Migration{
		Version: 20260101000001,
		Name:    "create_users",
		Up:      "CREATE TABLE users (id INTEGER PRIMARY KEY)",
		Down:    "DROP TABLE users",
	}
	migrator.Register(migration)

	// Run first time
	err := migrator.Up(ctx)
	assert.NoError(t, err)

	// Run second time - should be idempotent
	err = migrator.Up(ctx)
	assert.NoError(t, err)

	// Should only have one entry
	applied, err := migrator.getAppliedMigrations(ctx)
	assert.NoError(t, err)
	assert.Len(t, applied, 1)
}

func TestMigrator_Down(t *testing.T) {
	migrator := setupTestMigrator(t)
	ctx := context.Background()

	// Register and apply migrations
	migrations := []Migration{
		{
			Version: 20260101000001,
			Name:    "create_users",
			Up:      "CREATE TABLE users (id INTEGER PRIMARY KEY)",
			Down:    "DROP TABLE users",
		},
		{
			Version: 20260101000002,
			Name:    "create_posts",
			Up:      "CREATE TABLE posts (id INTEGER PRIMARY KEY)",
			Down:    "DROP TABLE posts",
		},
	}
	migrator.RegisterMultiple(migrations)

	err := migrator.Up(ctx)
	require.NoError(t, err)

	// Rollback last migration
	err = migrator.Down(ctx)
	assert.NoError(t, err)

	// Verify only first migration remains
	applied, err := migrator.getAppliedMigrations(ctx)
	assert.NoError(t, err)
	assert.Len(t, applied, 1)
	assert.Contains(t, applied, int64(20260101000001))

	// Verify posts table was dropped
	var tableName string
	err = migrator.db.QueryRow(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name='posts'").Scan(&tableName)
	assert.Error(t, err) // Should not exist
}

func TestMigrator_DownTo(t *testing.T) {
	migrator := setupTestMigrator(t)
	ctx := context.Background()

	// Register and apply migrations
	migrations := []Migration{
		{
			Version: 20260101000001,
			Name:    "create_users",
			Up:      "CREATE TABLE users (id INTEGER PRIMARY KEY)",
			Down:    "DROP TABLE users",
		},
		{
			Version: 20260101000002,
			Name:    "create_posts",
			Up:      "CREATE TABLE posts (id INTEGER PRIMARY KEY)",
			Down:    "DROP TABLE posts",
		},
		{
			Version: 20260101000003,
			Name:    "create_comments",
			Up:      "CREATE TABLE comments (id INTEGER PRIMARY KEY)",
			Down:    "DROP TABLE comments",
		},
	}
	migrator.RegisterMultiple(migrations)

	err := migrator.Up(ctx)
	require.NoError(t, err)

	// Rollback to version 1
	err = migrator.DownTo(ctx, 20260101000001)
	assert.NoError(t, err)

	// Verify only first migration remains
	applied, err := migrator.getAppliedMigrations(ctx)
	assert.NoError(t, err)
	assert.Len(t, applied, 1)
	assert.Contains(t, applied, int64(20260101000001))
}

func TestMigrator_Status(t *testing.T) {
	migrator := setupTestMigrator(t)
	ctx := context.Background()

	// Register migrations
	migrations := []Migration{
		{
			Version: 20260101000001,
			Name:    "create_users",
			Up:      "CREATE TABLE users (id INTEGER PRIMARY KEY)",
			Down:    "DROP TABLE users",
		},
		{
			Version: 20260101000002,
			Name:    "create_posts",
			Up:      "CREATE TABLE posts (id INTEGER PRIMARY KEY)",
			Down:    "DROP TABLE posts",
		},
	}
	migrator.RegisterMultiple(migrations)

	// Apply only first migration
	migrator.migrations = []Migration{migrations[0]}
	err := migrator.Up(ctx)
	require.NoError(t, err)

	// Restore all migrations
	migrator.migrations = migrations

	// Check status
	status, err := migrator.Status(ctx)
	assert.NoError(t, err)
	assert.Len(t, status, 2)

	// First should be applied
	assert.NotNil(t, status[0].AppliedAt)

	// Second should be pending
	assert.Nil(t, status[1].AppliedAt)
}

func TestMigrator_MigrationOrdering(t *testing.T) {
	migrator := setupTestMigrator(t)
	ctx := context.Background()

	// Register migrations out of order
	migrations := []Migration{
		{Version: 3, Name: "third", Up: "SELECT 3", Down: "SELECT 3"},
		{Version: 1, Name: "first", Up: "SELECT 1", Down: "SELECT 1"},
		{Version: 2, Name: "second", Up: "SELECT 2", Down: "SELECT 2"},
	}
	migrator.RegisterMultiple(migrations)

	err := migrator.Up(ctx)
	assert.NoError(t, err)

	// Verify they were applied in order
	applied, err := migrator.getAppliedMigrations(ctx)
	assert.NoError(t, err)

	// All should be applied
	assert.Len(t, applied, 3)

	// Check timestamps are in order (1 < 2 < 3)
	assert.True(t, applied[int64(1)].Before(applied[int64(2)]))
	assert.True(t, applied[int64(2)].Before(applied[int64(3)]))
}

func TestMigrator_TransactionRollback(t *testing.T) {
	migrator := setupTestMigrator(t)
	ctx := context.Background()

	// Migration with invalid SQL
	migration := Migration{
		Version: 20260101000001,
		Name:    "bad_migration",
		Up:      "INVALID SQL SYNTAX",
		Down:    "SELECT 1",
	}
	migrator.Register(migration)

	// Should fail
	err := migrator.Up(ctx)
	assert.Error(t, err)

	// Verify migration was not recorded
	applied, err := migrator.getAppliedMigrations(ctx)
	assert.NoError(t, err)
	assert.Len(t, applied, 0)
}

func TestGenerateVersion(t *testing.T) {
	version := GenerateVersion()
	assert.Greater(t, version, int64(20260101000000))
	assert.Less(t, version, int64(30000101000000))

	// Generate multiple versions in sequence
	v1 := GenerateVersion()
	time.Sleep(1 * time.Second)
	v2 := GenerateVersion()

	assert.Greater(t, v2, v1, "Later version should be greater")
}

func TestNewMigration(t *testing.T) {
	m := NewMigration("test_migration", "Test description")

	assert.NotZero(t, m.Version)
	assert.Equal(t, "test_migration", m.Name)
	assert.Equal(t, "Test description", m.Description)
}

func TestNewWithVersion(t *testing.T) {
	version := int64(20260101000001)
	m := NewWithVersion(version, "test_migration", "Test description")

	assert.Equal(t, version, m.Version)
	assert.Equal(t, "test_migration", m.Name)
	assert.Equal(t, "Test description", m.Description)
}
