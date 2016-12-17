package models

import (
	"database/sql"
	"errors"
)

// Errors relevant to all models
var (
	ErrCouldNotGetID         error = errors.New("Could not get ID of last created row.")
	ErrOperationNotSupported error = errors.New("Operation not supported.")
	ErrFieldTooLong                = errors.New("A field value is too long.")
)

// Model should be implemented by all model types to provide functionality for data validation and persistence.
type Model interface {
	Validate() error
	Save(*sql.DB) error
	Update(*sql.DB) error
	Delete(*sql.DB) error
}
