package models

import (
	"time"
)

// Page contains information about a single page of manga. Most important is its page number and location, which is the
// path to the page's image file on disk.
type Page struct {
	Id        string    `json:"id"`
	Number    int       `json:"page"`
	Location  string    `json:"-"` // Omit from JSON encodings
	CreatedAt time.Time `json:"createdAt"`
}

// Validate currently doesn't perform any integrity checks.
func (p Page) Validate() error {
	return nil
}
