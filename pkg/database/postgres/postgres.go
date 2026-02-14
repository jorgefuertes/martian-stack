package postgres

import (
	"fmt"
	"net/url"

	"github.com/jorgefuertes/martian-stack/pkg/database"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
)

const driverName = "pgx"

// Config represents PostgreSQL-specific configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string // disable, require, verify-ca, verify-full
	TimeZone string
	AppName  string
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Host:     "localhost",
		Port:     5432,
		SSLMode:  "prefer",
		TimeZone: "UTC",
		AppName:  "martian-stack",
	}
}

// New creates a new PostgreSQL database connection
func New(cfg *Config) (database.Database, error) {
	if cfg == nil {
		return nil, database.ErrInvalidConfig
	}

	if cfg.Host == "" || cfg.User == "" || cfg.Database == "" {
		return nil, database.ErrInvalidConfig
	}

	dsn := buildDSN(cfg)

	dbCfg := database.DefaultConfig(driverName, dsn)
	// PostgreSQL handles concurrent connections well
	dbCfg.MaxOpenConns = 25
	dbCfg.MaxIdleConns = 5

	return database.New(dbCfg)
}

// buildDSN builds the PostgreSQL DSN from config
func buildDSN(cfg *Config) string {
	// Build connection string in URL format
	// postgresql://user:password@host:port/database?sslmode=disable

	u := &url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:   cfg.Database,
	}

	if cfg.User != "" {
		if cfg.Password != "" {
			u.User = url.UserPassword(cfg.User, cfg.Password)
		} else {
			u.User = url.User(cfg.User)
		}
	}

	q := u.Query()
	if cfg.SSLMode != "" {
		q.Set("sslmode", cfg.SSLMode)
	}
	if cfg.TimeZone != "" {
		q.Set("timezone", cfg.TimeZone)
	}
	if cfg.AppName != "" {
		q.Set("application_name", cfg.AppName)
	}
	u.RawQuery = q.Encode()

	return u.String()
}
