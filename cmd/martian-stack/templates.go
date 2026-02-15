package main

const tplMain = `package main

import (
	"os"
	"strconv"
{{- if or .MwTimeout}}
	"time"
{{- end}}
{{- if .HasDatabase}}

	"{{.ModulePath}}/database"
	"{{.ModulePath}}/database/migrations"
{{- end}}
	"{{.ModulePath}}/handlers"

	"github.com/jorgefuertes/martian-stack/pkg/server"
{{- if .HasAuth}}
	"github.com/jorgefuertes/martian-stack/pkg/auth"
	"github.com/jorgefuertes/martian-stack/pkg/auth/jwt"
	"github.com/jorgefuertes/martian-stack/pkg/database/repository"
{{- end}}
{{- if .HasRedis}}
	"github.com/jorgefuertes/martian-stack/pkg/service/cache/redis"
{{- else}}
	"github.com/jorgefuertes/martian-stack/pkg/service/cache/memory"
{{- end}}
	"github.com/jorgefuertes/martian-stack/pkg/server/middleware"
	"github.com/jorgefuertes/martian-stack/pkg/service/logger"
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envOrInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func main() {
	// Logger
	l := logger.New(os.Stdout, logger.TextFormat, logger.LevelDebug)
{{- if .HasDatabase}}

	// Database
	db, err := database.Connect()
	if err != nil {
		l.From("main", "database").Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	// Migrations
	if err := migrations.Run(db); err != nil {
		l.From("main", "migrations").Error(err.Error())
		os.Exit(1)
	}
{{- end}}

	// Cache
{{- if .HasRedis}}
	cacheSvc := redis.New(
		l,
		envOr("REDIS_HOST", "localhost"),
		envOrInt("REDIS_PORT", 6379),
		envOr("REDIS_USER", ""),
		envOr("REDIS_PASS", ""),
		envOrInt("REDIS_DB", 0),
	)
	defer cacheSvc.Close()
{{- else}}
	cacheSvc := memory.New()
	defer cacheSvc.Close()
{{- end}}
	_ = cacheSvc
{{- if .HasAuth}}

	// JWT
	jwtCfg, err := jwt.DefaultConfig(envOr("JWT_SECRET", "change-me-to-a-secure-secret-at-least-32-bytes!!"))
	if err != nil {
		l.From("main", "jwt").Error(err.Error())
		os.Exit(1)
	}
	jwtService := jwt.NewService(jwtCfg)

	// Repositories
	accountRepo := repository.NewSQLAccountRepository(db)
	refreshTokenRepo := repository.NewSQLRefreshTokenRepository(db)
	resetTokenRepo := repository.NewSQLPasswordResetTokenRepository(db)

	// Auth
	authHandlers := auth.NewHandlers(accountRepo, jwtService, refreshTokenRepo, resetTokenRepo)
	authMiddleware := auth.NewMiddleware(jwtService)
{{- end}}

	// Server
	srv := server.New(
		envOr("HOST", "{{.DefaultHost}}"),
		envOr("PORT", "{{.DefaultPort}}"),
		envOrInt("TIMEOUT", {{.DefaultTimeout}}),
	)

	// Middlewares
	srv.Use(
{{- if .MwRecovery}}
		middleware.NewRecovery(),
{{- end}}
{{- if .MwSecurityHeaders}}
		middleware.NewSecurityHeaders(),
{{- end}}
{{- if .MwRateLimit}}
		middleware.NewRateLimit(middleware.DefaultRateLimitConfig()),
{{- end}}
{{- if .MwCORS}}
		middleware.NewCors(middleware.NewCorsOptions()),
{{- end}}
{{- if .MwTimeout}}
		middleware.NewTimeout(time.Duration(envOrInt("TIMEOUT", {{.DefaultTimeout}}))*time.Second),
{{- end}}
{{- if .MwLogging}}
		middleware.NewLog(l),
{{- end}}
	)

	// Routes
	handlers.RegisterRoutes(srv)
{{- if .HasAuth}}
	handlers.RegisterAuthRoutes(srv, authHandlers, authMiddleware)
{{- end}}
{{- if .HasAdmin}}
	handlers.RegisterAdminRoutes(srv, authMiddleware, accountRepo)
{{- end}}

	// Start
	l.From("main", "server").With(
		"host", envOr("HOST", "{{.DefaultHost}}"),
		"port", envOr("PORT", "{{.DefaultPort}}"),
	).Info("starting server")

	if err := srv.ListenAndShutdown(func() {
		l.From("main", "server").Info("shutting down")
{{- if .HasDatabase}}
		db.Close()
{{- end}}
{{- if .HasRedis}}
		cacheSvc.Close()
{{- end}}
	}); err != nil {
		l.From("main", "server").Error(err.Error())
		os.Exit(1)
	}
}
`

