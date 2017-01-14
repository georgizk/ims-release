package models

import (
	"errors"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"ims-release/assert"
	"strings"
	"testing"
	"time"
)

func TestNewContributor(t *testing.T) {
	tm := time.Now()
	c := NewContributor("name", "bio", tm)
	assert.Equal(t, "name", c.Name)
	assert.Equal(t, "bio", c.Biography)
	assert.Equal(t, tm, c.CreatedAt)
	assert.Equal(t, uint32(0), c.Id)
}

func TestFindContributor(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)

	defer db.Close()

	const id uint32 = 5
	const query_select string = "SELECT (`[a-z_]+`, ){2}`[a-z_]+` FROM `contributors` WHERE `id` = \\?"

	cols := []string{"name", "biography", "created_at"}
	rows := sqlmock.NewRows(cols)
	rows2 := sqlmock.NewRows(cols)
	tm := time.Now()
	m1 := Contributor{Id: id, Name: "name", Biography: "bio", CreatedAt: tm}
	rows2.AddRow(m1.Name, m1.Biography, m1.CreatedAt)
	// case of no rows
	mock.ExpectQuery(query_select).WithArgs(id).WillReturnRows(rows)

	// case of result found
	mock.ExpectQuery(query_select).WithArgs(id).WillReturnRows(rows2)

	// case of db error
	expErr := errors.New("error")
	mock.ExpectQuery(query_select).WithArgs(id).WillReturnError(expErr)
	_, err = FindContributor(db, id)
	assert.Equal(t, ErrNoSuchContributor, err)

	contributor, err := FindContributor(db, id)

	assert.Equal(t, nil, err)
	assert.Equal(t, m1, contributor)

	_, err = FindContributor(db, id)
	assert.Equal(t, expErr, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestListContributors(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	const query_select string = "SELECT (`[a-z_]+`, ){3}`[a-z_]+` FROM `contributors`"

	tm := time.Now()
	m1 := Contributor{Id: 1, Name: "name", Biography: "bio", CreatedAt: tm}
	m2 := Contributor{Id: 7, Name: "name2", Biography: "bio2", CreatedAt: tm}
	// error case
	expErr := errors.New("error")
	mock.ExpectQuery(query_select).WillReturnError(expErr)

	// no results case
	cols := []string{"id", "name", "biography", "created_at"}
	rows := sqlmock.NewRows(cols)
	mock.ExpectQuery(query_select).WillReturnRows(rows)

	// some results case
	rows2 := sqlmock.NewRows(cols)

	rows2.AddRow(m1.Id, m1.Name, m1.Biography, m1.CreatedAt)
	rows2.AddRow(m2.Id, m2.Name, m2.Biography, m2.CreatedAt)
	mock.ExpectQuery(query_select).WillReturnRows(rows2)

	// some results with error case
	rows3 := sqlmock.NewRows(cols)
	rows3.AddRow(m1.Id, m1.Name, m1.Biography, m1.CreatedAt)
	rows3.AddRow(m2.Id, m2.Name, m2.Biography, m2.CreatedAt)
	expErr2 := errors.New("row error")
	rows3.RowError(1, expErr2)
	mock.ExpectQuery(query_select).WillReturnRows(rows3)

	// some results with scan error case
	rows4 := sqlmock.NewRows(cols)
	rows4.AddRow(m1.Id, m1.Name, m1.Biography, m1.CreatedAt)
	rows4.AddRow(m2.Id, m2.Name, m2.Biography, "malformed time")
	mock.ExpectQuery(query_select).WillReturnRows(rows4)

	// tests the error case
	_, err = ListContributors(db)
	assert.Equal(t, expErr, err)

	// tests the no results case
	contributors, err := ListContributors(db)
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, len(contributors))

	// tests the some results case
	contributors, err = ListContributors(db)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(contributors))
	assert.Equal(t, m1, contributors[0])
	assert.Equal(t, m2, contributors[1])

	// tests some results with error case
	contributors, err = ListContributors(db)
	assert.Equal(t, expErr2, err)
	assert.Equal(t, 1, len(contributors))
	assert.Equal(t, m1, contributors[0])

	// tests some results with scan error case
	contributors, err = ListContributors(db)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, 1, len(contributors))
	assert.Equal(t, m1, contributors[0])

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestValidateContributor(t *testing.T) {
	c := Contributor{}

	c.Name = strings.Repeat("a", 65536)
	err := c.Validate()
	assert.Equal(t, ErrFieldTooLong, err)

	c.Name = strings.Repeat("а", 32767) // cyrilic character should take 2 bytes
	err = c.Validate()
	assert.Equal(t, nil, err)

	c.Biography = strings.Repeat("а", 32768)
	err = c.Validate()
	assert.Equal(t, ErrFieldTooLong, err)

	c.Biography = strings.Repeat("b", 65535)
	err = c.Validate()
	assert.Equal(t, nil, err)
}

func TestSaveContributor(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	const query string = "INSERT INTO `contributors`.*"
	c := Contributor{}
	c.Name = "the name"

	// tests validation failed case
	c.Biography = strings.Repeat("a", 65536)
	c, err = SaveContributor(db, c)
	assert.Equal(t, ErrFieldTooLong, err)

	// success case
	c.Biography = "bio"
	mock.ExpectExec(query).WithArgs(c.Name, c.Biography, c.CreatedAt).WillReturnResult(sqlmock.NewResult(7, 1))

	// error case
	expErr := errors.New("error")
	mock.ExpectExec(query).WithArgs(c.Name, c.Biography, c.CreatedAt).WillReturnError(expErr)

	// error result case
	expErr2 := errors.New("error2")
	mock.ExpectExec(query).WithArgs(c.Name, c.Biography, c.CreatedAt).WillReturnResult(sqlmock.NewErrorResult(expErr2))

	// tests success case
	c, err = SaveContributor(db, c)
	assert.Equal(t, nil, err)
	assert.Equal(t, uint32(7), c.Id)

	// tests error case
	c, err = SaveContributor(db, c)
	assert.Equal(t, expErr, err)

	// tests result error case
	c, err = SaveContributor(db, c)
	assert.Equal(t, expErr2, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestUpdateContributor(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	const query string = "UPDATE `contributors`.*WHERE `id` = \\? LIMIT 1"
	c := Contributor{}
	c.Name = "the name"

	// tests validation failed case
	c.Biography = strings.Repeat("a", 65536)
	c, err = UpdateContributor(db, c)
	assert.Equal(t, ErrFieldTooLong, err)

	// success case
	c.Biography = "habs fan"
	mock.ExpectExec(query).WithArgs(c.Name, c.Biography, c.Id).WillReturnResult(sqlmock.NewResult(7, 1))

	// error case
	expErr := errors.New("error")
	mock.ExpectExec(query).WithArgs(c.Name, c.Biography, c.Id).WillReturnError(expErr)

	// tests success case
	c, err = UpdateContributor(db, c)
	assert.Equal(t, nil, err)

	// tests error case
	c, err = UpdateContributor(db, c)
	assert.Equal(t, expErr, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestDeleteContributor(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	const query string = "DELETE FROM `contributors` WHERE `id` = \\? LIMIT 1"
	expErr := errors.New("error")
	c := Contributor{}
	c.Id = 7
	mock.ExpectExec(query).WillReturnError(expErr).WithArgs(c.Id)
	mock.ExpectExec(query).WithArgs(c.Id).WillReturnResult(sqlmock.NewResult(7, 1))
	c, err = DeleteContributor(db, c)
	assert.Equal(t, expErr, err)

	c, err = DeleteContributor(db, c)
	assert.Equal(t, nil, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}
