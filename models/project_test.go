package models

import (
	"../assert"
	"errors"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"strings"
	"testing"
	"time"
)

func TestFindProject(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)

	defer db.Close()

	const id uint32 = 5
	const query_select string = "SELECT (`[a-z_]+`, ){4}`[a-z_]+` FROM `[a-z_]+` WHERE `[a-z_]+` = ?"

	cols := []string{"name", "shorthand", "description", "status", "created_at"}
	rows := sqlmock.NewRows(cols)
	rows2 := sqlmock.NewRows(cols)
	tm := time.Now()
	rows2.AddRow("name", "shortname", "some desc", 2, tm)
	mock.ExpectQuery(query_select).WithArgs(id).WillReturnRows(rows)
	mock.ExpectQuery(query_select).WithArgs(id).WillReturnRows(rows2)
	_, err = FindProject(db, id)
	assert.Equal(t, ErrNoSuchProject, err)

	project, err := FindProject(db, id)

	assert.Equal(t, nil, err)
	assert.Equal(t, "name", project.Name)
	assert.Equal(t, id, project.Id)
	assert.Equal(t, "shortname", project.Shorthand)
	assert.Equal(t, "some desc", project.Description)
	assert.Equal(t, "active", project.Status)
	assert.Equal(t, tm, project.CreatedAt)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestListProjects(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	const query_select string = "SELECT (`[a-z_]+`, ){5}`[a-z_]+` FROM `[a-z_]+`"

	// error case
	expErr := errors.New("error")
	mock.ExpectQuery(query_select).WillReturnError(expErr)

	// no results case
	cols := []string{"id", "name", "shorthand", "description", "status", "created_at"}
	rows := sqlmock.NewRows(cols)
	mock.ExpectQuery(query_select).WillReturnRows(rows)

	// some results case
	rows2 := sqlmock.NewRows(cols)
	tm := time.Now()
	rows2.AddRow(1, "name", "shortname", "some desc", 2, tm)
	rows2.AddRow(3, "name2", "shortname2", "some desc", 3, tm)
	mock.ExpectQuery(query_select).WillReturnRows(rows2)

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

	project := projects[0]
	assert.Equal(t, nil, err)
	assert.Equal(t, "name", project.Name)
	assert.Equal(t, uint32(1), project.Id)
	assert.Equal(t, "shortname", project.Shorthand)
	assert.Equal(t, "some desc", project.Description)
	assert.Equal(t, "active", project.Status)
	assert.Equal(t, tm, project.CreatedAt)

	project = projects[1]
	assert.Equal(t, nil, err)
	assert.Equal(t, "name2", project.Name)
	assert.Equal(t, uint32(3), project.Id)
	assert.Equal(t, "shortname2", project.Shorthand)
	assert.Equal(t, "some desc", project.Description)
	assert.Equal(t, "stalled", project.Status)
	assert.Equal(t, tm, project.CreatedAt)

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

	const query string = "INSERT INTO `[a-z_]+`.*"
	p := Project{}
	p.Name = "the name"
	p.Shorthand = "short"

	// tests validation failed case
	p.Status = "invalid"
	err = p.Save(db)
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
	err = p.Save(db)
	assert.Equal(t, nil, err)
	assert.Equal(t, uint32(7), p.Id)

	// tests error case
	err = p.Save(db)
	assert.Equal(t, expErr, err)

	// tests result error case
	err = p.Save(db)
	assert.Equal(t, expErr2, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestUpdateProject(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	const query string = "UPDATE `[a-z_]+`.*"
	p := Project{}
	p.Name = "the name"
	p.Shorthand = "short"
	p.Id = 6

	// tests validation failed case
	p.Status = "invalid"
	err = p.Update(db)
	assert.Equal(t, ErrInvalidProjectStatus, err)

	// success case
	p.Status = "active"
	mock.ExpectExec(query).WithArgs(p.Name, p.Shorthand, p.Description, 2, p.Id).WillReturnResult(sqlmock.NewResult(7, 1))

	// error case
	expErr := errors.New("error")
	mock.ExpectExec(query).WithArgs(p.Name, p.Shorthand, p.Description, 2, p.Id).WillReturnError(expErr)

	// tests success case
	err = p.Update(db)
	assert.Equal(t, nil, err)

	// tests error case
	err = p.Update(db)
	assert.Equal(t, expErr, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestDeleteProject(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	const query string = "DELETE FROM `[a-z_]+` WHERE `[a-z_]+` = \\? LIMIT 1"
	expErr := errors.New("error")
	p := Project{}
	p.Id = 7
	mock.ExpectExec(query).WillReturnError(expErr).WithArgs(p.Id)
	mock.ExpectExec(query).WithArgs(p.Id).WillReturnResult(sqlmock.NewResult(7, 1))
	err = p.Delete(db)
	assert.Equal(t, expErr, err)

	err = p.Delete(db)
	assert.Equal(t, nil, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}
