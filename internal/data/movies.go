// Package data contains data models and database acess stuf
package data

import (
	"time"
)

// Movie - all fields with capital letter are exported and jsonencoding can see it
type Movie struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"-"` // hide this field
	Title     string    `json:"title"`
	Year      int       `json:"year,omitzero"`
	Runtime   Runtime   `json:"runtime,omitzero,string"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int       `json:"version"`
}
