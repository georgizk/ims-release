package models

import (
	"errors"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"ims-release/assert"
	"strings"
	"testing"
	"time"
)

func TestMimeType(t *testing.T) {
	tUnknown := NewMimeType("bla")
	tPng := NewMimeType("image/png")
	tJpg := NewMimeType("image/jpeg")

	assert.Equal(t, "image/png", tPng.String())
	assert.Equal(t, "image/jpeg", tJpg.String())
	assert.Equal(t, "application/octet-stream", tUnknown.String())

	assert.Equal(t, MimeType(1), tPng)
	assert.Equal(t, MimeType(2), tJpg)
	assert.Equal(t, MimeType(0), tUnknown)

	assert.Equal(t, tPng, MimeTypeFromFilename("bla.png"))
	assert.Equal(t, tJpg, MimeTypeFromFilename("bla.jpg"))
	assert.Equal(t, tUnknown, MimeTypeFromFilename("bla.bmp"))

	tUnknown = MimeType(6)
	assert.Equal(t, "application/octet-stream", tUnknown.String())
}

func TestNewPage(t *testing.T) {
	r := Release{Id: 5}
	tm := time.Now()
	p := NewPage(r, "file.png", tm)

	assert.Equal(t, uint32(0), p.Id)
	assert.Equal(t, r.Id, p.ReleaseID)
	assert.Equal(t, "file.png", p.Name)
	assert.Equal(t, MimeTypePng, p.MimeType)
	assert.Equal(t, tm, p.CreatedAt)
}

func TestFindPage(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)

	defer db.Close()

	r := Release{Id: 5}
	const id uint32 = 7
	const name = "pg001.png"
	const query = "SELECT (`[a-z_]+`, ){1}`[a-z_]+` FROM `pages` WHERE `id` = \\? AND `release_id` = \\?"
	cols := []string{"name", "created_at"}

	rows := sqlmock.NewRows(cols)
	rows2 := sqlmock.NewRows(cols)
	tm := time.Now()
	rows2.AddRow(name, tm)

	// case of no rows
	mock.ExpectQuery(query).WithArgs(id, r.Id).WillReturnRows(rows)

	// case of result found
	mock.ExpectQuery(query).WithArgs(id, r.Id).WillReturnRows(rows2)

	// case of db error
	expErr := errors.New("error")
	mock.ExpectQuery(query).WithArgs(id, r.Id).WillReturnError(expErr)

	// test no rows
	_, err = FindPage(db, r, id)
	assert.Equal(t, ErrNoSuchPage, err)

	page, err := FindPage(db, r, id)

	assert.Equal(t, nil, err)
	assert.Equal(t, name, page.Name)
	assert.Equal(t, MimeTypePng, page.MimeType)
	assert.Equal(t, r.Id, page.ReleaseID)
	assert.Equal(t, id, page.Id)
	assert.Equal(t, tm, page.CreatedAt)

	_, err = FindPage(db, r, id)
	assert.Equal(t, expErr, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestFindPageByName(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)

	defer db.Close()

	r := Release{Id: 5}
	const id uint32 = 7
	const name = "pg001.png"
	const query = "SELECT (`[a-z_]+`, ){1}`[a-z_]+` FROM `pages` WHERE `release_id` = \\? AND `name` = \\?"
	cols := []string{"id", "created_at"}

	rows := sqlmock.NewRows(cols)
	rows2 := sqlmock.NewRows(cols)
	tm := time.Now()
	rows2.AddRow(id, tm)

	// case of no rows
	mock.ExpectQuery(query).WithArgs(r.Id, name).WillReturnRows(rows)

	// case of result found
	mock.ExpectQuery(query).WithArgs(r.Id, name).WillReturnRows(rows2)

	// case of db error
	expErr := errors.New("error")
	mock.ExpectQuery(query).WithArgs(r.Id, name).WillReturnError(expErr)

	// test no rows
	_, err = FindPageByName(db, r, name)
	assert.Equal(t, ErrNoSuchPage, err)

	page, err := FindPageByName(db, r, name)

	assert.Equal(t, nil, err)
	assert.Equal(t, name, page.Name)
	assert.Equal(t, MimeTypePng, page.MimeType)
	assert.Equal(t, r.Id, page.ReleaseID)
	assert.Equal(t, id, page.Id)
	assert.Equal(t, tm, page.CreatedAt)

	_, err = FindPageByName(db, r, name)
	assert.Equal(t, expErr, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestGeneratePagePath(t *testing.T) {
	p := Project{Id: 5}
	r := Release{Id: 3}
	path := GeneratePagePath(p, r, "img001.jpg")
	assert.Equal(t, "5/3/img001.jpg", path)
}

func TestListPages(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	const query string = "SELECT (`[a-z_]+`, ){2}`[a-z_]+` FROM `pages`"
	r := Release{Id: 9}

	tm := time.Now()
	pg1 := Page{Id: 1, Name: "somepage.jpg", CreatedAt: tm, MimeType: MimeTypeJpg, ReleaseID: r.Id}
	pg2 := Page{Id: 3, Name: "somepage.png", CreatedAt: tm, MimeType: MimeTypePng, ReleaseID: r.Id}

	// error case
	expErr := errors.New("error")
	mock.ExpectQuery(query).WithArgs(r.Id).WillReturnError(expErr)

	// no results case
	cols := []string{"id", "name", "created_at"}
	rows := sqlmock.NewRows(cols)
	mock.ExpectQuery(query).WithArgs(r.Id).WillReturnRows(rows)

	// some results case
	rows2 := sqlmock.NewRows(cols)
	rows2.AddRow(pg1.Id, pg1.Name, pg1.CreatedAt)
	rows2.AddRow(pg2.Id, pg2.Name, pg2.CreatedAt)
	mock.ExpectQuery(query).WithArgs(r.Id).WillReturnRows(rows2)

	// some results with error case
	rows3 := sqlmock.NewRows(cols)
	rows3.AddRow(pg1.Id, pg1.Name, pg1.CreatedAt)
	rows3.AddRow(pg2.Id, pg2.Name, pg2.CreatedAt)
	expErr2 := errors.New("row error")
	rows3.RowError(1, expErr2)
	mock.ExpectQuery(query).WithArgs(r.Id).WillReturnRows(rows3)

	// some results with scan error case
	rows4 := sqlmock.NewRows(cols)
	rows4.AddRow(pg1.Id, pg1.Name, pg1.CreatedAt)
	rows4.AddRow(pg2.Id, pg2.Name, "malformed time")
	mock.ExpectQuery(query).WithArgs(r.Id).WillReturnRows(rows4)

	// tests the error case
	_, err = ListPages(db, r)
	assert.Equal(t, expErr, err)

	// tests the no results case
	pages, err := ListPages(db, r)
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, len(pages))

	// tests the some results case
	pages, err = ListPages(db, r)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(pages))
	assert.Equal(t, pg1, pages[0])
	assert.Equal(t, pg2, pages[1])

	// tests some results with error case
	pages, err = ListPages(db, r)
	assert.Equal(t, expErr2, err)
	assert.Equal(t, 1, len(pages))
	assert.Equal(t, pg1, pages[0])

	// tests some results with scan error case
	pages, err = ListPages(db, r)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, 1, len(pages))
	assert.Equal(t, pg1, pages[0])

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestValidatePage(t *testing.T) {
	p := Page{}
	err := p.Validate()
	assert.Equal(t, ErrPageNameEmpty, err)

	p.Name = "test"
	err = p.Validate()
	assert.Equal(t, ErrPageUnsupportedMimeType, err)

	p.Name = "test.png"
	err = p.Validate()
	assert.Equal(t, nil, err)

	p.Name = strings.Repeat("a", 256)
	err = p.Validate()
	assert.Equal(t, ErrPageNameTooLong, err)

	p.Name = strings.Repeat("a", 251) + ".jpg"
	err = p.Validate()
	assert.Equal(t, nil, err)
}

