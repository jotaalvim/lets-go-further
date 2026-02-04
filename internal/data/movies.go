// Package data contains data models and database acess stuf
package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

type MovieModel struct {
	DB *sql.DB
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

func (m MovieModel) Insert(movie *Movie) error {
	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id,created_at, version `
	// pq implements the drivers to convert our slice of strings to postgres text[]
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
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

	// context that holds a 3 second timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// cancel releases the resources associated with the context, otherwise after 3 seconds the resources will not be released
	defer cancel()

	//err := m.DB.QueryRow(query, id).Scan(
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
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
	// use uuid_generate_v4() so that the version is't guessable
	query := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version `

	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// check if no mathcing rows have been found, the version has changed
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m MovieModel) Delete(id int) error {

	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
	DELETE  FROM movies
	WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// if no rows affected id was not in database
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, error) {
	query := fmt.Sprintf(`
		SELECT id, created_at,title, year, runtime, genres, version 
		FROM movies
		WHERE ( to_tsvector('english', title) @@ plainto_tsquery('english', $1) OR $1 = '')
		AND   ( genres @> $2 OR $2 = '{}')
		ORDER BY %s %s , id ASC`, filters.sortCollumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, title, pq.Array(genres))
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	movies := []*Movie{}
	for rows.Next() {

		var movie Movie
		err := rows.Scan(
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, err
		}
		movies = append(movies, &movie)
	}
	if rows.Err() != nil {
		return nil, err
	}

	return movies, nil

}
