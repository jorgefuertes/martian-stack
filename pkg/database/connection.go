package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Connection implements the Database interface wrapping *sql.DB
type Connection struct {
	db     *sql.DB
	driver string
}

// New creates a new database connection from config
func New(cfg *Config) (Database, error) {
	if cfg == nil {
		return nil, ErrInvalidConfig
	}

	if cfg.Driver == "" || cfg.DSN == "" {
		return nil, ErrInvalidConfig
	}

	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	return &Connection{
		db:     db,
		driver: cfg.Driver,
	}, nil
}

// DB returns the underlying *sql.DB
func (c *Connection) DB() *sql.DB {
	return c.db
}

// Ping checks if the database connection is alive
func (c *Connection) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

// Close closes the database connection
func (c *Connection) Close() error {
	return c.db.Close()
}

// BeginTx starts a new transaction
func (c *Connection) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return c.db.BeginTx(ctx, nil)
}

// Exec executes a query without returning rows
func (c *Connection) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return c.db.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows
func (c *Connection) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return c.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns at most one row
func (c *Connection) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return c.db.QueryRowContext(ctx, query, args...)
}

// Stats returns database statistics
func (c *Connection) Stats() sql.DBStats {
	return c.db.Stats()
}
