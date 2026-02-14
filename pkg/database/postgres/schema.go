package postgres

import (
	"context"

	"github.com/jorgefuertes/martian-stack/pkg/database"
)

// AccountsTableSchema is the SQL schema for the accounts table
const AccountsTableSchema = `
CREATE TABLE IF NOT EXISTS accounts (
	id UUID PRIMARY KEY,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	last_login TIMESTAMP,
	username VARCHAR(50) NOT NULL UNIQUE,
	name VARCHAR(120) NOT NULL,
	email VARCHAR(255) NOT NULL UNIQUE,
	enabled BOOLEAN NOT NULL DEFAULT true,
	role VARCHAR(10) NOT NULL DEFAULT 'user',
	crypted_password BYTEA NOT NULL,
	CONSTRAINT username_length CHECK (length(username) >= 4),
	CONSTRAINT name_length CHECK (length(name) >= 3),
	CONSTRAINT email_length CHECK (length(email) >= 5)
);

CREATE INDEX IF NOT EXISTS idx_accounts_username ON accounts(username);
CREATE INDEX IF NOT EXISTS idx_accounts_email ON accounts(email);
CREATE INDEX IF NOT EXISTS idx_accounts_enabled ON accounts(enabled);
CREATE INDEX IF NOT EXISTS idx_accounts_role ON accounts(role);

-- Trigger to automatically update updated_at
CREATE OR REPLACE FUNCTION update_accounts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
	NEW.updated_at = CURRENT_TIMESTAMP;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_accounts_updated_at ON accounts;
CREATE TRIGGER trigger_accounts_updated_at
	BEFORE UPDATE ON accounts
	FOR EACH ROW
	EXECUTE FUNCTION update_accounts_updated_at();
`

// CreateAccountsTable creates the accounts table
func CreateAccountsTable(db database.Database) error {
	_, err := db.Exec(context.Background(), AccountsTableSchema)
	return err
}
