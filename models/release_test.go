package models

import (
	"errors"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"ims-release/assert"
	"strings"
	"testing"
	"time"
)

func TestReleaseStatus(t *testing.T) {
	sReleased := NewReleaseStatus("released")
	sDraft := NewReleaseStatus("draft")
	sUnknown := NewReleaseStatus("somestring")
	assert.Equal(t, ReleaseStatus(1), sReleased)
	assert.Equal(t, ReleaseStatus(2), sDraft)
	assert.Equal(t, ReleaseStatus(0), sUnknown)

	assert.Equal(t, "released", sReleased.String())
	assert.Equal(t, "draft", sDraft.String())
	assert.Equal(t, "unknown", sUnknown.String())

	sUnknown = ReleaseStatus(5)
	assert.Equal(t, "unknown", sUnknown.String())
}

func TestNewRelease(t *testing.T) {
	tm := time.Now()
	p := Project{Id: 7}
	r := NewRelease(p, "identifier", 1, "status", tm)
	assert.Equal(t, p.Id, r.ProjectID)
	assert.Equal(t, "identifier", r.Identifier)
	assert.Equal(t, uint32(1), r.Version)
	assert.Equal(t, "status", r.Status)
	assert.Equal(t, tm, r.ReleasedOn)
}

func TestFindRelease(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)

	defer db.Close()

	p := Project{Id: 7}
	const id uint32 = 5
	const query_select string = "SELECT (`[a-z_]+`, ){3}`[a-z_]+` FROM `releases` WHERE `id` = \\? AND `project_id` = \\?"

	cols := []string{"identifier", "version", "status", "released_on"}
	rows := sqlmock.NewRows(cols)
	rows2 := sqlmock.NewRows(cols)
	tm := time.Now()
	r1 := Release{Id: id, Identifier: "identifier", Version: 1, Status: RStatusReleasedStr, ReleasedOn: tm, ProjectID: p.Id, Scanlator: "ims"}
	rows2.AddRow(r1.Identifier, r1.Version, NewReleaseStatus(r1.Status), r1.ReleasedOn)
	// case of no rows
	mock.ExpectQuery(query_select).WithArgs(id, p.Id).WillReturnRows(rows)

	// case of result found
	mock.ExpectQuery(query_select).WithArgs(id, p.Id).WillReturnRows(rows2)

	// case of db error
	expErr := errors.New("error")
	mock.ExpectQuery(query_select).WithArgs(id, p.Id).WillReturnError(expErr)

	_, err = FindRelease(db, p, id)
	assert.Equal(t, ErrNoSuchRelease, err)

	release, err := FindRelease(db, p, id)

	assert.Equal(t, nil, err)
	assert.Equal(t, r1, release)

	_, err = FindRelease(db, p, id)
	assert.Equal(t, expErr, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestListReleases(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	const query_select string = "SELECT (`[a-z_]+`, ){4}`[a-z_]+` FROM `releases`"
	p := Project{Id: 9}

	tm := time.Now()
	r1 := Release{Id: 5, Identifier: "identifier", Version: 1, Status: RStatusReleasedStr, ReleasedOn: tm, ProjectID: p.Id, Scanlator: "ims"}
	r2 := Release{Id: 9, Identifier: "identifier2", Version: 5, Status: RStatusDraftStr, ReleasedOn: tm, ProjectID: p.Id, Scanlator: "ims"}
	// error case
	expErr := errors.New("error")
	mock.ExpectQuery(query_select).WithArgs(p.Id).WillReturnError(expErr)

	// no results case
	cols := []string{"id", "identifier", "version", "status", "released_on"}
	rows := sqlmock.NewRows(cols)
	mock.ExpectQuery(query_select).WithArgs(p.Id).WillReturnRows(rows)

	// some results case
	rows2 := sqlmock.NewRows(cols)
	rows2.AddRow(r1.Id, r1.Identifier, r1.Version, NewReleaseStatus(r1.Status), r1.ReleasedOn)
	rows2.AddRow(r2.Id, r2.Identifier, r2.Version, NewReleaseStatus(r2.Status), r2.ReleasedOn)
	mock.ExpectQuery(query_select).WithArgs(p.Id).WillReturnRows(rows2)

	// some results with error case
	rows3 := sqlmock.NewRows(cols)
	rows3.AddRow(r1.Id, r1.Identifier, r1.Version, NewReleaseStatus(r1.Status), r1.ReleasedOn)
	rows3.AddRow(r2.Id, r2.Identifier, r2.Version, NewReleaseStatus(r2.Status), r2.ReleasedOn)
	expErr2 := errors.New("row error")
	rows3.RowError(1, expErr2)
	mock.ExpectQuery(query_select).WithArgs(p.Id).WillReturnRows(rows3)

	// some results with scan error case
	rows4 := sqlmock.NewRows(cols)
	rows4.AddRow(r1.Id, r1.Identifier, r1.Version, NewReleaseStatus(r1.Status), r1.ReleasedOn)
	rows4.AddRow(r2.Id, r2.Identifier, r2.Version, NewReleaseStatus(r2.Status), "malformed time")
	mock.ExpectQuery(query_select).WithArgs(p.Id).WillReturnRows(rows4)

	// tests the error case
	_, err = ListReleases(db, p)
	assert.Equal(t, expErr, err)

	// tests the no results case
	releases, err := ListReleases(db, p)
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, len(releases))

	// tests the some results case
	releases, err = ListReleases(db, p)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(releases))
	assert.Equal(t, r1, releases[0])
	assert.Equal(t, r2, releases[1])

	// tests some results with error case
	releases, err = ListReleases(db, p)
	assert.Equal(t, expErr2, err)
	assert.Equal(t, 1, len(releases))
	assert.Equal(t, r1, releases[0])

	// tests some results with scan error case
	releases, err = ListReleases(db, p)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, 1, len(releases))
	assert.Equal(t, r1, releases[0])

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestValidateRelease(t *testing.T) {
	r := Release{}
	r.Status = "bla"
	err := r.Validate()
	assert.Equal(t, ErrInvalidReleaseStatus, err)

	r.Status = "draft"
	err = r.Validate()
	assert.Equal(t, nil, err)

	r.Identifier = strings.Repeat("a", 11)
	err = r.Validate()
	assert.Equal(t, ErrFieldTooLong, err)

	r.Identifier = "a"
	err = r.Validate()
	assert.Equal(t, nil, err)
}

