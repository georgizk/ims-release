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
	MimeType  MimeType  `json:"-"`
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

	PGmax_len_name = 255
)

// NewPage constructs a brand new Project instance, with a default state lacking information about its (future)
// position in a database.
func NewPage(release Release, name string, tm time.Time) Page {
	return Page{
		Id:        0,
		Name:      name,
		CreatedAt: tm,
		ReleaseID: release.Id,
		MimeType:  MimeTypeFromFilename(name),
	}
}

// FindPage attempts to lookup a page by ID.
func FindPage(db database.DB, release Release, pageId uint32) (Page, error) {
	p := Page{ReleaseID: release.Id, Id: pageId}
	const query = "SELECT " + PGc_name + ", " + PGc_created_at +
		" FROM " + t_pages + " WHERE " + PGc_id + " = ? AND " + PGc_release_id + " = ?"
	row := db.QueryRow(query, pageId, release.Id)
	err := row.Scan(&p.Name, &p.CreatedAt)
	if err == database.ErrNoRows {
		return Page{}, ErrNoSuchPage
	} else if err != nil {
		return Page{}, err
	}
	p.MimeType = MimeTypeFromFilename(p.Name)
	return p, nil
}

func FindPageByName(db database.DB, release Release, name string) (Page, error) {
	p := Page{ReleaseID: release.Id, Name: name, MimeType: MimeTypeFromFilename(name)}
	const query = "SELECT " + PGc_id + ", " + PGc_created_at +
		" FROM " + t_pages + " WHERE " + PGc_release_id + " = ? AND " + PGc_name + " = ?"
	row := db.QueryRow(query, release.Id, name)
	err := row.Scan(&p.Id, &p.CreatedAt)
	if err == database.ErrNoRows {
		return Page{}, ErrNoSuchPage
	} else if err != nil {
		return Page{}, err
	}
	return p, nil
}

func GeneratePagePath(p Project, r Release, name string) string {
	return fmt.Sprintf("%d/%d/%s", p.Id, r.Id, name)
}

// ListPages attempts to obtain a list of all pages
func ListPages(db database.DB, release Release) ([]Page, error) {
	pages := []Page{}

	const query = "SELECT " + PGc_id + ", " + PGc_name + ", " + PGc_created_at +
		" FROM " + t_pages + " WHERE " + PGc_release_id + " = ?" +
		" ORDER BY " + PGc_name + " ASC"

	rows, err := db.Query(query, release.Id)
	if err != nil {
		return []Page{}, err
	}
	defer rows.Close()
	for rows.Next() {
		p := Page{ReleaseID: release.Id}
		err = rows.Scan(&p.Id, &p.Name, &p.CreatedAt)
		if err != nil {
			return pages, err
		}
		p.MimeType = MimeTypeFromFilename(p.Name)
		pages = append(pages, p)
	}
	err = rows.Err()
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
	if MimeTypeUnknown == MimeTypeFromFilename(p.Name) {
		return ErrPageUnsupportedMimeType
	}
	return nil
}

// Save inserts the page into the database and updates its Id field.
func SavePage(db database.DB, p Page) (Page, error) {
	validErr := p.Validate()
	if validErr != nil {
		return p, validErr
	}
	// TODO - Make sure to save image data to disk before saving the Page.

	const query = "INSERT INTO " + t_pages + " (" +
		PGc_name + ", " + PGc_created_at + ", " + PGc_release_id + ") VALUES (?, ?, ?)"

	res, err := db.Exec(query, p.Name, p.CreatedAt, p.ReleaseID)
	if err != nil {
		return p, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return p, err
	}
	p.Id = uint32(id)
	return p, nil
}

// Update modifies all of the fields of a Page in place with whatever is currently in the struct.
func UpdatePage(db database.DB, p Page) (Page, error) {
	return p, ErrOperationNotSupported
}

// Delete removes the Page from the database and deletes the page image from disk.
func DeletePage(db database.DB, p Page) (Page, error) {
	const query = "DELETE FROM " + t_pages + " WHERE " + PGc_id + " = ? AND " + PGc_release_id + " = ? LIMIT 1"
	_, err := db.Exec(query, p.Id, p.ReleaseID)
	return p, err
}