const tplGoMod = `module {{.ModulePath}}

go 1.25

require (
	github.com/jorgefuertes/martian-stack v0.0.0
)
`

const tplMakefile = `APP_NAME := {{.ProjectName}}
MAIN := ./main.go

.PHONY: build run clean tidy

build:
	go build -o bin/$(APP_NAME) .

run: build
	./bin/$(APP_NAME)

clean:
	rm -rf bin/

tidy:
	go mod tidy
`

const tplEnvExample = `# Server
HOST=localhost
PORT=8080
TIMEOUT=15
{{- if .HasDatabase}}
{{- if eq .Database "sqlite"}}

# Database (SQLite)
DB_DRIVER=sqlite3
DB_DSN=./{{.ProjectName}}.db
{{- end}}
{{- if eq .Database "postgres"}}

# Database (PostgreSQL)
DB_DRIVER=postgres
DB_DSN=postgres://user:password@localhost:5432/{{.ProjectName}}?sslmode=disable
{{- end}}
{{- if eq .Database "mysql"}}

# Database (MySQL)
DB_DRIVER=mysql
DB_DSN=user:password@tcp(localhost:3306)/{{.ProjectName}}?parseTime=true
{{- end}}
{{- end}}
{{- if .HasRedis}}

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_USER=
REDIS_PASS=
REDIS_DB=0
{{- end}}
{{- if .HasAuth}}

# JWT
JWT_SECRET=change-me-to-a-secure-secret-at-least-32-bytes!!
{{- end}}
`

const tplGitignore = `# Binaries
bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test
*.test
*.out
coverage.html

# Env
.env

# Database
*.db
*.db-journal
*.db-wal
*.db-shm

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db
`

const tplRoutes = `package handlers

import (
	"github.com/jorgefuertes/martian-stack/pkg/server"
	"github.com/jorgefuertes/martian-stack/pkg/server/ctx"
	"github.com/jorgefuertes/martian-stack/pkg/server/web"
)

// RegisterRoutes registers the application routes
func RegisterRoutes(srv *server.Server) {
	srv.Route(web.MethodGet, "/", homeHandler())
	srv.Route(web.MethodGet, "/health", healthHandler())
}

func homeHandler() ctx.Handler {
	return func(c ctx.Ctx) error {
		return c.SendJSON(map[string]string{
			"message": "Welcome to {{.ProjectName}}!",
		})
	}
}

func healthHandler() ctx.Handler {
	return func(c ctx.Ctx) error {
		return c.SendJSON(map[string]string{
			"status": "ok",
		})
	}
}
`

const tplDatabase = `package database

import (
	"os"

	"github.com/jorgefuertes/martian-stack/pkg/database"
{{- if eq .Database "sqlite"}}
	_ "modernc.org/sqlite"
{{- end}}
{{- if eq .Database "postgres"}}
	_ "github.com/jackc/pgx/v5/stdlib"
{{- end}}
{{- if eq .Database "mysql"}}
	_ "github.com/go-sql-driver/mysql"
{{- end}}
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Connect creates a new database connection using environment variables
func Connect() (database.Database, error) {
{{- if eq .Database "sqlite"}}
	driver := envOr("DB_DRIVER", "sqlite3")
	dsn := envOr("DB_DSN", "./{{.ProjectName}}.db")
{{- end}}
{{- if eq .Database "postgres"}}
	driver := envOr("DB_DRIVER", "pgx")
	dsn := envOr("DB_DSN", "postgres://user:password@localhost:5432/{{.ProjectName}}?sslmode=disable")
{{- end}}
{{- if eq .Database "mysql"}}
	driver := envOr("DB_DRIVER", "mysql")
	dsn := envOr("DB_DSN", "user:password@tcp(localhost:3306)/{{.ProjectName}}?parseTime=true")
{{- end}}

	cfg := database.DefaultConfig(driver, dsn)

	return database.New(cfg)
}
`

