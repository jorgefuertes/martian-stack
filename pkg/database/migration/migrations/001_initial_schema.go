package migrations

import "github.com/jorgefuertes/martian-stack/pkg/database/migration"

// InitialSchema creates the initial accounts table
var InitialSchema = migration.Migration{
	Version:     20260212000001,
	Name:        "initial_schema",
	Description: "Create accounts table with indexes",
	Up: `
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

CREATE INDEX idx_accounts_username ON accounts(username);
CREATE INDEX idx_accounts_email ON accounts(email);
CREATE INDEX idx_accounts_enabled ON accounts(enabled);
CREATE INDEX idx_accounts_role ON accounts(role);
`,
	Down: `
DROP INDEX IF EXISTS idx_accounts_role;
DROP INDEX IF EXISTS idx_accounts_enabled;
DROP INDEX IF EXISTS idx_accounts_email;
DROP INDEX IF EXISTS idx_accounts_username;
DROP TABLE IF EXISTS accounts;
`,
}
