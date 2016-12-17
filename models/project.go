package models

import (
	"database/sql"
	"errors"
	"time"
)

// Project contains information about a scanlation project, which has a human-readable name, a unique shorthand name,
// and a publishing status amongst other things.
type Project struct {
	Id          uint32    `json:"id"`
	Name        string    `json:"name"`
	Shorthand   string    `json:"shorthand"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

// ProjectStatus is a type alias which will be used to create an enum of acceptable project status states.
type ProjectStatus uint32

// ProjectStatus pseudo-enum values
const (
	PStatusUnknown      ProjectStatus = 0
	PStatusUnknownStr   string        = "unknown"
	PStatusPublished    ProjectStatus = 1
	PStatusPublishedStr string        = "completed"
	PStatusOngoing      ProjectStatus = 2
	PStatusOngoingStr   string        = "active"
	PStatusStalled      ProjectStatus = 3
	PStatusStalledStr   string        = "stalled"
	PStatusDropped      ProjectStatus = 4
	PStatusDroppedStr   string        = "dropped"
)

func (s ProjectStatus) String() string {
	switch s {
	case PStatusPublished:
		return PStatusPublishedStr
	case PStatusOngoing:
		return PStatusOngoingStr
	case PStatusDropped:
		return PStatusDroppedStr
	case PStatusStalled:
		return PStatusStalledStr
	default:
		return PStatusUnknownStr
	}
}

func NewProjectStatus(val string) ProjectStatus {
	switch val {
	case PStatusPublishedStr:
		return PStatusPublished
	case PStatusOngoingStr:
		return PStatusOngoing
	case PStatusStalledStr:
		return PStatusStalled
	case PStatusDroppedStr:
		return PStatusDropped
	default:
		return PStatusUnknown
	}
}

// Database constants for projects
const (
	t_projects     string = "`projects`"
	Pc_id          string = "`id`"
	Pc_name        string = "`name`"
	Pc_shorthand   string = "`shorthand`"
	Pc_description string = "`description`"
	Pc_status      string = "`status`"
	Pc_created_at  string = "`created_at`"

	Pmax_len_shorthand   = 30
	Pmax_len_name        = 65535
	Pmax_len_description = 65535
)

// Errors pertaining to the data in a Project or operations on Projects.
var (
	ErrInvalidProjectStatus = errors.New("Invalid project status.")
	ErrNoSuchProject        = errors.New("Could not find project.")
)

// NewProject constructs a brand new Project instance, with a default state lacking information about its (future)
// position in a database.
func NewProject(name, shorthand, description string) Project {
	return Project{
		0,
		name,
		shorthand,
		description,
		PStatusOngoingStr,
		time.Now(),
	}
}

// FindProject attempts to lookup a project by ID.
func FindProject(db *sql.DB, id uint32) (Project, error) {
	p := Project{}
	var s ProjectStatus
	const query = "SELECT " + Pc_name + ", " + Pc_shorthand + ", " +
		Pc_description + ", " + Pc_status + ", " + Pc_created_at + " " +
		"FROM " + t_projects + " WHERE " + Pc_id + " = ?"

	row := db.QueryRow(query, id)
	err := row.Scan(&p.Name, &p.Shorthand, &p.Description, &s, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return Project{}, ErrNoSuchProject
	} else if err != nil {
		return Project{}, err
	}
	p.Status = s.String()
	p.Id = id
	return p, nil
}

// ListProjects attempts to obtain a list of all of the projects in the database.
func ListProjects(db *sql.DB) ([]Project, error) {
	projects := []Project{}

	const query = "SELECT " + Pc_id + ", " + Pc_name + ", " +
		Pc_shorthand + ", " + Pc_description + ", " + Pc_status + ", " +
		Pc_created_at + " FROM " + t_projects

	rows, err := db.Query(query)
	if err != nil {
		return []Project{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var p Project
		var s ProjectStatus
		err = rows.Scan(&p.Id, &p.Name, &p.Shorthand, &p.Description, &s, &p.CreatedAt)
		p.Status = s.String()
		projects = append(projects, p)
	}
	return projects, err
}

// Validate checks that the "status" of the project is one of the accepted ProjectStatus values.
func (p *Project) Validate() error {
	if PStatusUnknown == NewProjectStatus(p.Status) {
		return ErrInvalidProjectStatus
	}

	if len(p.Shorthand) > Pmax_len_shorthand || len(p.Name) > Pmax_len_name || len(p.Description) > Pmax_len_description {
		return ErrFieldTooLong
	}

	return nil
}

// Save inserts the project into the database and updates its Id field.
func (p *Project) Save(db *sql.DB) error {
	validErr := p.Validate()
	if validErr != nil {
		return validErr
	}

	const query = "INSERT INTO " + t_projects + " (" +
		Pc_name + ", " + Pc_shorthand + ", " + Pc_description + ", " +
		Pc_status + ", " + Pc_created_at + ") VALUES (?, ?, ?, ?, ?)"

	res, err := db.Exec(query, p.Name, p.Shorthand, p.Description, NewProjectStatus(p.Status), p.CreatedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	p.Id = uint32(id)
	return nil
}

// Update modifies all of the fields of a Project in place with whatever is currently in the struct.
func (p *Project) Update(db *sql.DB) error {
	validErr := p.Validate()
	if validErr != nil {
		return validErr
	}

	const query = "UPDATE " + t_projects + " SET " +
		Pc_name + " = ?, " + Pc_shorthand + " = ?, " + Pc_description + " = ?," +
		Pc_status + " = ? WHERE " + Pc_id + " = ? LIMIT 1"

	_, err := db.Exec(query, p.Name, p.Shorthand, p.Description, NewProjectStatus(p.Status), p.Id)
	return err
}

// Delete removes the Project and all associated releases from the database.
func (p *Project) Delete(db *sql.DB) error {
	const query = "DELETE FROM " + t_projects + " WHERE " +
		Pc_id + " = ? LIMIT 1"
	_, err := db.Exec(query, p.Id)
	return err
}
