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
		id int not null auto_increment,
		chapter varchar(255),
		version int,
		status varchar(255),
		checksum varchar(255),
		released_on timestamp,
		project_id int,
		foreign key(project_id) references projects(id),
		primary key(id)
);`

	QSaveRelease string = `insert into releases (
		chapter, version, status, checksum, released_on, project_id
) values (
		?, ?, ?, ?, ?, ?
);`

	QUpdateRelease string = `update releases set
chapter = ?, version = ?, status = ?, checksum = ?, released_on = ?
where id = ?;`

	QDeleteRelease string = `delete from releases where id = ?;`

	QListReleasesDesc string = `select
id, chapter, version, status, checksum, released_on
from releases
where project_id = ?
order by released_on desc;`

	QListReleasesAsc string = `select
id, chapter, version, status, checksum, released_on
from releases
where project_id = ?
order by released_on asc;`

	QFindRelease string = `select
chapter, version, status, checksum, released_on, project_id
from releases
where id = ?;`
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
	ReleasedOn time.Time     `json:"releasedOn"`
	ProjectID  int           `json:"projectId"`
}

// NewRelease constructs a brand new Release instance, with a default state lacking information its (future) position in
// a database.
func NewRelease(projectId, version int, chapterName string) Release {
	return Release{
		0,
		chapterName,
		version,
		RStatusDraft,
		"",
		time.Now(),
		projectId,
	}
}

// FindRelease attempts to lookup a release by ID.
func FindRelease(id int, db *sql.DB) (Release, error) {
	r := Release{}
	row := db.QueryRow(QFindRelease, id)
	if row == nil {
		return Release{}, ErrNoSuchRelease
	}
	err := row.Scan(&r.Chapter, &r.Version, &r.Status, &r.Checksum, &r.ReleasedOn, &r.ProjectID)
	if err != nil {
		return Release{}, err
	}
	r.Id = id
	return r, nil
}

// ListReleases attempts to obtain a list of all of the releases in the database.
func ListReleases(projectId int, ordering string, db *sql.DB) ([]Release, error) {
	releases := []Release{}
	query := QListReleasesDesc
	if ordering == "oldest" {
		query = QListReleasesAsc
	}
	rows, err := db.Query(query, projectId)
	if err != nil {
		return []Release{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, version, projectId int
		var chapter, status, checksum string
		var released time.Time
		scanErr := rows.Scan(&id, &chapter, &version, &status, &checksum, &released, &projectId)
		if scanErr != nil {
			err = scanErr
		}
		releases = append(releases, Release{
			id,
			chapter,
			version,
			ReleaseStatus(status),
			checksum,
			released,
			projectId,
		})
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
	validErr := r.Validate()
	if validErr != nil {
		return validErr
	}
	// TODO - Where should we compute checksums?
	_, err := db.Exec(QSaveRelease, r.Chapter, r.Version, string(r.Status), r.Checksum, r.ReleasedOn, r.ProjectID)
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
	_, err := db.Exec(QUpdateRelease, r.Chapter, r.Version, r.Status, r.Checksum, now, r.Id)
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
