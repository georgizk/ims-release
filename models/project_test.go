package models

import (
	"errors"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"ims-release/assert"
	"strings"
	"testing"
	"time"
)

func TestProjectStatus(t *testing.T) {
	sCompleted := NewProjectStatus("completed")
	sActive := NewProjectStatus("active")
	sStalled := NewProjectStatus("stalled")
	sDropped := NewProjectStatus("dropped")
	sUnknown := NewProjectStatus("somestring")
	assert.Equal(t, ProjectStatus(1), sCompleted)
	assert.Equal(t, ProjectStatus(2), sActive)
	assert.Equal(t, ProjectStatus(3), sStalled)
	assert.Equal(t, ProjectStatus(4), sDropped)
	assert.Equal(t, ProjectStatus(0), sUnknown)

	assert.Equal(t, "completed", sCompleted.String())
	assert.Equal(t, "active", sActive.String())
	assert.Equal(t, "stalled", sStalled.String())
	assert.Equal(t, "dropped", sDropped.String())
	assert.Equal(t, "unknown", sUnknown.String())

	sUnknown = ProjectStatus(5)
	assert.Equal(t, "unknown", sUnknown.String())
}

func TestNewProject(t *testing.T) {
	tm := time.Now()
	p := NewProject("name", "shorthand", "description", "status", tm)
	assert.Equal(t, "name", p.Name)
	assert.Equal(t, "shorthand", p.Shorthand)
	assert.Equal(t, "description", p.Description)
	assert.Equal(t, "status", p.Status)
	assert.Equal(t, tm, p.CreatedAt)
	assert.Equal(t, uint32(0), p.Id)
}

func TestFindProject(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)

	defer db.Close()

	const id uint32 = 5
	const query_select string = "SELECT (`[a-z_]+`, ){4}`[a-z_]+` FROM `projects` WHERE `id` = \\?"

	cols := []string{"name", "shorthand", "description", "status", "created_at"}
	rows := sqlmock.NewRows(cols)
	rows2 := sqlmock.NewRows(cols)
	tm := time.Now()
	p1 := Project{Id: id, Name: "name", Shorthand: "shortname", Description: "some desc", Status: PStatusCompletedStr, CreatedAt: tm}
	rows2.AddRow(p1.Name, p1.Shorthand, p1.Description, NewProjectStatus(p1.Status), p1.CreatedAt)
	// case of no rows
	mock.ExpectQuery(query_select).WithArgs(id).WillReturnRows(rows)

	// case of result found
	mock.ExpectQuery(query_select).WithArgs(id).WillReturnRows(rows2)

	// case of db error
	expErr := errors.New("error")
	mock.ExpectQuery(query_select).WithArgs(id).WillReturnError(expErr)
	_, err = FindProject(db, id)
	assert.Equal(t, ErrNoSuchProject, err)

	project, err := FindProject(db, id)

	assert.Equal(t, nil, err)
	assert.Equal(t, p1, project)

	_, err = FindProject(db, id)
	assert.Equal(t, expErr, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestListProjects(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	const query_select string = "SELECT (`[a-z_]+`, ){5}`[a-z_]+` FROM `projects`"

	tm := time.Now()
	p1 := Project{Id: 1, Name: "name", Shorthand: "shortname", Description: "some desc", Status: PStatusCompletedStr, CreatedAt: tm}
	p2 := Project{Id: 7, Name: "name2", Shorthand: "shortname2", Description: "some desc2", Status: PStatusDroppedStr, CreatedAt: tm}
	// error case
	expErr := errors.New("error")
	mock.ExpectQuery(query_select).WillReturnError(expErr)

	// no results case
	cols := []string{"id", "name", "shorthand", "description", "status", "created_at"}
	rows := sqlmock.NewRows(cols)
	mock.ExpectQuery(query_select).WillReturnRows(rows)

	// some results case
	rows2 := sqlmock.NewRows(cols)

	rows2.AddRow(p1.Id, p1.Name, p1.Shorthand, p1.Description, NewProjectStatus(p1.Status), p1.CreatedAt)
	rows2.AddRow(p2.Id, p2.Name, p2.Shorthand, p2.Description, NewProjectStatus(p2.Status), p2.CreatedAt)
	mock.ExpectQuery(query_select).WillReturnRows(rows2)

	// some results with error case
	rows3 := sqlmock.NewRows(cols)
	rows3.AddRow(p1.Id, p1.Name, p1.Shorthand, p1.Description, NewProjectStatus(p1.Status), p1.CreatedAt)
	rows3.AddRow(p2.Id, p2.Name, p2.Shorthand, p2.Description, NewProjectStatus(p2.Status), p2.CreatedAt)
	expErr2 := errors.New("row error")
	rows3.RowError(1, expErr2)
	mock.ExpectQuery(query_select).WillReturnRows(rows3)

	// some results with scan error case
	rows4 := sqlmock.NewRows(cols)
	rows4.AddRow(p1.Id, p1.Name, p1.Shorthand, p1.Description, NewProjectStatus(p1.Status), p1.CreatedAt)
	rows4.AddRow(p2.Id, p2.Name, p2.Shorthand, p2.Description, NewProjectStatus(p2.Status), "malformed time")
	mock.ExpectQuery(query_select).WillReturnRows(rows4)

	// tests the error case
	_, err = ListProjects(db)
	assert.Equal(t, expErr, err)

	// tests the no results case
	projects, err := ListProjects(db)
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, len(projects))

	// tests the some results case
	projects, err = ListProjects(db)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(projects))
	assert.Equal(t, p1, projects[0])
	assert.Equal(t, p2, projects[1])

	// tests some results with error case
	projects, err = ListProjects(db)
	assert.Equal(t, expErr2, err)
	assert.Equal(t, 1, len(projects))
	assert.Equal(t, p1, projects[0])

	// tests some results with scan error case
	projects, err = ListProjects(db)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, 1, len(projects))
	assert.Equal(t, p1, projects[0])

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestValidateProject(t *testing.T) {
	p := Project{}
	p.Status = "bla"
	err := p.Validate()
	assert.Equal(t, ErrInvalidProjectStatus, err)

	p.Status = "dropped"
	err = p.Validate()
	assert.Equal(t, nil, err)

	p.Shorthand = strings.Repeat("a", 50)
	err = p.Validate()
	assert.Equal(t, ErrFieldTooLong, err)

	p.Shorthand = "a"
	err = p.Validate()
	assert.Equal(t, nil, err)

	p.Description = strings.Repeat("♥", 21846) // one heart should take 3 bytes
	err = p.Validate()
	assert.Equal(t, ErrFieldTooLong, err)

	p.Description = ""
	p.Name = strings.Repeat("a", 65535) // this one should be fine
	err = p.Validate()
	assert.Equal(t, nil, err)

	p.Name = strings.Repeat("a", 65536)
	err = p.Validate()
	assert.Equal(t, ErrFieldTooLong, err)

	p.Name = strings.Repeat("а", 32767) // cyrilic character should take 2 bytes
	err = p.Validate()
	assert.Equal(t, nil, err)

	p.Name = strings.Repeat("а", 32768)
	err = p.Validate()
	assert.Equal(t, ErrFieldTooLong, err)
}

