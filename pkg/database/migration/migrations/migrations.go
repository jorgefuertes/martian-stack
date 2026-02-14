package migrations

import "github.com/jorgefuertes/martian-stack/pkg/database/migration"

// All returns all available migrations
func All() []migration.Migration {
	return []migration.Migration{
		InitialSchema,
		AddTokenTables,
		// Add new migrations here
	}
}
