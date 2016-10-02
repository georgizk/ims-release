package models

import (
	"database/sql"
	"errors"
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
	ErrInvalidReleaseStatus = errors.New("Invalid release status.")
	ErrNoSuchRelease        = errors.New("Could not find release.")
)

// Database queries for operations on Releases.
const (
	QInitTableReleases string = `create table if not exists releases (
		id int not null primary key,
		chapter varchar(255),
		version int,
		status varchar(255),
		checksum varchar(255),
		released_on timestamp,
		project_id int,
		foreign key(project_id) references projects(id)
);`

	QSaveRelease string = `insert into releases (
		chapter, version, status, checksum, released_on
) values (
		$1, $2, $3, $4, $5
);`

	QUpdateRelease string = `update releases set
chapter = $2, version = $3, status = $4, checksum = $5, released_on = $6
where id = $1;`

	QDeleteRelease string = `delete from releases where id = $1;`

	QListReleases string = `select (
		id, chapter, version, status, checksum, released_on
) from releases
where project_id = $1;`

	QFindRelease string = `select (
		chapter, version, status, checksum, released_on
) from releases
where id = $1;`
)

// Release contains information about a release, which there are many of under a given Project.  It contains information
// about which chapter of manga the release was created for, which version of the release of said chapter it is for, and
// the status of the release of the chapter itself, which may not be final right away.
type Release struct {
	Id         int           `json:"id"`
	Chapter    string        `json:"chapter"`
	Version    int           `json:"version"`
	Status     ReleaseStatus `json:"status"`
	Checksum   string        `json:"checksum"`
	ReleasedOn time.Time     `json:"releasedOn:`
}

// NewRelease constructs a brand new Release instance, with a default state lacking information its (future) position in
// a database.
func NewRelease(version int, chapterName string) Release {
	return Release{
		0,
		chapterName,
		version,
		RStatusDraft,
		"",
		time.Now(),
	}
}

// FindRelease attempts to lookup a release by ID.
func FindRelease(id int, db *sql.DB) (Release, error) {
	r := Release{}
	row := db.QueryRow(QFindRelease, id)
	if row == nil {
		return Release{}, ErrNoSuchRelease
	}
	err := row.Scan(&r.Chapter, &r.Version, &r.Status, &r.Checksum, &r.ReleasedOn)
	if err != nil {
		return Release{}, err
	}
	r.Id = id
	return r, nil
}

// ListReleases attempts to obtain a list of all of the releases in the database.
func ListReleases(projectId int, db *sql.DB) ([]Release, error) {
	releases := []Release{}
	rows, err := db.Query(QListReleases, projectId)
	if err != nil {
		return []Release{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, version int
		var chapter, status, checksum string
		var released time.Time
		scanErr := rows.Scan(&id, &chapter, &version, &status, &checksum, &released)
		if scanErr != nil {
			err = scanErr
		}
		releases = append(releases, Release{id, chapter, version, ReleaseStatus(status), checksum, released})
	}
	return releases, err
}

// Validate checks that the "status" of the project is one of the accepted ReleaseStatus values.
func (r *Release) Validate() error {
	for _, status := range RStatuses {
		if r.Status == status {
			return nil
		}
	}
	return ErrInvalidReleaseStatus
}

// Save inserts the release into the database and updates its Id field.
func (r *Release) Save(db *sql.DB) error {
	// TODO - Where should we compute checksums?
	_, err := db.Exec(QSaveRelease, r.Chapter, r.Version, r.Status, r.Checksum, r.ReleasedOn)
	if err != nil {
		return err
	}
	row := db.QueryRow(QLastInsertID)
	if row == nil {
		return ErrCouldNotGetID
	}
	return row.Scan(&r.Id)
}

// Update modifies all of the fields of a Release in place with whatever is currently in the struct.
func (r *Release) Update(db *sql.DB) error {
	now := time.Now()
	_, err := db.Exec(QUpdateRelease, r.Id, r.Chapter, r.Version, r.Status, r.Checksum, now)
	r.ReleasedOn = now
	return err
}

// Delete removes the Release and all associated pages from the database.
func (r *Release) Delete(db *sql.DB) error {
	pages, listErr := ListPages(r.Id, db)
	var deleteErr error
	for _, page := range pages {
		dErr := page.Delete(db)
		if dErr != nil {
			deleteErr = dErr
		}
	}
	_, err := db.Exec(QDeleteRelease, r.Id)
	if err != nil {
		return err
	}
	if listErr != nil {
		return listErr
	}
	return deleteErr
}
