package models

import (
	"time"
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
		"",
		pageNum,
		filePath,
		time.Now(),
	}
}

// Validate currently doesn't perform any integrity checks.
func (p Page) Validate() error {
	return nil
}
