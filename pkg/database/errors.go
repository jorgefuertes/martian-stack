package database

import "errors"

var (
	// ErrNotFound is returned when a record is not found
	ErrNotFound = errors.New("record not found")

	// ErrDuplicateKey is returned when trying to insert a duplicate key
	ErrDuplicateKey = errors.New("duplicate key")

	// ErrInvalidConfig is returned when database configuration is invalid
	ErrInvalidConfig = errors.New("invalid database configuration")

	// ErrConnectionFailed is returned when database connection fails
	ErrConnectionFailed = errors.New("database connection failed")

	// ErrTransactionFailed is returned when transaction fails
	ErrTransactionFailed = errors.New("transaction failed")

	// ErrMigrationFailed is returned when migration fails
	ErrMigrationFailed = errors.New("migration failed")
)