func TestSaveProject(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	const query string = "INSERT INTO `projects`.*"
	p := Project{}
	p.Name = "the name"
	p.Shorthand = "short"

	// tests validation failed case
	p.Status = "invalid"
	p, err = SaveProject(db, p)
	assert.Equal(t, ErrInvalidProjectStatus, err)

	// success case
	p.Status = "completed"
	mock.ExpectExec(query).WithArgs(p.Name, p.Shorthand, p.Description, 1, p.CreatedAt).WillReturnResult(sqlmock.NewResult(7, 1))

	// error case
	expErr := errors.New("error")
	mock.ExpectExec(query).WithArgs(p.Name, p.Shorthand, p.Description, 1, p.CreatedAt).WillReturnError(expErr)

	// error result case
	expErr2 := errors.New("error2")
	mock.ExpectExec(query).WithArgs(p.Name, p.Shorthand, p.Description, 1, p.CreatedAt).WillReturnResult(sqlmock.NewErrorResult(expErr2))

	// tests success case
	p, err = SaveProject(db, p)
	assert.Equal(t, nil, err)
	assert.Equal(t, uint32(7), p.Id)

	// tests error case
	p, err = SaveProject(db, p)
	assert.Equal(t, expErr, err)

	// tests result error case
	p, err = SaveProject(db, p)
	assert.Equal(t, expErr2, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestUpdateProject(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	const query string = "UPDATE `projects`.*WHERE `id` = \\? LIMIT 1"
	p := Project{}
	p.Name = "the name"
	p.Shorthand = "short"
	p.Id = 6

	// tests validation failed case
	p.Status = "invalid"
	p, err = UpdateProject(db, p)
	assert.Equal(t, ErrInvalidProjectStatus, err)

	// success case
	p.Status = "active"
	mock.ExpectExec(query).WithArgs(p.Name, p.Shorthand, p.Description, 2, p.Id).WillReturnResult(sqlmock.NewResult(7, 1))

	// error case
	expErr := errors.New("error")
	mock.ExpectExec(query).WithArgs(p.Name, p.Shorthand, p.Description, 2, p.Id).WillReturnError(expErr)

	// tests success case
	p, err = UpdateProject(db, p)
	assert.Equal(t, nil, err)

	// tests error case
	p, err = UpdateProject(db, p)
	assert.Equal(t, expErr, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestDeleteProject(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	const query string = "DELETE FROM `projects` WHERE `id` = \\? LIMIT 1"
	expErr := errors.New("error")
	p := Project{}
	p.Id = 7
	mock.ExpectExec(query).WillReturnError(expErr).WithArgs(p.Id)
	mock.ExpectExec(query).WithArgs(p.Id).WillReturnResult(sqlmock.NewResult(7, 1))
	p, err = DeleteProject(db, p)
	assert.Equal(t, expErr, err)

	p, err = DeleteProject(db, p)
	assert.Equal(t, nil, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}
