package models

import (
	"database/sql"
	"errors"
	"os"
	"time"
)

// Errors pertaining to the data in a Page or operations on Pages.
var (
	ErrNoSuchPage      = errors.New("Could not find page.")
	ErrPageNumberEmpty = errors.New("Page number is empty.")
)

// Database queries for operations on Pages.
const (
	QInitTablePage string = `create table if not exists pages (
		id int not null auto_increment,
		number varchar(255),
		location varchar(255),
		created_at timestamp,
		release_id int,
		foreign key(release_id) references releases(id),
		primary key(id)
);`

	QSavePage string = `insert into pages (
		number, location, created_at, release_id
) values (
		?, ?, ?, ?
);`

	QUpdatePage string = `update pages set
number = ?, location = ?, created_at = ?
where id = ?;`

	QDeletePage string = `delete from pages where id = ?;`

	QListPages string = `select
id, number, location, created_at
from pages
where release_id = ?
order by number asc;`

	QFindPage string = `select
number, location, created_at
from pages
where id = ?;`

	QLookupPage string = `select
P.id, P.location, P.created_at, P.release_id
from pages P
where number = ?
	and exists (
		select R.id
		from releases R
		where R.id = P.release_id
			and R.chapter = ?
			and R.version = ?
			and exists (
				select P2.id
				from projects P2
				where P2.id = R.project_id
					and P2.project_name = ?
			)
	);`
)

// Page contains information about a single page of manga. Most important is its page number and location, which is the
// path to the page's image file on disk.
type Page struct {
	Id        int       `json:"id"`
	Number    string    `json:"page"`
	Location  string    `json:"-"` // Omit from JSON encodings
	CreatedAt time.Time `json:"createdAt"`
	ReleaseID int       `json:"releaseId"`
}

// NewPage constructs a brand new Project instance, with a default state lacking information about its (future)
// position in a database.
func NewPage(pageNum, filePath string, releaseId int) Page {
	return Page{
		0,
		pageNum,
		filePath,
		time.Now(),
		releaseId,
	}
}

// FindPage attempts to lookup a page by ID.
func FindPage(id int, db *sql.DB) (Page, error) {
	p := Page{}
	row := db.QueryRow(QFindPage, id)
	if row == nil {
		return Page{}, ErrNoSuchPage
	}
	err := row.Scan(&p.Number, &p.Location, &p.CreatedAt)
	if err != nil {
		return Page{}, err
	}
	p.Id = id
	return p, nil
}

// LookupPage attempts to find a specific page assigned to a release in a project.
func LookupPage(pageNumber, releaseChapter string, releaseVersion int, projectName string, db *sql.DB) (Page, error) {
	p := Page{}
	row := db.QueryRow(QLookupPage, pageNumber, releaseChapter, releaseVersion, projectName)
	if row == nil {
		return Page{}, ErrNoSuchPage
	}
	err := row.Scan(&p.Id, &p.Location, &p.CreatedAt, &p.ReleaseID)
	if err != nil {
		return Page{}, err
	}
	p.Number = pageNumber
	return p, nil
}

// ListPages attempts to obtain a list of all pages
func ListPages(releaseId int, db *sql.DB) ([]Page, error) {
	pages := []Page{}
	rows, err := db.Query(QListPages, releaseId)
	if err != nil {
		return []Page{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var number, location string
		var created time.Time
		scanErr := rows.Scan(&id, &number, &location, &created)
		if scanErr != nil {
			err = scanErr
		}
		pages = append(pages, Page{id, number, location, created, releaseId})
	}
	return pages, err
}

// Validate currently doesn't perform any integrity checks.
func (p *Page) Validate() error {
	if len(p.Number) == 0 {
		return ErrPageNumberEmpty
	}
	return nil
}

// Save inserts the page into the database and updates its Id field.
func (p *Page) Save(db *sql.DB) error {
	validErr := p.Validate()
	if validErr != nil {
		return validErr
	}
	// TODO - Make sure to save image data to disk before saving the Page.
	_, err := db.Exec(QSavePage, p.Number, p.Location, p.CreatedAt, p.ReleaseID)
	if err != nil {
		return err
	}
	row := db.QueryRow(QLastInsertID)
	if row == nil {
		return ErrCouldNotGetID
	}
	return row.Scan(&p.Id)
}

// Update modifies all of the fields of a Page in place with whatever is currently in the struct.
func (p *Page) Update(db *sql.DB) error {
	_, err := db.Exec(QUpdatePage, p.Number, p.Location, p.CreatedAt, p.Id)
	return err
}

// Delete removes the Page from the database and deletes the page image from disk.
func (p *Page) Delete(db *sql.DB) error {
	_, err := db.Exec(QDeletePage, p.Id)
	rmErr := os.Remove(p.Location)
	if err != nil {
		return err
	}
	return rmErr
}
