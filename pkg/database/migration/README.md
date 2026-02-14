# Database Migrations

Sistema de migraciones para gestionar el schema de la base de datos de forma versionada.

## Características

- ✅ Versionado basado en timestamps (YYYYMMDDHHmmss)
- ✅ Soporte para Up/Down migrations
- ✅ Transacciones automáticas
- ✅ Tracking de migraciones aplicadas
- ✅ Compatible con SQLite, PostgreSQL, MySQL/MariaDB
- ✅ Estado de migraciones (aplicadas/pendientes)

## Uso Básico

### 1. Crear una migración

```go
package migrations

import "github.com/jorgefuertes/martian-stack/pkg/database/migration"

var AddUserRoles = migration.Migration{
    Version:     20260212120000,
    Name:        "add_user_roles",
    Description: "Add roles column to users table",
    Up: `
        ALTER TABLE users ADD COLUMN role VARCHAR(20) DEFAULT 'user';
        CREATE INDEX idx_users_role ON users(role);
    `,
    Down: `
        DROP INDEX idx_users_role;
        ALTER TABLE users DROP COLUMN role;
    `,
}
```

### 2. Registrar y ejecutar migraciones

```go
package main

import (
    "context"
    "log"

    "github.com/jorgefuertes/martian-stack/pkg/database"
    "github.com/jorgefuertes/martian-stack/pkg/database/migration"
    "github.com/jorgefuertes/martian-stack/pkg/database/migration/migrations"
    "github.com/jorgefuertes/martian-stack/pkg/database/sqlite"
)

func main() {
    // Conectar a la base de datos
    db, err := sqlite.New(sqlite.DefaultConfig("./app.db"))
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Crear migrator
    migrator := migration.New(db)

    // Registrar todas las migraciones
    migrator.RegisterMultiple(migrations.All())

    // Ejecutar migraciones pendientes
    ctx := context.Background()
    if err := migrator.Up(ctx); err != nil {
        log.Fatal(err)
    }

    log.Println("Migrations applied successfully")
}
```

### 3. Ver estado de migraciones

```go
status, err := migrator.Status(ctx)
if err != nil {
    log.Fatal(err)
}

for _, m := range status {
    if m.AppliedAt != nil {
        fmt.Printf("✓ %d - %s (applied: %s)\n", m.Version, m.Name, m.AppliedAt)
    } else {
        fmt.Printf("✗ %d - %s (pending)\n", m.Version, m.Name)
    }
}
```

### 4. Rollback

```go
// Rollback última migración
if err := migrator.Down(ctx); err != nil {
    log.Fatal(err)
}

// Rollback hasta una versión específica
if err := migrator.DownTo(ctx, 20260212000001); err != nil {
    log.Fatal(err)
}
```

## Generar Nueva Migración

Usa el helper para generar el template:

```go
package main

import (
    "fmt"
    "github.com/jorgefuertes/martian-stack/pkg/database/migration"
)

func main() {
    template := migration.Template("add_email_verification")
    fmt.Println(template)
}
```

O crea directamente:

```go
var MyMigration = migration.NewMigration(
    "my_migration_name",
    "Description of what this migration does",
)
MyMigration.Up = `-- SQL here`
MyMigration.Down = `-- Rollback SQL here`
```

## Convenciones

1. **Nombrado**: Usa snake_case para nombres de migración
2. **Versionado**: Usa timestamps (YYYYMMDDHHmmss) para versiones
3. **Organización**: Una migración por archivo en `pkg/database/migration/migrations/`
4. **Registro**: Agrega todas las migraciones a `migrations.All()`
5. **Rollback**: Siempre provee un script `Down` cuando sea posible

## Estructura de Archivos

```text
pkg/database/migration/
├── migration.go              # Core migrator
├── generator.go              # Helpers para generar migraciones
├── README.md                 # Esta documentación
└── migrations/
    ├── migrations.go         # Registry de todas las migraciones
    ├── 001_initial_schema.go # Primera migración
    ├── 002_add_feature.go    # Segunda migración
    └── ...
```

## Tabla de Tracking

El sistema crea automáticamente la tabla `schema_migrations`:

```sql
CREATE TABLE schema_migrations (
    version BIGINT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

## Consideraciones

- Las migraciones se ejecutan en **transacciones** (atomic)
- Si una migración falla, se hace **rollback automático**
- Las migraciones se aplican en **orden de versión** (ascendente)
- No modifiques migraciones ya aplicadas en producción
- Usa migraciones nuevas para corregir errores en migraciones anteriores
