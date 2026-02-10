// Package data is ...
package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflit")
)

// this struct is so that in future it's easier to add new models types to the app

type Models struct {

	Movies      MovieModel
	Users       UserModel
	Tokens      TokenModel
	Permissions PermissionsModel

}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies:      MovieModel{DB: db},
		Users:       UserModel{DB: db},
		Tokens:      TokenModel{DB: db},
		Permissions: PermissionsModel{DB: db},
	}
}
