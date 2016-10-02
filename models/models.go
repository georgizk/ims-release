package models

import (
	"database/sql"
	"errors"
)

// SQL queries that can be used by all models.
const (
	QLastInsertID string = `select LAST_INSERT_ID();`
)

// Errors relevant to all models
var (
	ErrCouldNotGetID error = errors.New("Could not get ID of last created row.")
)

// Model should be implemented by all model types to provide functionality for data validation and persistence.
type Model interface {
	Validate() error
	Save(*sql.DB) error
	Update(*sql.DB) error
	Delete(*sql.DB) error
}
