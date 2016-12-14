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

// InitDB initializes all of the database tables, only creating them if they do not already exist.
func InitDB(db *sql.DB) error {
	return nil
}

func GetLastInsertId(db *sql.DB) (uint64, error) {
	// SQL queries that can be used by all models.
	const query = "SELECT LAST_INSERT_ID();"
	row := db.QueryRow(query)
	if row == nil {
		return 0, ErrCouldNotGetID
	}
	var id uint64
	err := row.Scan(&id)
	return id, err
}
