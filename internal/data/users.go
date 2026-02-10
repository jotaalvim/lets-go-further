package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"greenlight/internal/validator"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int       `json:"int"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Ativated  bool      `json:"activated"`
	Version   int       `json:"-"`
}

// the *string is for us to be able to distinguish a string not being present and a empty string as a password
type password struct {
	plaintext *string
	hash      []byte
}

var (
	ErrDuplicateEmail = errors.New("duplicated email")
	AnonymousUser     = &User{}
)

type UserModel struct {
	DB *sql.DB
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

func (m *UserModel) GetByEmail(email string) (*User, error) {

	query := `SELECT id, created_at, name, email, password_hash, activated,version
			  FROM users
			  WHERE email = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//err := m.DB.QueryRow(query, id).Scan(
	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Ativated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m *UserModel) Insert(user *User) error {

	query := `INSERT INTO users (name,email,password_hash, activated)
		      VALUES ( $1, $2, $3, $4 )
			  RETURNING id, created_at, version`

	args := []any{user.Name, user.Email, user.Password.hash, user.Ativated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)

	if err != nil {
		switch {
		case strings.HasPrefix(err.Error(), `pq: duplicate key value violates unique constraint "users_email_key"`):
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

func (m *UserModel) Update(user *User) error {

	query := `UPDATE users 
			  SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
			  WHERE id = $5 AND version = $6
			  RETURNING version`

	args := []any{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Ativated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (p *password) Set(plaintextPassword string) error {

	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

// Checks if the hash matches the plain text password in this struct
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "invalid email")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be bigger than 8")
	v.Check(len(password) <= 72, "password", "must be no longer than 72")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", user.Name, "must be provided")
	v.Check(len(user.Name) <= 500, user.Name, "cant be longer than 500")

	ValidateEmail(v, user.Email)
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

// GetForToken finds user that has a specific token
func (m *UserModel) GetForToken(scope string, tokenPlainText string) (*User, error) {

	tokenHash := sha256.Sum256([]byte(tokenPlainText))

	query := `SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version 
				FROM users
				INNER JOIN tokens
				ON users.id = tokens.user_id 
				WHERE tokens.hash = $1
				AND tokens.scope = $2
				AND tokens.expiry > $3 `

	args := []any{tokenHash[:], scope, time.Now()}

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Ativated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}