func TestSaveRelease(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	const query string = "INSERT INTO `releases`.*"
	r := Release{}
	r.Identifier = "Ch1"
	r.Version = 1
	r.ProjectID = 5

	// tests validation failed case
	r.Status = "invalid"
	r, err = SaveRelease(db, r)
	assert.Equal(t, ErrInvalidReleaseStatus, err)

	// success case
	r.Status = "draft"
	mock.ExpectExec(query).WithArgs(r.Identifier, r.Version, NewReleaseStatus(r.Status), r.ReleasedOn, r.ProjectID).WillReturnResult(sqlmock.NewResult(7, 1))

	// error case
	expErr := errors.New("error")
	mock.ExpectExec(query).WithArgs(r.Identifier, r.Version, NewReleaseStatus(r.Status), r.ReleasedOn, r.ProjectID).WillReturnError(expErr)

	// error result case
	expErr2 := errors.New("error2")
	mock.ExpectExec(query).WithArgs(r.Identifier, r.Version, NewReleaseStatus(r.Status), r.ReleasedOn, r.ProjectID).WillReturnResult(sqlmock.NewErrorResult(expErr2))

	// tests success case
	r, err = SaveRelease(db, r)
	assert.Equal(t, nil, err)
	assert.Equal(t, uint32(7), r.Id)

	// tests error case
	r, err = SaveRelease(db, r)
	assert.Equal(t, expErr, err)

	// tests result error case
	r, err = SaveRelease(db, r)
	assert.Equal(t, expErr2, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestUpdateRelease(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	const query string = "UPDATE `releases`.*WHERE `id` = \\? AND `project_id` = \\? LIMIT 1"
	r := Release{}
	r.Identifier = "Ch1"
	r.Version = 1
	r.ProjectID = 5
	r.Id = 6

	// tests validation failed case
	r.Status = "invalid"
	r, err = UpdateRelease(db, r)
	assert.Equal(t, ErrInvalidReleaseStatus, err)

	// success case
	r.Status = "released"
	mock.ExpectExec(query).WithArgs(r.Identifier, r.Version, NewReleaseStatus(r.Status), r.ReleasedOn, r.Id, r.ProjectID).WillReturnResult(sqlmock.NewResult(7, 1))

	// error case
	expErr := errors.New("error")
	mock.ExpectExec(query).WithArgs(r.Identifier, r.Version, NewReleaseStatus(r.Status), r.ReleasedOn, r.Id, r.ProjectID).WillReturnError(expErr)

	// tests success case
	r, err = UpdateRelease(db, r)
	assert.Equal(t, nil, err)

	// tests error case
	r, err = UpdateRelease(db, r)
	assert.Equal(t, expErr, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestDeleteRelease(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	const query string = "DELETE FROM `releases` WHERE `id` = \\? AND `project_id` = \\? LIMIT 1"
	expErr := errors.New("error")
	r := Release{}
	r.Id = 7
	r.ProjectID = 4
	mock.ExpectExec(query).WillReturnError(expErr).WithArgs(r.Id, r.ProjectID)
	mock.ExpectExec(query).WithArgs(r.Id, r.ProjectID).WillReturnResult(sqlmock.NewResult(7, 1))
	r, err = DeleteRelease(db, r)
	assert.Equal(t, expErr, err)

	r, err = DeleteRelease(db, r)
	assert.Equal(t, nil, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestGenerateArchiveName(t *testing.T) {
	r := Release{Id: 6, Version: 1, Scanlator: "scans", Identifier: "v1"}
	p := Project{Shorthand: "short"}

	name := GenerateArchiveName(p, r)
	assert.Equal(t, "short - v1[1][scans].zip", name)
}
