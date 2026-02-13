package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInMemory(t *testing.T) {
	db, err := NewInMemory()
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.Ping(ctx)
	assert.NoError(t, err)
}

func TestNew(t *testing.T) {
	// Use in-memory database for testing
	cfg := &Config{
		Path:        ":memory:",
		ForeignKeys: true,
		JournalMode: "MEMORY",
	}

	db, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Test connection
	ctx := context.Background()
	err = db.Ping(ctx)
	assert.NoError(t, err)
}

func TestCreateAccountsTable(t *testing.T) {
	db, err := NewInMemory()
	require.NoError(t, err)
	defer db.Close()

	// Create table
	err = CreateAccountsTable(db)
	assert.NoError(t, err)

	// Verify table exists
	ctx := context.Background()
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name='accounts'`

	var tableName string
	err = db.QueryRow(ctx, query).Scan(&tableName)
	assert.NoError(t, err)
	assert.Equal(t, "accounts", tableName)
}

func TestSQLiteConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Path:        ":memory:",
				ForeignKeys: true,
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, db)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)
				if db != nil {
					db.Close()
				}
			}
		})
	}
}

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		contains []string
	}{
		{
			name: "memory database",
			config: &Config{
				Path: ":memory:",
			},
			contains: []string{":memory:"},
		},
		{
			name: "file database with options",
			config: &Config{
				Path:        "./test.db",
				Mode:        "rwc",
				Cache:       "shared",
				ForeignKeys: true,
				JournalMode: "WAL",
				BusyTimeout: 5000,
			},
			contains: []string{
				"file:test.db",
				"mode=rwc",
				"cache=shared",
				"_foreign_keys=1",
				"_journal_mode=WAL",
				"_busy_timeout=5000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := buildDSN(tt.config)
			for _, substr := range tt.contains {
				assert.Contains(t, dsn, substr)
			}
		})
	}
}
