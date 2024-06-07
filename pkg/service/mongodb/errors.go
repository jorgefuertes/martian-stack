package mongodb

import "errors"

// errors
var (
	ErrDbConn                = errors.New("error connecting to database")
	ErrDbClientNotFound      = errors.New("db client name not found")
	ErrDbValidation          = errors.New("validation errors")
	ErrDbZeroID              = errors.New("object ID is zero, cannot update")
	ErrDbNotZeroID           = errors.New("object ID is not zero, cannot create")
	ErrDbUnknownTypeID       = errors.New("ID should be and string or an objectID")
	ErrNotFoundInCache       = errors.New("key not found in cache")
	ErrMissingIDField        = errors.New("entity with no ID field")
	ErrTagNotFound           = errors.New("tag not found in struct")
	ErrReflectNotStruct      = errors.New("entity is not a struct")
	ErrMissingCreatedAtField = errors.New("entity has no created_at field")
	ErrMissingUpdatedAtField = errors.New("entity has no updated_at field")
)
