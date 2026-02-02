// Package data contains data models and database acess stuf
package data

import (
	"database/sql"
	"errors"
	"time"

	"greenlight/internal/validator"

	"github.com/lib/pq"
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

type MovieModel struct {
	DB *sql.DB
}

func (m MovieModel) Insert(movie *Movie) error {
	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id,created_at, version `
	// pq implements the drivers to convert our slice of strings to postgres text[]
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at,title, year, runtime, genres, version 
		FROM movies
		WHERE id = $1`

	var movie Movie

	err := m.DB.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

func (m MovieModel) Update(movie *Movie) error {
	return nil
}

func (m MovieModel) Delete(id int) error {
	return nil
}
