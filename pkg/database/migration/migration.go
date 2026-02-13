package migration

import (
	"context"
	"fmt"
	"sort"
	"time"

	"git.martianoids.com/martianoids/martian-stack/pkg/database"
)

// Migration represents a database migration
type Migration struct {
	Version     int64
	Name        string
	Up          string
	Down        string
	AppliedAt   *time.Time
	Description string
}

// Migrator manages database migrations
type Migrator struct {
	db         database.Database
	migrations []Migration
}

// New creates a new Migrator
func New(db database.Database) *Migrator {
	return &Migrator{
		db:         db,
		migrations: make([]Migration, 0),
	}
}

// Register adds a migration to the migrator
func (m *Migrator) Register(migration Migration) {
	m.migrations = append(m.migrations, migration)
}

// RegisterMultiple adds multiple migrations to the migrator
func (m *Migrator) RegisterMultiple(migrations []Migration) {
	m.migrations = append(m.migrations, migrations...)
}

// Init creates the migrations table if it doesn't exist
func (m *Migrator) Init(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := m.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("%w: failed to create migrations table: %v", database.ErrMigrationFailed, err)
	}

	return nil
}

// Up runs all pending migrations
func (m *Migrator) Up(ctx context.Context) error {
	// Ensure migrations table exists
	if err := m.Init(ctx); err != nil {
		return err
	}

	// Get applied migrations
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	// Run pending migrations
	for _, migration := range m.migrations {
		if _, exists := applied[migration.Version]; exists {
			continue // Already applied
		}

		if err := m.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("%w: failed to apply migration %d (%s): %v",
				database.ErrMigrationFailed, migration.Version, migration.Name, err)
		}
	}

	return nil
}

// Down rolls back the last migration
func (m *Migrator) Down(ctx context.Context) error {
	// Get applied migrations
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	if len(applied) == 0 {
		return nil // No migrations to rollback
	}

	// Find the last applied migration
	var lastVersion int64
	for version := range applied {
		if version > lastVersion {
			lastVersion = version
		}
	}

	// Find the migration definition
	var targetMigration *Migration
	for i := range m.migrations {
		if m.migrations[i].Version == lastVersion {
			targetMigration = &m.migrations[i]
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf(
			"%w: migration %d not found in registered migrations",
			database.ErrMigrationFailed,
			lastVersion,
		)
	}

	if targetMigration.Down == "" {
		return fmt.Errorf("%w: migration %d has no down script", database.ErrMigrationFailed, lastVersion)
	}

	return m.rollbackMigration(ctx, *targetMigration)
}

// DownTo rolls back migrations to a specific version
func (m *Migrator) DownTo(ctx context.Context, targetVersion int64) error {
	// Get applied migrations
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	// Sort migrations by version (descending)
	var versions []int64
	for version := range applied {
		if version > targetVersion {
			versions = append(versions, version)
		}
	}
	sort.Slice(versions, func(i, j int) bool {
		return versions[i] > versions[j]
	})

	// Rollback migrations
	for _, version := range versions {
		var targetMigration *Migration
		for i := range m.migrations {
			if m.migrations[i].Version == version {
				targetMigration = &m.migrations[i]
				break
			}
		}

		if targetMigration == nil {
			return fmt.Errorf("%w: migration %d not found", database.ErrMigrationFailed, version)
		}

		if err := m.rollbackMigration(ctx, *targetMigration); err != nil {
			return err
		}
	}

	return nil
}

// Status returns the current migration status
func (m *Migrator) Status(ctx context.Context) ([]Migration, error) {
	// Ensure migrations table exists
	if err := m.Init(ctx); err != nil {
		return nil, err
	}

	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return nil, err
	}

	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	// Build status
	status := make([]Migration, len(m.migrations))
	for i, migration := range m.migrations {
		status[i] = migration
		if appliedAt, exists := applied[migration.Version]; exists {
			status[i].AppliedAt = &appliedAt
		}
	}

	return status, nil
}

// getAppliedMigrations retrieves all applied migrations from the database
func (m *Migrator) getAppliedMigrations(ctx context.Context) (map[int64]time.Time, error) {
	query := `SELECT version, applied_at FROM schema_migrations ORDER BY version`

	rows, err := m.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int64]time.Time)
	for rows.Next() {
		var version int64
		var appliedAt time.Time
		if err := rows.Scan(&version, &appliedAt); err != nil {
			return nil, err
		}
		applied[version] = appliedAt
	}

	return applied, rows.Err()
}

// applyMigration applies a single migration
func (m *Migrator) applyMigration(ctx context.Context, migration Migration) error {
	// Start transaction
	tx, err := m.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration
	if _, err := tx.ExecContext(ctx, migration.Up); err != nil {
		return err
	}

	// Record migration
	insertQuery := `INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)`
	if _, err := tx.ExecContext(ctx, insertQuery, migration.Version, migration.Name, time.Now()); err != nil {
		return err
	}

	return tx.Commit()
}

// rollbackMigration rolls back a single migration
func (m *Migrator) rollbackMigration(ctx context.Context, migration Migration) error {
	// Start transaction
	tx, err := m.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute rollback
	if _, err := tx.ExecContext(ctx, migration.Down); err != nil {
		return err
	}

	// Remove migration record
	deleteQuery := `DELETE FROM schema_migrations WHERE version = ?`
	if _, err := tx.ExecContext(ctx, deleteQuery, migration.Version); err != nil {
		return err
	}

	return tx.Commit()
}
