package database

import (
	"context"
	"database/sql"
	"time"
)

// Database represents a generic database connection interface
type Database interface {
	// DB returns the underlying *sql.DB for direct access when needed
	DB() *sql.DB

	// Ping checks if the database connection is alive
	Ping(ctx context.Context) error

	// Close closes the database connection
	Close() error

	// BeginTx starts a new transaction
	BeginTx(ctx context.Context) (*sql.Tx, error)

	// Exec executes a query without returning rows
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// Query executes a query that returns rows
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)

	// QueryRow executes a query that returns at most one row
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row

	// Stats returns database statistics
	Stats() sql.DBStats
}

// Config represents common database configuration
type Config struct {
	// Driver name: sqlite3, postgres, mysql
	Driver string

	// Connection string or DSN
	DSN string

	// Maximum number of open connections
	MaxOpenConns int

	// Maximum number of idle connections
	MaxIdleConns int

	// Maximum lifetime of a connection
	ConnMaxLifetime time.Duration

	// Maximum idle time of a connection
	ConnMaxIdleTime time.Duration
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig(driver, dsn string) *Config {
	return &Config{
		Driver:          driver,
		DSN:             dsn,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}
}
