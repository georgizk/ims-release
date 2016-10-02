package models

import (
	"database/sql"
	"errors"
	"os"
	"time"
)

// Errors pertaining to the data in a Page or operations on Pages.
var (
	ErrNoSuchPage = errors.New("Could not find page.")
)

// Database queries for operations on Pages.
const (
	QInitTablePage string = `create table if not exists pages (
		id int not null primary key,
		number varchar(255),
		location varchar(255),
		created_at timestamp,
		release_id int,
		foreign key(release_id) references releases(id)
);`

	QSavePage string = `insert into pages (
		number, location, created_at
) values (
		$1, $2, $3
);`

	QUpdatePage string = `update pages set
number = $2, location = $3, created_at = $4
where id = $1;`

	QDeletePage string = `delete from pages where id = $1;`

	QListPages string = `select (
		id, number, location, created_at
) from pages
where release_id = $1;`

	QFindPage string = `select (
		number, location, created_at
) from pages
where id = $1;`
)

// Page contains information about a single page of manga. Most important is its page number and location, which is the
// path to the page's image file on disk.
type Page struct {
	Id        int       `json:"id"`
	Number    string    `json:"page"`
	Location  string    `json:"-"` // Omit from JSON encodings
	CreatedAt time.Time `json:"createdAt"`
}

// NewPage constructs a brand new Project instance, with a default state lacking information about its (future)
// position in a database.
func NewPage(pageNum, filePath string) Page {
	return Page{
		0,
		pageNum,
		filePath,
		time.Now(),
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
		pages = append(pages, Page{id, number, location, created})
	}
	return pages, err
}

// Validate currently doesn't perform any integrity checks.
func (p Page) Validate() error {
	return nil
}

// Save inserts the page into the database and updates its Id field.
func (p *Page) Save(db *sql.DB) error {
	// TODO - Make sure to save image data to disk before saving the Page.
	_, err := db.Exec(QSavePage, p.Number, p.Location, p.CreatedAt)
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
	_, err := db.Exec(QUpdatePage, p.Id, p.Number, p.Location, p.CreatedAt)
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
