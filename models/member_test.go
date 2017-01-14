package models

import (
	"errors"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"ims-release/assert"
	"strings"
	"testing"
	"time"
)

func TestNewMember(t *testing.T) {
	tm := time.Now()
	m := NewMember("name", "bio", tm)
	assert.Equal(t, "name", m.Name)
	assert.Equal(t, "bio", m.Biography)
	assert.Equal(t, tm, m.CreatedAt)
	assert.Equal(t, uint32(0), m.Id)
}

func TestFindMember(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)

	defer db.Close()

	const id uint32 = 5
	const query_select string = "SELECT (`[a-z_]+`, ){2}`[a-z_]+` FROM `members` WHERE `id` = \\?"

	cols := []string{"name", "biography", "created_at"}
	rows := sqlmock.NewRows(cols)
	rows2 := sqlmock.NewRows(cols)
	tm := time.Now()
	m1 := Member{Id: id, Name: "name", Biography: "bio", CreatedAt: tm}
	rows2.AddRow(m1.Name, m1.Biography, m1.CreatedAt)
	// case of no rows
	mock.ExpectQuery(query_select).WithArgs(id).WillReturnRows(rows)

	// case of result found
	mock.ExpectQuery(query_select).WithArgs(id).WillReturnRows(rows2)

	// case of db error
	expErr := errors.New("error")
	mock.ExpectQuery(query_select).WithArgs(id).WillReturnError(expErr)
	_, err = FindMember(db, id)
	assert.Equal(t, ErrNoSuchMember, err)

	member, err := FindMember(db, id)

	assert.Equal(t, nil, err)
	assert.Equal(t, m1, member)

	_, err = FindMember(db, id)
	assert.Equal(t, expErr, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestListMembers(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	const query_select string = "SELECT (`[a-z_]+`, ){3}`[a-z_]+` FROM `members`"

	tm := time.Now()
	m1 := Member{Id: 1, Name: "name", Biography: "bio", CreatedAt: tm}
	m2 := Member{Id: 7, Name: "name2", Biography: "bio2", CreatedAt: tm}
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
	_, err = ListMembers(db)
	assert.Equal(t, expErr, err)

	// tests the no results case
	members, err := ListMembers(db)
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, len(members))

	// tests the some results case
	members, err = ListMembers(db)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(members))
	assert.Equal(t, m1, members[0])
	assert.Equal(t, m2, members[1])

	// tests some results with error case
	members, err = ListMembers(db)
	assert.Equal(t, expErr2, err)
	assert.Equal(t, 1, len(members))
	assert.Equal(t, m1, members[0])

	// tests some results with scan error case
	members, err = ListMembers(db)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, 1, len(members))
	assert.Equal(t, m1, members[0])

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestValidateMember(t *testing.T) {
	m := Member{}

	m.Name = strings.Repeat("a", 65536)
	err := m.Validate()
	assert.Equal(t, ErrFieldTooLong, err)

	m.Name = strings.Repeat("а", 32767) // cyrilic character should take 2 bytes
	err = m.Validate()
	assert.Equal(t, nil, err)

	m.Biography = strings.Repeat("а", 32768)
	err = m.Validate()
	assert.Equal(t, ErrFieldTooLong, err)

	m.Biography = strings.Repeat("b", 65535)
	err = m.Validate()
	assert.Equal(t, nil, err)
}

func TestSaveMember(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	const query string = "INSERT INTO `members`.*"
	m := Member{}
	m.Name = "the name"

	// tests validation failed case
	m.Biography = strings.Repeat("a", 65536)
	m, err = SaveMember(db, m)
	assert.Equal(t, ErrFieldTooLong, err)

	// success case
	m.Biography = "bio"
	mock.ExpectExec(query).WithArgs(m.Name, m.Biography, m.CreatedAt).WillReturnResult(sqlmock.NewResult(7, 1))

	// error case
	expErr := errors.New("error")
	mock.ExpectExec(query).WithArgs(m.Name, m.Biography, m.CreatedAt).WillReturnError(expErr)

	// error result case
	expErr2 := errors.New("error2")
	mock.ExpectExec(query).WithArgs(m.Name, m.Biography, m.CreatedAt).WillReturnResult(sqlmock.NewErrorResult(expErr2))

	// tests success case
	m, err = SaveMember(db, m)
	assert.Equal(t, nil, err)
	assert.Equal(t, uint32(7), m.Id)

	// tests error case
	m, err = SaveMember(db, m)
	assert.Equal(t, expErr, err)

	// tests result error case
	m, err = SaveMember(db, m)
	assert.Equal(t, expErr2, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestUpdateMember(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	const query string = "UPDATE `members`.*WHERE `id` = \\? LIMIT 1"
	m := Member{}
	m.Name = "the name"

	// tests validation failed case
	m.Biography = strings.Repeat("a", 65536)
	m, err = UpdateMember(db, m)
	assert.Equal(t, ErrFieldTooLong, err)

	// success case
	m.Biography = "habs fan"
	mock.ExpectExec(query).WithArgs(m.Name, m.Biography, m.Id).WillReturnResult(sqlmock.NewResult(7, 1))

	// error case
	expErr := errors.New("error")
	mock.ExpectExec(query).WithArgs(m.Name, m.Biography, m.Id).WillReturnError(expErr)

	// tests success case
	m, err = UpdateMember(db, m)
	assert.Equal(t, nil, err)

	// tests error case
	m, err = UpdateMember(db, m)
	assert.Equal(t, expErr, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestDeleteMember(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	const query string = "DELETE FROM `members` WHERE `id` = \\? LIMIT 1"
	expErr := errors.New("error")
	m := Member{}
	m.Id = 7
	mock.ExpectExec(query).WillReturnError(expErr).WithArgs(m.Id)
	mock.ExpectExec(query).WithArgs(m.Id).WillReturnResult(sqlmock.NewResult(7, 1))
	m, err = DeleteMember(db, m)
	assert.Equal(t, expErr, err)

	m, err = DeleteMember(db, m)
	assert.Equal(t, nil, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}
