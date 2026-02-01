// Package data contains data models and database acess stuf
package data

import (
	"time"

	"greenlight/internal/validator"
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

func ValidateMovie(v *validator.Validator, movie *Movie) {

	v.Check(movie.Title != "", "title", "this field cannot be empty")
	v.Check(len(movie.Title) < 500, "title", "this field must be smaller than500 chars")

	v.Check(movie.Year != 0, "year", "this field cannot be empty")
	v.Check(movie.Year >= 1888, "year", "must be grater then 1888")
	v.Check(movie.Year <= time.Now().Year(), "year", "this field must nor be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be provided")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")

	v.Check(validator.UniqueValues(movie.Genres), "genres", "must not contain duplicate values")

}
