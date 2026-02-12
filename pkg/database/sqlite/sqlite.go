package sqlite

import (
	"fmt"
	"path/filepath"

	"git.martianoids.com/martianoids/martian-stack/pkg/database"
	_ "modernc.org/sqlite" // SQLite driver
)

const driverName = "sqlite"

// Config represents SQLite-specific configuration
type Config struct {
	// Path to the database file (:memory: for in-memory)
	Path string

	// Mode: ro (read-only), rw (read-write), rwc (read-write-create), memory
	Mode string

	// Cache: shared or private
	Cache string

	// Enable foreign keys
	ForeignKeys bool

	// Journal mode: DELETE, TRUNCATE, PERSIST, MEMORY, WAL, OFF
	JournalMode string

	// Busy timeout in milliseconds
	BusyTimeout int
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig(path string) *Config {
	return &Config{
		Path:        path,
		Mode:        "rwc",
		Cache:       "shared",
		ForeignKeys: true,
		JournalMode: "WAL",
		BusyTimeout: 5000,
	}
}

// New creates a new SQLite database connection
func New(cfg *Config) (database.Database, error) {
	if cfg == nil {
		return nil, database.ErrInvalidConfig
	}

	dsn := buildDSN(cfg)

	dbCfg := database.DefaultConfig(driverName, dsn)
	// SQLite doesn't benefit from many connections
	dbCfg.MaxOpenConns = 1
	dbCfg.MaxIdleConns = 1

	return database.New(dbCfg)
}

// NewInMemory creates an in-memory SQLite database
func NewInMemory() (database.Database, error) {
	return New(&Config{
		Path:        ":memory:",
		ForeignKeys: true,
		JournalMode: "MEMORY",
	})
}

// buildDSN builds the SQLite DSN from config
func buildDSN(cfg *Config) string {
	if cfg.Path == ":memory:" {
		return ":memory:"
	}

	// Clean the path
	path := filepath.Clean(cfg.Path)

	// Build query parameters
	params := make(map[string]string)

	if cfg.Mode != "" {
		params["mode"] = cfg.Mode
	}

	if cfg.Cache != "" {
		params["cache"] = cfg.Cache
	}

	if cfg.ForeignKeys {
		params["_foreign_keys"] = "1"
	}

	if cfg.JournalMode != "" {
		params["_journal_mode"] = cfg.JournalMode
	}

	if cfg.BusyTimeout > 0 {
		params["_busy_timeout"] = fmt.Sprintf("%d", cfg.BusyTimeout)
	}

	// Build DSN
	dsn := "file:" + path + "?"
	first := true
	for k, v := range params {
		if !first {
			dsn += "&"
		}
		dsn += k + "=" + v
		first = false
	}

	return dsn
}
