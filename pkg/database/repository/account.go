package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"git.martianoids.com/martianoids/martian-stack/pkg/database"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/adapter"
	"github.com/google/uuid"
)

// SQLAccountRepository implements adapter.AccountRepository using SQL database
type SQLAccountRepository struct {
	db database.Database
}

// NewSQLAccountRepository creates a new SQL-based account repository
func NewSQLAccountRepository(db database.Database) *SQLAccountRepository {
	return &SQLAccountRepository{
		db: db,
	}
}

// Get retrieves an account by ID
func (r *SQLAccountRepository) Get(id string) (*adapter.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, created_at, updated_at, last_login, username, name, email, enabled, role, crypted_password
		FROM accounts
		WHERE id = ?
	`

	var acc adapter.Account
	var lastLogin sql.NullTime

	err := r.db.QueryRow(ctx, query, id).Scan(
		&acc.ID,
		&acc.CreatedAt,
		&acc.UpdatedAt,
		&lastLogin,
		&acc.Username,
		&acc.Name,
		&acc.Email,
		&acc.Enabled,
		&acc.Role,
		&acc.CryptedPassword,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, adapter.ErrAccountNotFound
		}
		return nil, err
	}

	if lastLogin.Valid {
		acc.LastLogin = lastLogin.Time
	}

	return &acc, nil
}

// Exists checks if an account with the given ID exists
func (r *SQLAccountRepository) Exists(id string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT COUNT(*) FROM accounts WHERE id = ?`

	var count int
	err := r.db.QueryRow(ctx, query, id).Scan(&count)
	if err != nil {
		return false
	}

	return count > 0
}

// GetByEmail retrieves an account by email
func (r *SQLAccountRepository) GetByEmail(email string) (*adapter.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, created_at, updated_at, last_login, username, name, email, enabled, role, crypted_password
		FROM accounts
		WHERE email = ?
	`

	var acc adapter.Account
	var lastLogin sql.NullTime

	err := r.db.QueryRow(ctx, query, email).Scan(
		&acc.ID,
		&acc.CreatedAt,
		&acc.UpdatedAt,
		&lastLogin,
		&acc.Username,
		&acc.Name,
		&acc.Email,
		&acc.Enabled,
		&acc.Role,
		&acc.CryptedPassword,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, adapter.ErrAccountNotFound
		}
		return nil, err
	}

	if lastLogin.Valid {
		acc.LastLogin = lastLogin.Time
	}

	return &acc, nil
}

// GetByUsername retrieves an account by username
func (r *SQLAccountRepository) GetByUsername(username string) (*adapter.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, created_at, updated_at, last_login, username, name, email, enabled, role, crypted_password
		FROM accounts
		WHERE username = ?
	`

	var acc adapter.Account
	var lastLogin sql.NullTime

	err := r.db.QueryRow(ctx, query, username).Scan(
		&acc.ID,
		&acc.CreatedAt,
		&acc.UpdatedAt,
		&lastLogin,
		&acc.Username,
		&acc.Name,
		&acc.Email,
		&acc.Enabled,
		&acc.Role,
		&acc.CryptedPassword,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, adapter.ErrAccountNotFound
		}
		return nil, err
	}

	if lastLogin.Valid {
		acc.LastLogin = lastLogin.Time
	}

	return &acc, nil
}

// Create creates a new account
func (r *SQLAccountRepository) Create(a *adapter.Account) error {
	if err := a.Validate(); err != nil {
		return err
	}

	if a.ID != "" {
		return adapter.ErrCannotCreateWithID
	}

	if len(a.CryptedPassword) == 0 {
		return adapter.ErrPasswordNotSet
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Generate new UUID
	a.ID = uuid.NewString()
	now := time.Now()
	a.CreatedAt = now
	a.UpdatedAt = now

	query := `
		INSERT INTO accounts (id, created_at, updated_at, last_login, username, name, email, enabled, role, crypted_password)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var lastLogin interface{}
	if !a.LastLogin.IsZero() {
		lastLogin = a.LastLogin
	}

	_, err := r.db.Exec(ctx, query,
		a.ID,
		a.CreatedAt,
		a.UpdatedAt,
		lastLogin,
		a.Username,
		a.Name,
		a.Email,
		a.Enabled,
		a.Role,
		a.CryptedPassword,
	)

	if err != nil {
		// Check for duplicate key errors
		if isDuplicateKeyError(err) {
			return database.ErrDuplicateKey
		}
		return err
	}

	return nil
}

// Update updates an existing account
func (r *SQLAccountRepository) Update(a *adapter.Account) error {
	if err := a.Validate(); err != nil {
		return err
	}

	if len(a.CryptedPassword) == 0 {
		return adapter.ErrPasswordNotSet
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a.UpdatedAt = time.Now()

	query := `
		UPDATE accounts
		SET updated_at = ?, last_login = ?, username = ?, name = ?, email = ?, enabled = ?, role = ?, crypted_password = ?
		WHERE id = ?
	`

	var lastLogin interface{}
	if !a.LastLogin.IsZero() {
		lastLogin = a.LastLogin
	}

	result, err := r.db.Exec(ctx, query,
		a.UpdatedAt,
		lastLogin,
		a.Username,
		a.Name,
		a.Email,
		a.Enabled,
		a.Role,
		a.CryptedPassword,
		a.ID,
	)

	if err != nil {
		if isDuplicateKeyError(err) {
			return database.ErrDuplicateKey
		}
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return adapter.ErrAccountNotFound
	}

	return nil
}

// Delete deletes an account by ID
func (r *SQLAccountRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `DELETE FROM accounts WHERE id = ?`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return adapter.ErrAccountNotFound
	}

	return nil
}

// isDuplicateKeyError checks if the error is a duplicate key constraint violation
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	// Check for common duplicate key error messages across different databases
	errMsg := err.Error()
	return contains(errMsg, "UNIQUE constraint failed") || // SQLite
		contains(errMsg, "duplicate key value") || // PostgreSQL
		contains(errMsg, "Duplicate entry") // MySQL/MariaDB
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
