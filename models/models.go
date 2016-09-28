package models

import (
	"fmt"
	"strings"
)

// Model should be implemented by all model types to provide functionality for data validation and persistence.
type Model interface {
	Validate() error
}

// ProjectStatus is a type alias which will be used to create an enum of acceptable project status states.
type ProjectStatus string

// ProjectStatus pseudo-enum values
const (
	StatusPublished ProjectStatus = "published"

	Statuses = []ProjectStatus{StatusPublished}
)

// Model-related errors
var (
	ErrInvalidProjectStatus = fmt.Errorf("Project status must be one of the following: %s\n", strings.Join([]string(Statuses), ", "))
)
