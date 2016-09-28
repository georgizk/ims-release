package models

import (
	"fmt"
	"time"
)

// ReleaseStatus is a type alias which will be used to create an enum of acceptable release status states.
type ReleaseStatus string

// ReleaseStatus pseudo-enum values
const (
	RStatusReleased ReleaseStatus = "released"
	RStatusDraft    ReleaseStatus = "draft"
)

var (
	RStatuses = []ReleaseStatus{RStatusReleased, RStatusDraft}
)

// Errors pertaining to the data in a Release or operations on Releases.
var (
	ErrInvalidReleaseStatus = fmt.Errorf("Invalid release status.")
)

// Release contains information about a release, which there are many of under a given Project.  It contains information
// about which chapter of manga the release was created for, which version of the release of said chapter it is for, and
// the status of the release of the chapter itself, which may not be final right away.
type Release struct {
	Id         string        `json:"id"`
	Chapter    string        `json:"chapter"`
	Version    int           `json:"version"`
	Status     ReleaseStatus `json:"status"`
	Checksum   string        `json:"checksum"`
	ReleasedOn time.Time     `json:"releasedOn:`
}

// Validate checks that the "status" of the project is one of the accepted ReleaseStatus values.
func (r Release) Validate() error {
	for _, status := range RStatuses {
		if r.Status == status {
			return nil
		}
	}
	return ErrInvalidReleaseStatus
}
