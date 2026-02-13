package migrations

import "git.martianoids.com/martianoids/martian-stack/pkg/database/migration"

// All returns all available migrations
func All() []migration.Migration {
	return []migration.Migration{
		InitialSchema,
		AddTokenTables,
		// Add new migrations here
	}
}