func TestSavePage(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	const query = "INSERT INTO `pages`.*"
	const id = 7

	// success case
	pSuccess := NewPage(Release{Id: 5}, "img.png", time.Now())
	mock.ExpectExec(query).WithArgs(pSuccess.Name, pSuccess.CreatedAt, pSuccess.ReleaseID).WillReturnResult(sqlmock.NewResult(id, 1))

	// error case
	expErr := errors.New("error")
	mock.ExpectExec(query).WithArgs(pSuccess.Name, pSuccess.CreatedAt, pSuccess.ReleaseID).WillReturnError(expErr)

	// error result case
	expErr2 := errors.New("error2")
	mock.ExpectExec(query).WithArgs(pSuccess.Name, pSuccess.CreatedAt, pSuccess.ReleaseID).WillReturnResult(sqlmock.NewErrorResult(expErr2))

	// tests success case
	err = pSuccess.Save(db)
	assert.Equal(t, nil, err)
	assert.Equal(t, uint32(id), pSuccess.Id)

	// tests error case
	err = pSuccess.Save(db)
	assert.Equal(t, expErr, err)

	// tests result error case
	err = pSuccess.Save(db)
	assert.Equal(t, expErr2, err)

	// tests validation failed case
	pErr := Page{Name: "bla"}
	err = pErr.Save(db)
	assert.NotEqual(t, nil, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestUpdatePage(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	p := Page{}
	err = p.Update(db)
	assert.Equal(t, ErrOperationNotSupported, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}

func TestDeletePage(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	p := Page{Id: 1, ReleaseID: 2}
	expErr := errors.New("error")
	const query string = "DELETE FROM `pages` WHERE `id` = \\? AND `release_id` = \\? LIMIT 1"
	mock.ExpectExec(query).WillReturnError(expErr).WithArgs(p.Id, p.ReleaseID)
	mock.ExpectExec(query).WithArgs(p.Id, p.ReleaseID).WillReturnResult(sqlmock.NewResult(7, 1))

	err = p.Delete(db)
	assert.Equal(t, expErr, err)

	err = p.Delete(db)
	assert.Equal(t, nil, err)

	err = mock.ExpectationsWereMet()
	assert.Equal(t, nil, err)
}
