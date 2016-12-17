package models

import (
	"errors"
	"fmt"
	"ims-release/database"
	"strings"
	"time"
)

// Page contains information about a single page of manga. Most important is its page name, which is the
// path to the page's image file on disk.
type Page struct {
	Id        uint32    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	ReleaseID uint32    `json:"-"`
	MimeType  string    `json:"mimeType"`
}

type MimeType uint32

const (
	MimeTypeUnknown    MimeType = 0
	MimeTypeUnknownStr string   = "application/octet-stream"
	MimeTypePng        MimeType = 1
	MimeTypePngStr     string   = "image/png"
	MimeTypeJpg        MimeType = 2
	MimeTypeJpgStr     string   = "image/jpeg"
)

func (t MimeType) String() string {
	switch t {
	case MimeTypePng:
		return MimeTypePngStr
	case MimeTypeJpg:
		return MimeTypeJpgStr
	default:
		return MimeTypeUnknownStr
	}
}

func NewMimeType(val string) MimeType {
	switch val {
	case MimeTypePngStr:
		return MimeTypePng
	case MimeTypeJpgStr:
		return MimeTypeJpg
	default:
		return MimeTypeUnknown
	}
}

func MimeTypeFromFilename(filename string) MimeType {
	if strings.HasSuffix(filename, ".png") {
		return MimeTypePng
	} else if strings.HasSuffix(filename, ".jpg") {
		return MimeTypeJpg
	}
	return MimeTypeUnknown
}

// Errors pertaining to the data in a Page or operations on Pages.
var (
	ErrNoSuchPage              = errors.New("Could not find page.")
	ErrPageNameEmpty           = errors.New("Page name is empty.")
	ErrPageNameTooLong         = errors.New("Page name is too long.")
	ErrPageUnsupportedMimeType = errors.New("Unsupported mime type.")
)

// Database queries for operations on Pages.
const (
	t_pages        string = "`pages`"
	PGc_id         string = "`id`"
	PGc_name       string = "`name`"
	PGc_created_at string = "`created_at`"
	PGc_release_id string = "`release_id`"
	PGc_mime_type  string = "`mime_type`"

	PGmax_len_name = 255
)

// NewPage constructs a brand new Project instance, with a default state lacking information about its (future)
// position in a database.
func NewPage(name string, releaseId uint32, mimeType MimeType) Page {
	return Page{
		0,
		name,
		time.Now(),
		releaseId,
		mimeType.String(),
	}
}

// FindPage attempts to lookup a page by ID.
func FindPage(db database.DB, releaseId uint32, pageId uint32) (Page, error) {
	p := Page{}
	const query = "SELECT " + PGc_name + ", " + PGc_created_at + ", " + PGc_mime_type +
		" FROM " + t_pages + " WHERE " + PGc_id + " = ? AND " + PGc_release_id + " = ?"
	row := db.QueryRow(query, pageId, releaseId)
	if row == nil {
		return Page{}, ErrNoSuchPage
	}
	var t MimeType
	err := row.Scan(&p.Name, &p.CreatedAt, &t)
	if err != nil {
		return Page{}, err
	}
	p.MimeType = t.String()
	p.Id = pageId
	return p, nil
}

func FindPageByName(db database.DB, releaseId uint32, name string) (Page, error) {
	p := Page{}
	const query = "SELECT " + PGc_id + ", " + PGc_name + ", " + PGc_created_at + ", " + PGc_mime_type +
		" FROM " + t_pages + " WHERE " + PGc_release_id + " = ? AND " + PGc_name + " = ?"
	row := db.QueryRow(query, releaseId, name)
	if row == nil {
		return Page{}, ErrNoSuchPage
	}
	var t MimeType
	err := row.Scan(&p.Id, &p.Name, &p.CreatedAt, &t)
	if err != nil {
		return Page{}, err
	}
	p.MimeType = t.String()
	return p, nil
}

func GeneratePagePath(p Project, r Release, name string) string {
	return fmt.Sprintf("%d/%d/%s", p.Id, r.Id, name)
}

// ListPages attempts to obtain a list of all pages
func ListPages(db database.DB, releaseId uint32) ([]Page, error) {
	pages := []Page{}

	const query = "SELECT " + PGc_id + ", " + PGc_name + ", " + PGc_created_at + ", " + PGc_mime_type +
		" FROM " + t_pages + " WHERE " + PGc_release_id + " = ?" +
		" ORDER BY " + PGc_name + " ASC"

	rows, err := db.Query(query, releaseId)
	if err != nil {
		return []Page{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var id uint32
		var name string
		var created time.Time
		var t MimeType
		scanErr := rows.Scan(&id, &name, &created, &t)
		if scanErr != nil {
			err = scanErr
		}
		pages = append(pages, Page{id, name, created, releaseId, t.String()})
	}
	return pages, err
}

// Validate currently doesn't perform any integrity checks.
func (p *Page) Validate() error {
	if len(p.Name) == 0 {
		return ErrPageNameEmpty
	}
	if len(p.Name) > PGmax_len_name {
		return ErrPageNameTooLong
	}

	if MimeTypeUnknown == NewMimeType(p.MimeType) {
		return ErrPageUnsupportedMimeType
	}
	return nil
}

// Save inserts the page into the database and updates its Id field.
func (p *Page) Save(db database.DB) error {
	validErr := p.Validate()
	if validErr != nil {
		return validErr
	}
	// TODO - Make sure to save image data to disk before saving the Page.

	const query = "INSERT INTO " + t_pages + " (" +
		PGc_name + ", " + PGc_created_at + ", " + PGc_release_id + ", " + PGc_mime_type + ") VALUES (?, ?, ?, ?)"

	res, err := db.Exec(query, p.Name, p.CreatedAt, p.ReleaseID, NewMimeType(p.MimeType))
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

// Update modifies all of the fields of a Page in place with whatever is currently in the struct.
func (p *Page) Update(db database.DB) error {
	return ErrOperationNotSupported
}

// Delete removes the Page from the database and deletes the page image from disk.
func (p *Page) Delete(db database.DB) error {
	const query = "DELETE FROM " + t_pages + " WHERE " + PGc_id + " = ? LIMIT 1"
	_, err := db.Exec(query, p.Id)
	return err
}
