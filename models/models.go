package models

// Model should be implemented by all model types to provide functionality for data validation and persistence.
type Model interface {
	Validate() error
}
