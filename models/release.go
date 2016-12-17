package models

import (
	"errors"
	"fmt"
	"ims-release/database"
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
func NewRelease(p Project, identifier string, version uint32, status string, tm time.Time) Release {
	return Release{
		0,
		identifier,
		"ims", // @TODO make this variable
		version,
		status,
		tm,
		p.Id,
	}
}

// FindRelease attempts to lookup a release by ID.
func FindRelease(db database.DB, project Project, releaseId uint32) (Release, error) {
	r := Release{}
	var s ReleaseStatus

	const query = "SELECT " + Rc_identifier + ", " + Rc_version + ", " +
		Rc_status + ", " + Rc_released_on +
		" FROM " + t_releases + " WHERE " + Rc_id + " = ? AND " + Rc_project_id + " = ?"

	row := db.QueryRow(query, releaseId, project.Id)
	err := row.Scan(&r.Identifier, &r.Version, &s, &r.ReleasedOn)

	if err == database.ErrNoRows {
		return Release{}, ErrNoSuchRelease
	} else if err != nil {
		return Release{}, err
	}

	r.Id = releaseId
	r.ProjectID = project.Id
	r.Status = s.String()
	r.Scanlator = "ims" // @TODO make this variable
	return r, nil
}

// ListReleases attempts to obtain a list of all of the releases in the database.
func ListReleases(db database.DB, project Project) ([]Release, error) {
	releases := []Release{}

	const query = "SELECT " + Rc_id + ", " + Rc_identifier + ", " +
		Rc_version + ", " + Rc_status + ", " + Rc_released_on +
		" FROM " + t_releases + " WHERE " + Rc_project_id + " = ?"
	rows, err := db.Query(query, project.Id)
	if err != nil {
		return releases, err
	}
	defer rows.Close()
	for rows.Next() {
		// @TODO make scanlator variable
		release := Release{ProjectID: project.Id, Scanlator: "ims"}
		var status ReleaseStatus
		err = rows.Scan(&release.Id, &release.Identifier, &release.Version, &status, &release.ReleasedOn)
		if err != nil {
			return releases, err
		}

		release.Status = status.String()
		releases = append(releases, release)
	}
	err = rows.Err()
	return releases, err
}

// Validate checks that the "status" of the project is one of the accepted ReleaseStatus values.
func (r *Release) Validate() error {
	if NewReleaseStatus(r.Status) == RStatusUnknown {
		return ErrInvalidReleaseStatus
	}
	if len(r.Identifier) > Rmax_len_identifier {
		return ErrFieldTooLong
	}
	return nil
}

// Save inserts the release into the database and updates its Id field.
func (r *Release) Save(db database.DB) error {
	validErr := r.Validate()
	if validErr != nil {
		return validErr
	}

	const query = "INSERT INTO " + t_releases + " (" +
		Rc_identifier + ", " + Rc_version + ", " + Rc_status + ", " +
		Rc_released_on + ", " + Rc_project_id + ") VALUES (?, ?, ?, ?, ?)"
	res, err := db.Exec(query, r.Identifier, r.Version, NewReleaseStatus(r.Status), r.ReleasedOn, r.ProjectID)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	r.Id = uint32(id)
	return nil
}

// Update modifies all of the fields of a Release in place with whatever is currently in the struct.
func (r *Release) Update(db database.DB) error {

	validErr := r.Validate()
	if validErr != nil {
		return validErr
	}
	const query = "UPDATE " + t_releases + " SET " +
		Rc_identifier + " = ?, " + Rc_version + " = ?," + Rc_status + " = ?," +
		Rc_released_on + " = ? WHERE " + Rc_id + " = ? AND " + Rc_project_id + " = ? LIMIT 1"
	_, err := db.Exec(query, r.Identifier, r.Version, NewReleaseStatus(r.Status), r.ReleasedOn, r.Id, r.ProjectID)
	return err
}

// Delete removes the Release and all associated pages from the database.
func (r *Release) Delete(db database.DB) error {
	const query = "DELETE FROM " + t_releases + " WHERE " + Rc_id + " = ?  AND " + Rc_project_id + " = ? LIMIT 1"
	_, err := db.Exec(query, r.Id, r.ProjectID)
	return err
}

func GenerateArchiveName(p Project, r Release) string {
	return fmt.Sprintf("%s - %s[%d][%s].zip", p.Shorthand, r.Identifier, r.Version, r.Scanlator)
}
