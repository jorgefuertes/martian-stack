# Martian Stack Framework

A complete, production-ready web framework for building modern applications in Go.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Tests](https://img.shields.io/badge/tests-184%20passing-success)](/)
[![License](https://img.shields.io/badge/license-GPL--3.0-blue)](LICENSE.md)

## Features

### Multi-Database Support

- **SQLite** - Pure Go, in-memory & file-based
- **PostgreSQL** - Advanced features with pgx driver
- **MySQL/MariaDB** - Full compatibility
- **Redis** - Caching & sessions

### JWT Authentication

- Access & refresh tokens with rotation
- Role-based access control (RBAC)
- Login/Logout/Refresh/Password-reset handlers
- Stateless token validation
- Middleware protection (RequireAuth, RequireRole, OptionalAuth)
- Minimum 32-byte secret key enforcement

### Security

- Secure cookies (HttpOnly, Secure, SameSite)
- Security headers middleware (CSP, X-Frame-Options, HSTS-ready)
- Request body size limits (1 MB default)
- Constant-time authentication comparisons
- bcrypt password hashing
- SHA256 token hashing (never store plaintext)
- SQL injection prevention via parameterized queries
- Anti-enumeration responses on auth endpoints
- Rate limiting middleware (per-IP, fixed-window)
- Panic recovery middleware

### Database Migrations

- Version-based migrations
- Up/Down support
- Atomic transactions
- Status tracking
- Easy rollback

### HTTP Server

- Custom routing with path parameters (`:param` or `{param}`)
- Route groups with shared prefix and middleware
- Middleware pipeline (server-level and group-level)
- Session management with flash messages
- CORS support
- Content negotiation (Accept header parsing with quality values)
- Static file serving (directory and `embed.FS`)
- TLS/HTTPS support
- Graceful shutdown with signal handling (SIGINT/SIGTERM)
- Per-route timeout middleware
- Request ID propagation (X-Request-ID)
- HTTP redirects
- Error handling with content negotiation (JSON/HTML/text)
- HTMX templates (goht)

### Caching

- In-memory cache with automatic expiration
- Redis cache
- Common interface for both backends

## Installation

```bash
go get github.com/jorgefuertes/martian-stack
```

## Project Generator

Scaffold a new project interactively:

```bash
go install github.com/jorgefuertes/martian-stack/cmd/martian-stack@latest
martian-stack
```

The TUI guides you through choosing database, cache, middlewares,
JWT authentication, and admin panel options. It generates a ready-to-run
project with all the boilerplate wired up.

### Non-interactive mode

```bash
martian-stack -y \
  -name=myapp \
  -module=github.com/user/myapp \
  -output=./myapp \
  -db=sqlite \
  -cache=memory \
  -auth \
  -admin \
  -middlewares=cors,logging,security,recovery,ratelimit,timeout
```

Available flags:

| Flag | Default | Description |
|---|---|---|
| `-name` | | Project name (required) |
| `-module` | | Go module path (required) |
| `-output` | | Output directory (required) |
| `-db` | `sqlite` | Database: `sqlite`, `postgres`, `mysql`, `none` |
| `-cache` | `memory` | Cache: `memory`, `redis` |
| `-auth` | `true` | Include JWT authentication |
| `-admin` | `true` | Include admin panel with user CRUD |
| `-middlewares` | `cors,logging,security,recovery` | Comma-separated middleware list |
| `-y` | `false` | Skip confirmation (non-interactive) |

## Quick Start

### 1. Basic Server

```go
package main

import (
    "os"

    "github.com/jorgefuertes/martian-stack/pkg/server"
    "github.com/jorgefuertes/martian-stack/pkg/server/ctx"
    "github.com/jorgefuertes/martian-stack/pkg/server/middleware"
    "github.com/jorgefuertes/martian-stack/pkg/server/web"
    "github.com/jorgefuertes/martian-stack/pkg/service/logger"
)

func main() {
    l := logger.New(os.Stdout, logger.TextFormat, logger.LevelDebug)

    srv := server.New("localhost", "8080", 10)
    srv.Use(
        middleware.NewRecovery(),
        middleware.NewSecurityHeaders(),
        middleware.NewCors(middleware.NewCorsOptions()),
        middleware.NewLog(l),
    )

    srv.Route(web.MethodGet, "/", func(c ctx.Ctx) error {
        return c.SendString("Hello, Martian Stack!")
    })

    // Blocks until SIGINT/SIGTERM, then graceful shutdown
    srv.ListenAndShutdown()
}
```

### 2. Route Groups

```go
// Public routes
srv.Route(web.MethodPost, "/auth/login", authHandlers.Login())

// API group with auth middleware
api := srv.Group("/api/v1", authMw.RequireAuth())
api.Route(web.MethodGet, "/users", listUsers)
api.Route(web.MethodPost, "/users", createUser)

// Admin sub-group
admin := api.Group("/admin", authMw.RequireRole("admin"))
admin.Route(web.MethodGet, "/stats", getStats)
```

### 3. Static Files

```go
// Serve from directory
srv.Static("/static/", "./public")

// Serve from embedded filesystem
//go:embed static
var staticFiles embed.FS
srv.StaticFS("/static/", staticFiles)
```

### 4. Rate Limiting

```go
// Global rate limiter: 60 requests per minute per IP
srv.Use(middleware.NewRateLimit(middleware.DefaultRateLimitConfig()))

// Custom config
cfg := middleware.RateLimitConfig{
    Max:    10,
    Window: time.Minute,
}
srv.Use(middleware.NewRateLimit(cfg))
```

### 5. Per-Route Timeout

```go
import "time"

// Apply timeout to specific route group
slow := srv.Group("/reports", middleware.NewTimeout(30*time.Second))
slow.Route(web.MethodGet, "/generate", generateReport)
```

### 6. TLS/HTTPS

```go
// Simple TLS
srv.StartTLS("cert.pem", "key.pem")

// TLS with graceful shutdown
srv.ListenAndShutdownTLS("cert.pem", "key.pem", func() {
    db.Close()
})

// Custom TLS config
srv.SetTLSConfig(&tls.Config{
    MinVersion: tls.VersionTLS13,
})
srv.StartTLS("cert.pem", "key.pem")
```

### 7. Complete App with Authentication

```go
package main

import (
    "context"
    "os"

    "github.com/jorgefuertes/martian-stack/pkg/auth"
    authjwt "github.com/jorgefuertes/martian-stack/pkg/auth/jwt"
    "github.com/jorgefuertes/martian-stack/pkg/database/sqlite"
    "github.com/jorgefuertes/martian-stack/pkg/database/repository"
    "github.com/jorgefuertes/martian-stack/pkg/database/migration"
    "github.com/jorgefuertes/martian-stack/pkg/database/migration/migrations"
    "github.com/jorgefuertes/martian-stack/pkg/server"
    "github.com/jorgefuertes/martian-stack/pkg/server/ctx"
    "github.com/jorgefuertes/martian-stack/pkg/server/middleware"
    "github.com/jorgefuertes/martian-stack/pkg/server/web"
    "github.com/jorgefuertes/martian-stack/pkg/service/logger"
)

func main() {
    l := logger.New(os.Stdout, logger.TextFormat, logger.LevelInfo)

    // Database
    db, err := sqlite.New(sqlite.DefaultConfig("./app.db"))
    if err != nil {
        l.Error(err.Error())
        return
    }

    // Migrations
    migrator := migration.New(db)
    migrator.RegisterMultiple(migrations.All())
    if err := migrator.Up(context.Background()); err != nil {
        l.Error("Migration failed: " + err.Error())
        return
    }

    // Repositories
    accountRepo := repository.NewSQLAccountRepository(db)
    refreshTokenRepo := repository.NewSQLRefreshTokenRepository(db)
    resetTokenRepo := repository.NewSQLPasswordResetTokenRepository(db)

    // JWT (secret key must be at least 32 bytes)
    jwtCfg, err := authjwt.DefaultConfig("your-secret-key-at-least-32-bytes!")
    if err != nil {
        l.Error(err.Error())
        return
    }
    jwtService := authjwt.NewService(jwtCfg)

    // Auth
    authHandlers := auth.NewHandlers(accountRepo, jwtService, refreshTokenRepo, resetTokenRepo)
    authMw := auth.NewMiddleware(jwtService)

    // Server
    srv := server.New("localhost", "8080", 30)
    srv.Use(
        middleware.NewRecovery(),
        middleware.NewSecurityHeaders(),
        middleware.NewRateLimit(middleware.DefaultRateLimitConfig()),
        middleware.NewCors(middleware.NewCorsOptions()),
        middleware.NewLog(l),
    )

    // Public routes
    srv.Route(web.MethodPost, "/auth/login", authHandlers.Login())
    srv.Route(web.MethodPost, "/auth/refresh", authHandlers.Refresh())

    // Protected routes
    api := srv.Group("/api", authMw.RequireAuth())
    api.Route(web.MethodPost, "/auth/logout", authHandlers.Logout())
    api.Route(web.MethodGet, "/me", authHandlers.Me())

    // Admin routes
    admin := api.Group("/admin", authMw.RequireRole("admin"))
    admin.Route(web.MethodGet, "/dashboard", func(c ctx.Ctx) error {
        return c.SendJSON(map[string]string{"message": "admin dashboard"})
    })

    // Graceful shutdown: close DB when server stops
    srv.ListenAndShutdown(func() {
        db.Close()
    })
}
```

## Documentation

### Database

#### SQLite (recommended for development)

```go
db, _ := sqlite.New(sqlite.DefaultConfig("./app.db"))

// In-memory
db, _ := sqlite.NewInMemory()
```

#### PostgreSQL

```go
db, _ := postgres.New(&postgres.Config{
    Host:     "localhost",
    Port:     5432,
    User:     "postgres",
    Password: "password",
    Database: "myapp",
    SSLMode:  "disable",
})
```

#### MySQL/MariaDB

```go
db, _ := mysql.New(&mysql.Config{
    Host:     "localhost",
    Port:     3306,
    User:     "root",
    Password: "password",
    Database: "myapp",
})
```

#### Migrations

Create a migration:

```go
// pkg/database/migration/migrations/002_add_users.go
package migrations

import "github.com/jorgefuertes/martian-stack/pkg/database/migration"

var AddUsers = migration.Migration{
    Version:     20260213000001,
    Name:        "add_users_table",
    Description: "Create users table",
    Up: `
        CREATE TABLE users (
            id VARCHAR(36) PRIMARY KEY,
            name VARCHAR(120) NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    `,
    Down: `
        DROP TABLE users;
    `,
}
```

Register and run:

```go
migrator := migration.New(db)
migrator.Register(AddUsers)
migrator.Up(context.Background())

// Check status
status, _ := migrator.Status(context.Background())

// Rollback
migrator.Down(context.Background())
migrator.DownTo(context.Background(), 20260213000001)
```

See [Migration Guide](pkg/database/migration/README.md) for more details.

#### Repository Pattern

```go
repo := repository.NewSQLAccountRepository(db)

// Create
account := &adapter.Account{
    Username: "johndoe",
    Name:     "John Doe",
    Email:    "john@example.com",
    Role:     "user",
    Enabled:  true,
}
account.SetPassword("secure-password")
repo.Create(account)

// Read
user, _ := repo.Get(id)
user, _ := repo.GetByEmail("john@example.com")
user, _ := repo.GetByUsername("johndoe")
exists := repo.Exists(id)

// Update
account.Name = "John Smith"
repo.Update(account)

// Delete
repo.Delete(id)
```

### Authentication

#### JWT Tokens

```go
// Setup (secret key must be at least 32 bytes)
jwtCfg, err := jwt.DefaultConfig("your-secret-key-at-least-32-bytes!")
if err != nil {
    log.Fatal(err) // jwt.ErrWeakSecretKey
}
jwtService := jwt.NewService(jwtCfg)

// Generate tokens
accessToken, _ := jwtService.GenerateAccessToken(
    userID, username, email, role,
)
refreshToken, _ := jwtService.GenerateRefreshToken(userID)

// Validate
claims, err := jwtService.ValidateToken(token)
if err == jwt.ErrExpiredToken {
    // Token expired
}

// Check expiry
isExpired := jwtService.IsExpired(token)
expiryTime, _ := jwtService.GetExpiryTime(token)
```

#### Context Helpers

```go
func myHandler(c ctx.Ctx) error {
    user, ok := auth.GetUserFromContext(c)
    userID, _ := auth.GetUserIDFromContext(c)
    role, _ := auth.GetRoleFromContext(c)

    if auth.IsAuthenticated(c) { /* ... */ }
    if auth.HasRole(c, "admin") { /* ... */ }
    if auth.HasAnyRole(c, "admin", "moderator") { /* ... */ }

    return c.SendJSON(user)
}
```

### Server & Routing

```go
srv := server.New("localhost", "8080", 30)

// Server-level middleware
srv.Use(
    middleware.NewRecovery(),           // panic recovery
    middleware.NewSecurityHeaders(),    // security headers
    middleware.NewRateLimit(cfg),       // rate limiting
    middleware.NewCors(corsOpts),       // CORS
    middleware.NewLog(logger),          // request logging
)

// Simple routes
srv.Route(web.MethodGet, "/", homeHandler)
srv.Route(web.MethodGet, "/users/{id}", getUserHandler)

// Route groups
api := srv.Group("/api/v1", authMiddleware)
api.Route(web.MethodGet, "/users", listUsersHandler)

// Static files
srv.Static("/assets/", "./public")

// Start options
srv.Start()                                    // plain HTTP
srv.StartTLS("cert.pem", "key.pem")           // HTTPS
srv.ListenAndShutdown()                        // HTTP + graceful shutdown
srv.ListenAndShutdownTLS("cert.pem", "key.pem") // HTTPS + graceful shutdown
```

### Context API

```go
func handler(c ctx.Ctx) error {
    // Request info
    method := c.Method()
    path := c.Path()
    ip := c.UserIP()              // IPv4 and IPv6 safe
    param := c.Param("id")       // path or query param
    cookie := c.GetCookie("session")
    reqID := c.ID()               // unique request ID (UUID)

    // Unmarshal body
    var req MyRequest
    c.UnmarshalBody(&req)         // JSON decode with 1MB limit

    // Unmarshal + validate (uses go-playground/validator tags)
    var req ValidatedRequest
    if err := c.UnmarshalAndValidate(&req); err != nil {
        return c.Error(400, err.Error())
    }

    // Response
    c.SendString("Hello")
    c.SendHTML("<h1>Hello</h1>")
    c.SendJSON(map[string]string{"msg": "hello"})
    c.WithStatus(201).SendJSON(data)

    // Redirect
    c.Redirect(http.StatusFound, "/new-location")

    // Headers
    c.SetHeader("X-Custom", "value")       // replaces existing value
    c.AddHeader("X-Custom", "extra")       // appends value
    c.SetCookie("token", "value", time.Hour)

    // Content negotiation
    c.AcceptsJSON()       // true if Accept header includes application/json
    c.AcceptsHTML()       // true if Accept header includes text/html
    c.AcceptsPlainText()  // true if Accept header includes text/plain

    // Error response
    return c.Error(404, "Not found")

    // Session & store
    session := c.Session()
    c.Store().Set("key", "value")

    // Context propagation
    c = c.WithContext(reqCtx) // for deadlines/cancellation

    // Middleware chain
    return c.Next()
}
```

### Middleware Reference

| Middleware | Description |
|---|---|
| `NewRecovery()` | Recovers from panics, returns 500 |
| `NewSecurityHeaders()` | Sets CSP, X-Frame-Options, X-Content-Type-Options, etc. |
| `NewRateLimit(cfg)` | Per-IP rate limiting with fixed-window counter |
| `NewCors(opts)` | CORS with preflight support |
| `NewLog(logger)` | Request logging with status codes |
| `NewBasicAuth(user, pass)` | HTTP Basic Authentication (constant-time) |
| `NewTimeout(duration)` | Per-route request timeout |
| `NewSession(cache, autostart)` | Session management backed by cache |

## Testing

```bash
# Run all tests
make test

# Run tests with clean cache
make test-clean

# Lint
make lint

# Run specific package tests
go test ./pkg/auth/jwt/...
go test ./pkg/database/repository/...
```

## Project Structure

```text
martian-stack/
├── cmd/
│   ├── martian-stack/       # Project generator CLI
│   └── testserver/          # Example server
├── pkg/
│   ├── auth/                # Authentication system
│   │   ├── jwt/            # JWT service
│   │   ├── handlers.go     # Login/Logout handlers
│   │   └── middleware.go   # Auth middleware
│   ├── database/            # Database layer
│   │   ├── sqlite/         # SQLite driver
│   │   ├── postgres/       # PostgreSQL driver
│   │   ├── mysql/          # MySQL/MariaDB driver
│   │   ├── repository/     # Repository implementations
│   │   └── migration/      # Migration system
│   ├── server/              # HTTP server
│   │   ├── ctx/            # Request context
│   │   ├── middleware/     # Middleware
│   │   ├── session/        # Session management
│   │   └── view/           # HTMX templates
│   ├── service/
│   │   ├── cache/          # Redis & memory cache
│   │   └── logger/         # Structured logger
│   └── store/               # Key-value store
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Requirements

- **Go:** 1.25 or higher
- **Redis:** Optional (for cache/sessions)
- **PostgreSQL:** Optional
- **MySQL/MariaDB:** Optional

## License

This project is licensed under the [GNU General Public License v3.0](LICENSE.md).

## Credits

Built by [Jorge Fuertes](https://github.com/jorgefuertes).

### Dependencies

- [pgx](https://github.com/jackc/pgx) - PostgreSQL driver
- [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql) - MySQL driver
- [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) - Pure Go SQLite
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt) - JWT implementation
- [go-playground/validator](https://github.com/go-playground/validator) - Validation
- [go-redis](https://github.com/redis/go-redis) - Redis client
- [goht](https://github.com/stackus/goht) - HTMX templates
- [huh](https://github.com/charmbracelet/huh) - Interactive terminal forms
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
