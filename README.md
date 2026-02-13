# Martian Stack Framework

A complete, production-ready web framework for building modern applications in Go.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Tests](https://img.shields.io/badge/tests-56%2F56%20passing-success)](/)
[![License](https://img.shields.io/badge/license-Private-red)](/)

## 🚀 Features

### 🗄️ **Multi-Database Support**
- **SQLite** - Pure Go, in-memory & file-based
- **PostgreSQL** - Advanced features with pgx driver
- **MySQL/MariaDB** - Full compatibility
- **MongoDB** - Document database
- **Redis** - Caching & sessions

### 🔐 **JWT Authentication**
- Access & refresh tokens with rotation
- Role-based access control (RBAC)
- Login/Logout/Refresh/Password-reset handlers
- Stateless token validation
- Middleware protection
- Minimum 32-byte secret key enforcement

### 🔒 **Security**
- Secure cookies (HttpOnly, Secure, SameSite)
- Security headers middleware (CSP, X-Frame-Options, HSTS-ready)
- Request body size limits (1 MB default)
- Constant-time authentication comparisons
- bcrypt password hashing
- SHA256 token hashing (never store plaintext)
- SQL injection prevention via parameterized queries
- Anti-enumeration responses on auth endpoints

### 🔄 **Database Migrations**
- Version-based migrations
- Up/Down support
- Atomic transactions
- Status tracking
- Easy rollback

### 🌐 **HTTP Server**
- Custom routing system
- Middleware pipeline
- Session management
- CORS support
- Error handling
- HTMX templates

### 🧪 **Comprehensive Testing**
- 56 tests (100% passing)
- Integration tests
- In-memory databases
- Mock repositories

## 📦 Installation

```bash
go get git.martianoids.com/martianoids/martian-stack
```

## 🏃 Quick Start

### 1. Basic Server

```go
package main

import (
    "git.martianoids.com/martianoids/martian-stack/pkg/server"
    "git.martianoids.com/martianoids/martian-stack/pkg/server/ctx"
    "git.martianoids.com/martianoids/martian-stack/pkg/server/middleware"
    "git.martianoids.com/martianoids/martian-stack/pkg/service/logger"
    "os"
)

func main() {
    // Logger
    l := logger.New(os.Stdout, logger.TextFormat, logger.LevelDebug)

    // Server
    srv := server.New("localhost", "8080", 10)

    // Middleware
    srv.Use(
        middleware.NewSecurityHeaders(),
        middleware.NewCors(middleware.NewCorsOptions()),
        middleware.NewLog(l),
    )

    // Routes
    srv.Route("GET", "/", func(c ctx.Ctx) error {
        return c.SendString("Hello, Martian Stack!")
    })

    // Start
    l.From("main").Info("Starting server on :8080")
    srv.Start()
}
```

### 2. With Database

```go
package main

import (
    "context"
    "git.martianoids.com/martianoids/martian-stack/pkg/database/sqlite"
    "git.martianoids.com/martianoids/martian-stack/pkg/database/repository"
    "git.martianoids.com/martianoids/martian-stack/pkg/database/migration"
    "git.martianoids.com/martianoids/martian-stack/pkg/database/migration/migrations"
    "git.martianoids.com/martianoids/martian-stack/pkg/server"
)

func main() {
    // Database
    db, _ := sqlite.New(sqlite.DefaultConfig("./app.db"))
    defer db.Close()

    // Run migrations
    migrator := migration.New(db)
    migrator.RegisterMultiple(migrations.All())
    migrator.Up(context.Background())

    // Repository
    accountRepo := repository.NewSQLAccountRepository(db)

    // Server
    srv := server.New("localhost", "8080", 30)

    // Use repository in handlers...
    srv.Start()
}
```

### 3. Complete App with Authentication

```go
package main

import (
    "context"
    "os"

    "git.martianoids.com/martianoids/martian-stack/pkg/auth"
    "git.martianoids.com/martianoids/martian-stack/pkg/auth/jwt"
    "git.martianoids.com/martianoids/martian-stack/pkg/database/sqlite"
    "git.martianoids.com/martianoids/martian-stack/pkg/database/repository"
    "git.martianoids.com/martianoids/martian-stack/pkg/database/migration"
    "git.martianoids.com/martianoids/martian-stack/pkg/database/migration/migrations"
    "git.martianoids.com/martianoids/martian-stack/pkg/server"
    "git.martianoids.com/martianoids/martian-stack/pkg/server/middleware"
    "git.martianoids.com/martianoids/martian-stack/pkg/service/logger"
)

func main() {
    // Logger
    l := logger.New(os.Stdout, logger.TextFormat, logger.LevelInfo)

    // Database
    db, err := sqlite.New(sqlite.DefaultConfig("./app.db"))
    if err != nil {
        l.Error(err.Error())
        return
    }
    defer db.Close()

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

    // JWT Service (secret key must be at least 32 bytes)
    jwtCfg, err := jwt.DefaultConfig("your-secret-key-at-least-32-bytes!")
    if err != nil {
        l.Error("JWT config error: " + err.Error())
        return
    }
    jwtService := jwt.NewService(jwtCfg)

    // Auth
    authHandlers := auth.NewHandlers(accountRepo, jwtService, refreshTokenRepo, resetTokenRepo)
    authMw := auth.NewMiddleware(jwtService)

    // Server
    srv := server.New("localhost", "8080", 30)
    srv.Use(
        middleware.NewSecurityHeaders(),
        middleware.NewCors(middleware.NewCorsOptions()),
        middleware.NewLog(l),
    )

    // Public routes
    srv.Route("POST", "/auth/login", authHandlers.Login())
    srv.Route("POST", "/auth/refresh", authHandlers.Refresh())
    srv.Route("POST", "/auth/logout", authHandlers.Logout())

    // Protected routes
    srv.Route("GET", "/me",
        authMw.RequireAuth(),
        authHandlers.Me(),
    )

    // Admin routes
    srv.Route("GET", "/admin/dashboard",
        authMw.RequireAuth(),
        authMw.RequireRole("admin"),
        adminDashboardHandler,
    )

    // Start
    l.Info("Server starting on :8080")
    if err := srv.Start(); err != nil {
        l.Error(err.Error())
    }
}

func adminDashboardHandler(c ctx.Ctx) error {
    return c.SendJSON(map[string]string{
        "message": "Welcome to admin dashboard",
    })
}
```

## 📚 Documentation

### Database

#### Supported Databases

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

import "git.martianoids.com/martianoids/martian-stack/pkg/database/migration"

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

#### Login/Logout

##### Login Request

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

##### Response

```json
{
  "access_token": "eyJhbGci...",
  "refresh_token": "eyJhbGci...",
  "expires_at": "2026-02-12T20:15:00Z",
  "user": {
    "id": "uuid",
    "username": "johndoe",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "user"
  }
}
```

##### Refresh Token

```bash
curl -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGci..."
  }'
```

##### Authenticated Request

```bash
curl http://localhost:8080/me \
  -H "Authorization: Bearer eyJhbGci..."
```

#### Middleware

```go
authMw := auth.NewMiddleware(jwtService)

// Require authentication
srv.Route("GET", "/protected",
    authMw.RequireAuth(),
    protectedHandler,
)

// Require specific role
srv.Route("GET", "/admin",
    authMw.RequireAuth(),
    authMw.RequireRole("admin"),
    adminHandler,
)

// Require any of multiple roles
srv.Route("GET", "/moderator",
    authMw.RequireAuth(),
    authMw.RequireRole("admin", "moderator"),
    moderatorHandler,
)

// Optional authentication
srv.Route("GET", "/public",
    authMw.OptionalAuth(),
    publicHandler,
)
```

#### Context Helpers

```go
func myHandler(c ctx.Ctx) error {
    // Get user from context
    user, ok := auth.GetUserFromContext(c)
    if !ok {
        return c.Error(401, "Not authenticated")
    }

    // Get user ID
    userID, _ := auth.GetUserIDFromContext(c)

    // Get role
    role, _ := auth.GetRoleFromContext(c)

    // Check authentication
    if auth.IsAuthenticated(c) {
        // User is logged in
    }

    // Check role
    if auth.HasRole(c, "admin") {
        // User is admin
    }

    if auth.HasAnyRole(c, "admin", "moderator") {
        // User is admin or moderator
    }

    return c.SendJSON(user)
}
```

### Server & Routing

```go
srv := server.New("localhost", "8080", 30)

// Add middleware
srv.Use(
    middleware.NewSecurityHeaders(),   // X-Content-Type-Options, X-Frame-Options, CSP, etc.
    middleware.NewCors(middleware.NewCorsOptions()),
    middleware.NewLog(logger),
)

// Routes
srv.Route("GET", "/", homeHandler)
srv.Route("POST", "/users", createUserHandler)
srv.Route("GET", "/users/{id}", getUserHandler)
srv.Route("PUT", "/users/{id}", updateUserHandler)
srv.Route("DELETE", "/users/{id}", deleteUserHandler)

// Start server
srv.Start()
```

### Context API

```go
func handler(c ctx.Ctx) error {
    // Request
    method := c.Method()
    path := c.Path()
    ip := c.UserIP()
    param := c.Param("id")
    cookie := c.GetCookie("session")

    // Unmarshal body
    var req MyRequest
    c.UnmarshalBody(&req)

    // Response
    c.SendString("Hello")
    c.SendHTML("<h1>Hello</h1>")
    c.SendJSON(map[string]string{"msg": "hello"})

    // Status
    c.WithStatus(201).SendJSON(data)

    // Headers
    c.SetHeader("X-Custom", "value")
    c.SetCookie("token", "value", time.Hour)

    // Error
    return c.Error(404, "Not found")

    // Session
    session := c.Session()
    session.Data().Set("key", "value")

    // Store (request-scoped)
    c.Store().Set("key", "value")
    var val string
    c.Store().Get("key", &val)

    // Next middleware
    return c.Next()
}
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
go test ./... -cover

# Run specific package tests
go test ./pkg/auth/jwt/...
go test ./pkg/database/repository/...

# Verbose output
go test ./... -v
```

### Test Statistics

- **Total Tests:** 56
- **Pass Rate:** 100%
- **Coverage:** 100% (core components)

#### Breakdown

- JWT: 16 tests (includes secret key validation)
- SQL Repository: 18 tests
- Migration System: 13 tests
- SQLite: 5 tests
- Middleware: 4 tests (CORS, security headers)

## 🏗️ Project Structure

```text
martian-stack/
├── cmd/
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

## 🔧 Configuration

### Environment Variables

```bash
# Database
DB_DRIVER=sqlite           # sqlite, postgres, mysql
DB_PATH=./app.db          # SQLite
DB_HOST=localhost         # PostgreSQL/MySQL
DB_PORT=5432              # PostgreSQL/MySQL
DB_NAME=myapp
DB_USER=user
DB_PASSWORD=password

# JWT (secret must be at least 32 bytes)
JWT_SECRET=your-secret-key-at-least-32-bytes!
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h   # 7 days

# Server
SERVER_HOST=localhost
SERVER_PORT=8080
SERVER_TIMEOUT=30
```

## 📋 Requirements

- **Go:** 1.25 or higher
- **Redis:** Optional (for cache/sessions)
- **PostgreSQL:** Optional (if using PostgreSQL)
- **MySQL/MariaDB:** Optional (if using MySQL)

## 🤝 Contributing

This is a private project, but contributions from team members are welcome.

1. Create a feature branch (`git checkout -b feature/amazing-feature`)
2. Commit your changes (`git commit -m 'Add amazing feature'`)
3. Push to the branch (`git push origin feature/amazing-feature`)
4. Open a Pull Request

## 📝 License

Private - All rights reserved.

## 🙏 Credits

Built with ❤️ by the Martianoids team.

### Dependencies

- [pgx](https://github.com/jackc/pgx) - PostgreSQL driver
- [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql) - MySQL driver
- [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) - Pure Go SQLite
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt) - JWT implementation
- [go-playground/validator](https://github.com/go-playground/validator) - Validation
- [goht](https://github.com/stackus/goht) - HTMX templates

---

**Questions?** Contact the team or check the documentation in each package.
