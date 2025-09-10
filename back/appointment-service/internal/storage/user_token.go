package storage

import (
	"database/sql"
	"errors"
	"time"

	common "scheduler/appointment-service/internal"
)

var (
	ErrTokenAlreadyUsed = errors.New("token already used")
	ErrTokenExpired     = errors.New("token expired")
)

type dbUserToken struct {
	Token     string `db:"token"`
	UserID    string `db:"user_id"`
	ExpiresAt int64  `db:"expires_at"`
	IsUsed    bool   `db:"is_used"`
}

const queryCreateUserTokensTable = `CREATE TABLE IF NOT EXISTS "user_tokens" (
	"token"      TEXT NOT NULL UNIQUE,
	"user_id"    TEXT NOT NULL UNIQUE,
	"expires_at" INTEGER NOT NULL,
	"is_used"    BOOLEAN NOT NULL DEFAULT FALSE,
	PRIMARY KEY ("token")
);`

func CreateTableUserTokens(db *Storage) error {
	_, err := db.Exec(queryCreateUserTokensTable)
	return err
}

// TODO fix
func (db *Storage) AddUserToken(userID common.ID, token string, expiresAt time.Time) error {
	// First try to update existing token for the user
	result, err := db.Exec(`
		UPDATE user_tokens 
		SET token = $1, expires_at = $2, is_used = FALSE 
		WHERE user_id = $3
	`, token, expiresAt.Unix(), string(userID))

	if err != nil {
		return err
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no rows were updated, insert a new record
	if rowsAffected == 0 {
		_, err = db.Exec(`
			INSERT INTO user_tokens (token, user_id, expires_at, is_used)
			VALUES ($1, $2, $3, FALSE)
		`, token, string(userID), expiresAt.Unix())
		return err
	}

	return nil
}

// TODO fix it. Begin / Commit
func (db *Storage) ExchangeToken(token string) (common.ID, error) {
	var dbToken dbUserToken

	// Get token from database
	err := db.Get(&dbToken, `
		SELECT token, user_id, expires_at, is_used 
		FROM user_tokens 
		WHERE token = $1
	`, token)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", common.ErrNotFound
		}
		return "", err
	}

	// Check if token is already used
	if dbToken.IsUsed {
		return "", ErrTokenAlreadyUsed
	}

	// Check if token is expired
	if time.Now().Unix() > dbToken.ExpiresAt {
		return "", ErrTokenExpired
	}

	// Mark token as used
	_, err = db.Exec(`
		UPDATE user_tokens 
		SET is_used = TRUE 
		WHERE token = $1
	`, token)

	if err != nil {
		return "", err
	}

	return common.ID(dbToken.UserID), nil
}

// CleanupExpiredTokens removes expired tokens from the database
func (db *Storage) CleanupExpiredTokens() error {
	_, err := db.Exec(`
		DELETE FROM user_tokens 
		WHERE expires_at < $1
	`, time.Now().Unix())
	return err
}
