package models

import (
	"fmt"
	"time"
)

// ProjectStatus is a type alias which will be used to create an enum of acceptable project status states.
type ProjectStatus string

// ProjectStatus pseudo-enum values
const (
	PStatusPublished ProjectStatus = "published"
)

var (
	PStatuses = []ProjectStatus{PStatusPublished}
)

// Errors pertaining to the data in a Project or operations on Projects.
var (
	ErrInvalidProjectStatus = fmt.Errorf("Invalid project status.")
)

// Project contains information about a scanlation project, which has a human-readable name, a unique shorthand name,
// and a publishing status amongst other things.
type Project struct {
	Id          string        `json:"id"`
	Name        string        `json:"name"`
	Shorthand   string        `json:"projectName"`
	Description string        `json:"description"`
	Status      ProjectStatus `json:"status"`
	CreatedAt   time.Time     `json:"createdAt"`
}

// Validate checks that the "status" of the project is one of the accepted ProjectStatus values.
func (p Project) Validate() error {
	for _, status := range PStatuses {
		if p.Status == status {
			return nil
		}
	}
	return ErrInvalidProjectStatus
}
