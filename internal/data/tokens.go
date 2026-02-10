package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"greenlight/internal/validator"
	"time"
)

// ScopeActivation holds what the token is beeing used for as we're going to reutilize the token table for other types of tokens

var (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)


func ValidateTokenPlainText(v *validator.Validator, tokenPlainText string) {
	v.Check(tokenPlainText != "", "token", "must be provided")
	v.Check(len(tokenPlainText) == 26, "token", "must be longer then 26 chars")
}

type TokenModel struct {
	DB *sql.DB
}

type Token struct {

	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int       `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`

}

// New generates a token and inserts into tokens database table
func (m *TokenModel) New(userID int, ttl time.Duration, scope string) (*Token, error) {

	token := generateToken(userID, ttl, scope)

	err := m.Insert(token)
	return token, err

}

func (m *TokenModel) Insert(token *Token) error {

	query := `INSERT INTO tokens (hash , user_id, expiry, scope) 
			  VALUES ($1, $2, $3, $4)`

	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err

}

// DeleteAllForUser deletes all tokens of a specific scope for a use
func (m *TokenModel) DeleteAllForUser(scope string, userID int) error {
	query := `DELETE FROM tokens
	          WHERE user_id = $2 AND scope = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}

// generateToken creates a random token
func generateToken(userID int, ttl time.Duration, scope string) *Token {
	token := &Token{
		Plaintext: rand.Text(),
		UserID:    userID,
		Expiry:    time.Now().Add(ttl),
		Scope:     scope,
	}
	hash := sha256.Sum256([]byte(token.Plaintext))
	// we convert the array to a slice using [:] notation
	token.Hash = hash[:]
	return token

}
