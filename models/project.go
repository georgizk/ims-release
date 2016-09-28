package models

import (
	"time"
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
	for _, status := range Statuses {
		if p.Status == status {
			return nil
		}
	}
	return ErrInvalidProjectStatus
}
