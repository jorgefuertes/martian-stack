package sqlite

import (
	"context"

	"github.com/jorgefuertes/martian-stack/pkg/database"
)

// AccountsTableSchema is the SQL schema for the accounts table
const AccountsTableSchema = `
CREATE TABLE IF NOT EXISTS accounts (
	id TEXT PRIMARY KEY,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	last_login DATETIME,
	username TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE,
	enabled INTEGER NOT NULL DEFAULT 1,
	role TEXT NOT NULL DEFAULT 'user',
	crypted_password BLOB NOT NULL,
	CHECK(length(username) >= 4),
	CHECK(length(name) >= 3),
	CHECK(length(email) >= 5)
);

CREATE INDEX IF NOT EXISTS idx_accounts_username ON accounts(username);
CREATE INDEX IF NOT EXISTS idx_accounts_email ON accounts(email);
CREATE INDEX IF NOT EXISTS idx_accounts_enabled ON accounts(enabled);
CREATE INDEX IF NOT EXISTS idx_accounts_role ON accounts(role);
`

// CreateAccountsTable creates the accounts table
func CreateAccountsTable(db database.Database) error {
	_, err := db.Exec(context.Background(), AccountsTableSchema)
	return err
}
