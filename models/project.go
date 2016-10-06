package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ProjectStatus is a type alias which will be used to create an enum of acceptable project status states.
type ProjectStatus string

// ProjectStatus pseudo-enum values
const (
	PStatusPublished ProjectStatus = "published"
	PStatusOngoing   ProjectStatus = "ongoing"
)

var (
	PStatuses = []ProjectStatus{PStatusPublished}
)

// Database queries for operations on Projects.
const (
	QInitTableProjects string = `create table if not exists projects (
		id int not null auto_increment,
		name varchar(255),
		project_name varchar(255) unique,
		description text,
		status varchar(255),
		created_at timestamp,
		primary key(id)
);`

	QSaveProject string = `insert into projects (
		name, project_name, description, status, created_at
) values (
		?, ?, ?, ?, ?
);`

	QUpdateProject string = `update projects set
name = ?, project_name = ?, description = ?, status = ?
where id = ?;`

	QDeleteProject string = `delete from projects where id = ?;`

	QListProjectsDesc string = `select id, name, project_name, description, status, created_at
from projects
order by created_at desc;`

	QListProjectsAsc string = `select id, name, project_name, description, status, created_at
from projects
order by created_at asc;`

	QFindProject string = `select
name, project_name, description, status, created_at
from projects
where id = ?;`
)

// Errors pertaining to the data in a Project or operations on Projects.
var (
	ErrInvalidProjectStatus = fmt.Errorf("Invalid project status.")
	ErrNoSuchProject        = errors.New("Could not find project.")
)

// Project contains information about a scanlation project, which has a human-readable name, a unique shorthand name,
// and a publishing status amongst other things.
type Project struct {
	Id          int           `json:"id"`
	Name        string        `json:"name"`
	Shorthand   string        `json:"projectName"`
	Description string        `json:"description"`
	Status      ProjectStatus `json:"status"`
	CreatedAt   time.Time     `json:"createdAt"`
}

// NewProject constructs a brand new Project instance, with a default state lacking information about its (future)
// position in a database.
func NewProject(name, shorthand, description string) Project {
	return Project{
		0,
		name,
		shorthand,
		description,
		PStatusOngoing,
		time.Now(),
	}
}

// FindProject attempts to lookup a project by ID.
func FindProject(id int, db *sql.DB) (Project, error) {
	p := Project{}
	row := db.QueryRow(QFindProject, id)
	if row == nil {
		return Project{}, ErrNoSuchProject
	}
	err := row.Scan(&p.Name, &p.Shorthand, &p.Description, &p.Status, &p.CreatedAt)
	if err != nil {
		return Project{}, err
	}
	p.Id = id
	return p, nil
}

// ListProjects attempts to obtain a list of all of the projects in the database.
func ListProjects(ordering string, db *sql.DB) ([]Project, error) {
	projects := []Project{}
	query := QListProjectsDesc
	if ordering == "oldest" {
		query = QListProjectsAsc
	}
	rows, err := db.Query(query)
	if err != nil {
		return []Project{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name, shorthand, description, status string
		var created time.Time
		err = rows.Scan(&id, &name, &shorthand, &description, &status, &created)
		projects = append(projects, Project{id, name, shorthand, description, ProjectStatus(status), created})
	}
	return projects, err
}

// Validate checks that the "status" of the project is one of the accepted ProjectStatus values.
func (p *Project) Validate() error {
	for _, status := range PStatuses {
		if p.Status == status {
			return nil
		}
	}
	return ErrInvalidProjectStatus
}

// Save inserts the project into the database and updates its Id field.
func (p *Project) Save(db *sql.DB) error {
	validErr := p.Validate()
	if validErr != nil {
		return validErr
	}
	_, err := db.Exec(QSaveProject, p.Name, p.Shorthand, p.Description, string(p.Status), p.CreatedAt)
	if err != nil {
		return err
	}
	row := db.QueryRow(QLastInsertID)
	if row == nil {
		return ErrCouldNotGetID
	}
	return row.Scan(&p.Id)
}

// Update modifies all of the fields of a Project in place with whatever is currently in the struct.
func (p *Project) Update(db *sql.DB) error {
	_, err := db.Exec(QUpdateProject, p.Id, p.Name, p.Shorthand, p.Description, p.Status)
	return err
}

// Delete removes the Project and all associated releases from the database.
func (p *Project) Delete(db *sql.DB) error {
	releases, listErr := ListReleases(p.Id, "newest", db)
	var deleteErr error
	for _, release := range releases {
		dErr := release.Delete(db)
		if dErr != nil {
			deleteErr = dErr
		}
	}
	_, err := db.Exec(QDeleteProject, p.Id)
	if err != nil {
		return err
	}
	if listErr != nil {
		return listErr
	}
	return deleteErr
}
