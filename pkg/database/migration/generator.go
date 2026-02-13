package migration

import (
	"fmt"
	"time"
)

// GenerateVersion generates a migration version based on current timestamp
// Format: YYYYMMDDHHmmss
func GenerateVersion() int64 {
	now := time.Now()
	return int64(now.Year())*10000000000 +
		int64(now.Month())*100000000 +
		int64(now.Day())*1000000 +
		int64(now.Hour())*10000 +
		int64(now.Minute())*100 +
		int64(now.Second())
}

// New creates a new migration with a generated version
func NewMigration(name, description string) Migration {
	return Migration{
		Version:     GenerateVersion(),
		Name:        name,
		Description: description,
	}
}

// NewWithVersion creates a new migration with a specific version
func NewWithVersion(version int64, name, description string) Migration {
	return Migration{
		Version:     version,
		Name:        name,
		Description: description,
	}
}

// Template generates a migration file template
func Template(name string) string {
	version := GenerateVersion()
	return fmt.Sprintf(`package migrations

import "git.martianoids.com/martianoids/martian-stack/pkg/database/migration"

// Migration%d_%s
var Migration%d = migration.Migration{
	Version: %d,
	Name:    "%s",
	Description: "Add description here",
	Up: ` + "`" + `
-- Add your SQL here
` + "`" + `,
	Down: ` + "`" + `
-- Add rollback SQL here
` + "`" + `,
}
`, version, name, version, version, name)
}
