package mysql

import (
	"fmt"

	"github.com/jorgefuertes/martian-stack/pkg/database"
	_ "github.com/go-sql-driver/mysql" // MySQL/MariaDB driver
)

const driverName = "mysql"

// Config represents MySQL/MariaDB-specific configuration
type Config struct {
	Host      string
	Port      int
	User      string
	Password  string
	Database  string
	Charset   string
	ParseTime bool
	Loc       string // Location for time.Time
	TLS       string // TLS configuration: true, false, skip-verify, preferred
	Timeout   string // Connection timeout
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Host:      "localhost",
		Port:      3306,
		Charset:   "utf8mb4",
		ParseTime: true,
		Loc:       "UTC",
		TLS:       "preferred",
		Timeout:   "10s",
	}
}

// New creates a new MySQL/MariaDB database connection
func New(cfg *Config) (database.Database, error) {
	if cfg == nil {
		return nil, database.ErrInvalidConfig
	}

	if cfg.Host == "" || cfg.User == "" || cfg.Database == "" {
		return nil, database.ErrInvalidConfig
	}

	dsn := buildDSN(cfg)

	dbCfg := database.DefaultConfig(driverName, dsn)
	// MySQL/MariaDB handles concurrent connections well
	dbCfg.MaxOpenConns = 25
	dbCfg.MaxIdleConns = 5

	return database.New(dbCfg)
}

// buildDSN builds the MySQL/MariaDB DSN from config
func buildDSN(cfg *Config) string {
	// Format: user:password@tcp(host:port)/database?param1=value1&param2=value2

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	// Add parameters
	params := make([]string, 0)

	if cfg.Charset != "" {
		params = append(params, fmt.Sprintf("charset=%s", cfg.Charset))
	}

	if cfg.ParseTime {
		params = append(params, "parseTime=true")
	}

	if cfg.Loc != "" {
		params = append(params, fmt.Sprintf("loc=%s", cfg.Loc))
	}

	if cfg.TLS != "" {
		params = append(params, fmt.Sprintf("tls=%s", cfg.TLS))
	}

	if cfg.Timeout != "" {
		params = append(params, fmt.Sprintf("timeout=%s", cfg.Timeout))
	}

	if len(params) > 0 {
		dsn += "?"
		for i, param := range params {
			if i > 0 {
				dsn += "&"
			}
			dsn += param
		}
	}

	return dsn
}