const tplMigrations = `package migrations

import (
	"context"

	"github.com/jorgefuertes/martian-stack/pkg/database"
	"github.com/jorgefuertes/martian-stack/pkg/database/migration"
)

// Run executes all pending database migrations
func Run(db database.Database) error {
	m := migration.New(db)
	m.RegisterMultiple(All())

	ctx := context.Background()
	if err := m.Init(ctx); err != nil {
		return err
	}

	return m.Up(ctx)
}

// All returns all available migrations
func All() []migration.Migration {
	return []migration.Migration{
		InitialSchema,
{{- if .HasAuth}}
		AddTokenTables,
{{- end}}
	}
}
`

const tplMigration001Accounts = `package migrations

import "github.com/jorgefuertes/martian-stack/pkg/database/migration"

// InitialSchema creates the initial accounts table
var InitialSchema = migration.Migration{
	Version:     1,
	Name:        "initial_schema",
	Description: "Create accounts table with indexes",
	Up: ` + "`" + `
CREATE TABLE IF NOT EXISTS accounts (
	id VARCHAR(36) PRIMARY KEY,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	last_login TIMESTAMP NULL,
	username VARCHAR(50) NOT NULL UNIQUE,
	name VARCHAR(120) NOT NULL,
	email VARCHAR(255) NOT NULL UNIQUE,
	enabled BOOLEAN NOT NULL DEFAULT true,
	role VARCHAR(10) NOT NULL DEFAULT 'user',
	crypted_password BLOB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_accounts_username ON accounts(username);
CREATE INDEX IF NOT EXISTS idx_accounts_email ON accounts(email);
CREATE INDEX IF NOT EXISTS idx_accounts_enabled ON accounts(enabled);
CREATE INDEX IF NOT EXISTS idx_accounts_role ON accounts(role);
` + "`" + `,
	Down: ` + "`" + `
DROP INDEX IF EXISTS idx_accounts_role;
DROP INDEX IF EXISTS idx_accounts_enabled;
DROP INDEX IF EXISTS idx_accounts_email;
DROP INDEX IF EXISTS idx_accounts_username;
DROP TABLE IF EXISTS accounts;
` + "`" + `,
}
`

const tplMigration001Example = `package migrations

import "github.com/jorgefuertes/martian-stack/pkg/database/migration"

// InitialSchema creates an example table
var InitialSchema = migration.Migration{
	Version:     1,
	Name:        "initial_schema",
	Description: "Create example table",
	Up: ` + "`" + `
CREATE TABLE IF NOT EXISTS items (
	id VARCHAR(36) PRIMARY KEY,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	name VARCHAR(255) NOT NULL,
	description TEXT
);
` + "`" + `,
	Down: ` + "`" + `
DROP TABLE IF EXISTS items;
` + "`" + `,
}
`

const tplMigration002 = `package migrations

import "github.com/jorgefuertes/martian-stack/pkg/database/migration"

// AddTokenTables creates tables for refresh tokens and password reset tokens
var AddTokenTables = migration.Migration{
	Version:     2,
	Name:        "add_token_tables",
	Description: "Create refresh_tokens and password_reset_tokens tables",
	Up: ` + "`" + `
CREATE TABLE IF NOT EXISTS refresh_tokens (
	id VARCHAR(36) PRIMARY KEY,
	user_id VARCHAR(36) NOT NULL,
	token_hash VARCHAR(64) NOT NULL UNIQUE,
	expires_at TIMESTAMP NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	revoked_at TIMESTAMP NULL,
	FOREIGN KEY (user_id) REFERENCES accounts(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
	id VARCHAR(36) PRIMARY KEY,
	user_id VARCHAR(36) NOT NULL,
	token_hash VARCHAR(64) NOT NULL UNIQUE,
	expires_at TIMESTAMP NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	used_at TIMESTAMP NULL,
	FOREIGN KEY (user_id) REFERENCES accounts(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_token_hash ON password_reset_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);
` + "`" + `,
	Down: ` + "`" + `
DROP INDEX IF EXISTS idx_password_reset_tokens_expires_at;
DROP INDEX IF EXISTS idx_password_reset_tokens_token_hash;
DROP INDEX IF EXISTS idx_password_reset_tokens_user_id;
DROP TABLE IF EXISTS password_reset_tokens;

DROP INDEX IF EXISTS idx_refresh_tokens_expires_at;
DROP INDEX IF EXISTS idx_refresh_tokens_token_hash;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP TABLE IF EXISTS refresh_tokens;
` + "`" + `,
}
`

