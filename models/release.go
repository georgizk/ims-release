package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Release contains information about a release, which there are many of under a given Project.  It contains information
// about which chapter of manga the release was created for, which version of the release of said chapter it is for, and
// the status of the release of the chapter itself, which may not be final right away.
type Release struct {
	Id         uint32    `json:"id"`
	Identifier string    `json:"identifier"`
	Scanlator  string    `json:"scanlator"`
	Version    uint32    `json:"version"`
	Status     string    `json:"status"`
	ReleasedOn time.Time `json:"releasedOn"`
	ProjectID  uint32    `json:"-"`
}

type ReleaseStatus int

// ReleaseStatus pseudo-enum values
const (
	RStatusUnknown     ReleaseStatus = 0
	RStatusUnknownStr  string        = "unknown"
	RStatusReleased    ReleaseStatus = 1
	RStatusReleasedStr string        = "released"
	RStatusDraft       ReleaseStatus = 2
	RStatusDraftStr    string        = "draft"
)

func (s ReleaseStatus) String() string {
	switch s {
	case RStatusReleased:
		return RStatusReleasedStr
	case RStatusDraft:
		return RStatusDraftStr
	default:
		return RStatusUnknownStr
	}
}

func NewReleaseStatus(val string) ReleaseStatus {
	switch val {
	case RStatusReleasedStr:
		return RStatusReleased
	case RStatusDraftStr:
		return RStatusDraft
	default:
		return RStatusUnknown
	}
}

// Errors pertaining to the data in a Release or operations on Releases.
var (
	ErrInvalidReleaseStatus = errors.New("Invalid release status.")
	ErrNoSuchRelease        = errors.New("Could not find release.")
)

// Database queries for operations on Releases.
const (
	t_releases     string = "`releases`"
	Rc_id          string = "`id`"
	Rc_identifier  string = "`identifier`"
	Rc_version     string = "`version`"
	Rc_status      string = "`status`"
	Rc_released_on string = "`released_on`"
	Rc_project_id  string = "`project_id`"

	Rmax_len_identifier = 10
)

// NewRelease constructs a brand new Release instance, with a default state lacking information its (future) position in
// a database.
func NewRelease(projectId, version uint32, chapterName string) Release {
	return Release{
		0,
		chapterName,
		"ims", // @TODO make this variable
		version,
		RStatusDraftStr,
		time.Now(),
		projectId,
	}
}

// FindRelease attempts to lookup a release by ID.
func FindRelease(db *sql.DB, projectId uint32, releaseId uint32) (Release, error) {
	r := Release{}
	var s ReleaseStatus

	const query = "SELECT " + Rc_identifier + ", " + Rc_version + ", " +
		Rc_status + ", " + Rc_released_on + ", " + Rc_project_id +
		" FROM " + t_releases + " WHERE " + Rc_id + " = ? AND " + Rc_project_id + " = ?"

	row := db.QueryRow(query, releaseId, projectId)
	if row == nil {
		return Release{}, ErrNoSuchRelease
	}
	err := row.Scan(&r.Identifier, &r.Version, &s, &r.ReleasedOn, &r.ProjectID)
	if err != nil {
		return Release{}, err
	}
	r.Id = releaseId
	r.Status = s.String()
	r.Scanlator = "ims" // @TODO make this variable
	return r, nil
}

// ListReleases attempts to obtain a list of all of the releases in the database.
func ListReleases(db *sql.DB, projectId uint32) ([]Release, error) {
	releases := []Release{}

	const query = "SELECT " + Rc_id + ", " + Rc_identifier + ", " +
		Rc_version + ", " + Rc_status + ", " + Rc_released_on +
		" FROM " + t_releases + " WHERE " + Rc_project_id + " = ?"
	rows, err := db.Query(query, projectId)
	if err != nil {
		return []Release{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, version uint32
		var chapter string
		var status ReleaseStatus
		var released time.Time
		scanErr := rows.Scan(&id, &chapter, &version, &status, &released)
		if scanErr != nil {
			err = scanErr
		}
		releases = append(releases, Release{
			id,
			chapter,
			"ims", // @TODO make this variable
			version,
			status.String(),
			released,
			projectId,
		})
	}
	return releases, err
}

// Validate checks that the "status" of the project is one of the accepted ReleaseStatus values.
func (r *Release) Validate() error {
	if NewReleaseStatus(r.Status) != RStatusUnknown {
		return nil
	}
	if len(r.Identifier) > Rmax_len_identifier {
		return ErrFieldTooLong
	}
	return ErrInvalidReleaseStatus
}

// Save inserts the release into the database and updates its Id field.
func (r *Release) Save(db *sql.DB) error {
	validErr := r.Validate()
	if validErr != nil {
		return validErr
	}

	const query = "INSERT INTO " + t_releases + " (" +
		Rc_identifier + ", " + Rc_version + ", " + Rc_status + ", " +
		Rc_released_on + ", " + Rc_project_id + ") VALUES (?, ?, ?, ?, ?)"
	_, err := db.Exec(query, r.Identifier, r.Version, NewReleaseStatus(r.Status), r.ReleasedOn, r.ProjectID)
	if err != nil {
		return err
	}
	id, err := GetLastInsertId(db)
	if err != nil {
		return err
	}
	r.Id = uint32(id)
	return nil
}

// Update modifies all of the fields of a Release in place with whatever is currently in the struct.
func (r *Release) Update(db *sql.DB) error {

	validErr := r.Validate()
	if validErr != nil {
		return validErr
	}
	now := time.Now()
	const query = "UPDATE " + t_releases + " SET " +
		Rc_identifier + " = ?, " + Rc_version + " = ?," + Rc_status + " = ?," +
		Rc_released_on + " = ? WHERE " + Rc_id + " = ? LIMIT 1"
	_, err := db.Exec(query, r.Identifier, r.Version, NewReleaseStatus(r.Status), now, r.Id)
	r.ReleasedOn = now
	return err
}

// Delete removes the Release and all associated pages from the database.
func (r *Release) Delete(db *sql.DB) error {
	const query = "DELETE FROM " + t_releases + " WHERE " + Rc_id + " = ? LIMIT 1"
	_, err := db.Exec(query, r.Id)
	return err
}

func GenerateArchiveName(p Project, r Release) string {
	return fmt.Sprintf("%s - %s[%d][%s].zip", p.Shorthand, r.Identifier, r.Version, r.Scanlator)
}
