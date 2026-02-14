package mysql

import (
	"context"

	"github.com/jorgefuertes/martian-stack/pkg/database"
)

// AccountsTableSchema is the SQL schema for the accounts table
const AccountsTableSchema = `
CREATE TABLE IF NOT EXISTS accounts (
	id CHAR(36) PRIMARY KEY,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	last_login DATETIME NULL,
	username VARCHAR(50) NOT NULL UNIQUE,
	name VARCHAR(120) NOT NULL,
	email VARCHAR(255) NOT NULL UNIQUE,
	enabled BOOLEAN NOT NULL DEFAULT true,
	role VARCHAR(10) NOT NULL DEFAULT 'user',
	crypted_password VARBINARY(255) NOT NULL,
	CONSTRAINT username_length CHECK (CHAR_LENGTH(username) >= 4),
	CONSTRAINT name_length CHECK (CHAR_LENGTH(name) >= 3),
	CONSTRAINT email_length CHECK (CHAR_LENGTH(email) >= 5),
	INDEX idx_accounts_username (username),
	INDEX idx_accounts_email (email),
	INDEX idx_accounts_enabled (enabled),
	INDEX idx_accounts_role (role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

// CreateAccountsTable creates the accounts table
func CreateAccountsTable(db database.Database) error {
	_, err := db.Exec(context.Background(), AccountsTableSchema)
	return err
}