const tplAuthRoutes = `package handlers

import (
	"github.com/jorgefuertes/martian-stack/pkg/auth"
	"github.com/jorgefuertes/martian-stack/pkg/server"
	"github.com/jorgefuertes/martian-stack/pkg/server/web"
)

// RegisterAuthRoutes registers authentication routes
func RegisterAuthRoutes(srv *server.Server, h *auth.Handlers, mw *auth.Middleware) {
	// Public auth routes
	pub := srv.Group("/auth")
	pub.Route(web.MethodPost, "/login", h.Login())
	pub.Route(web.MethodPost, "/refresh", h.Refresh())
	pub.Route(web.MethodPost, "/password-reset/request", h.RequestPasswordReset())
	pub.Route(web.MethodPost, "/password-reset", h.ResetPassword())

	// Protected auth routes (require authentication)
	priv := srv.Group("/auth", mw.RequireAuth())
	priv.Route(web.MethodPost, "/logout", h.Logout())
	priv.Route(web.MethodGet, "/me", h.Me())
}
`

const tplAdminRoutes = `package handlers

import (
	"net/http"

	"github.com/jorgefuertes/martian-stack/pkg/auth"
	"github.com/jorgefuertes/martian-stack/pkg/database/repository"
	"github.com/jorgefuertes/martian-stack/pkg/server"
	"github.com/jorgefuertes/martian-stack/pkg/server/ctx"
	"github.com/jorgefuertes/martian-stack/pkg/server/web"
)

// RegisterAdminRoutes registers admin panel routes
func RegisterAdminRoutes(
	srv *server.Server,
	mw *auth.Middleware,
	accountRepo *repository.SQLAccountRepository,
) {
	admin := srv.Group("/admin", mw.RequireAuth(), mw.RequireRole("admin"))

	// User management CRUD
	admin.Route(web.MethodGet, "/users", listUsers(accountRepo))
	admin.Route(web.MethodPost, "/users", createUser(accountRepo))
	admin.Route(web.MethodGet, "/users/{id}", getUser(accountRepo))
	admin.Route(web.MethodPut, "/users/{id}", updateUser(accountRepo))
	admin.Route(web.MethodDelete, "/users/{id}", deleteUser(accountRepo))
}

func listUsers(_ *repository.SQLAccountRepository) ctx.Handler {
	return func(c ctx.Ctx) error {
		// TODO: implement user listing
		return c.SendJSON(map[string]string{"message": "list users - not yet implemented"})
	}
}

func createUser(_ *repository.SQLAccountRepository) ctx.Handler {
	return func(c ctx.Ctx) error {
		// TODO: implement user creation
		return c.Error(http.StatusNotImplemented, "create user - not yet implemented")
	}
}

func getUser(_ *repository.SQLAccountRepository) ctx.Handler {
	return func(c ctx.Ctx) error {
		// TODO: implement get user by ID
		return c.Error(http.StatusNotImplemented, "get user - not yet implemented")
	}
}

func updateUser(_ *repository.SQLAccountRepository) ctx.Handler {
	return func(c ctx.Ctx) error {
		// TODO: implement user update
		return c.Error(http.StatusNotImplemented, "update user - not yet implemented")
	}
}

func deleteUser(_ *repository.SQLAccountRepository) ctx.Handler {
	return func(c ctx.Ctx) error {
		// TODO: implement user deletion
		return c.Error(http.StatusNotImplemented, "delete user - not yet implemented")
	}
}
`
