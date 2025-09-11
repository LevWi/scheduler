package storage

import (
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
	return adjustDbError(err)
}

func (db *Storage) AddUserToken(userID common.ID, token string, expiresAt time.Time) error {
	_, err := db.Exec(
		`INSERT INTO user_tokens (user_id, token, expires_at, is_used)
			VALUES ($1, $2, $3, FALSE)
		ON CONFLICT(user_id) DO UPDATE SET
			token = $2,
			expires_at = $3,
			is_used = FALSE`, string(userID), token, expiresAt.Unix())

	return adjustDbError(err)
}

func (db *Storage) ExchangeToken(token string) (common.ID, error) {
	tx, err := db.Beginx()
	if err != nil {
		return "", adjustDbError(err)
	}
	defer tx.Rollback()

	//Normally we would use Select -> check -> UPDATE
	//But SQLite doesn't support SELECT FOR UPDATE
	now := time.Now().Unix()
	res, err := tx.Exec(
		`UPDATE user_tokens
		SET is_used = TRUE
		WHERE token = $1
		  AND is_used = FALSE
		  AND expires_at > $2`,
		token, now)
	if err != nil {
		return "", adjustDbError(err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		var dbToken dbUserToken
		err = tx.Get(&dbToken,
			`SELECT user_id, is_used, expires_at
			FROM user_tokens
			WHERE token = $1`, token)
		if err != nil {
			return "", adjustDbError(err)
		}

		if dbToken.IsUsed {
			return "", ErrTokenAlreadyUsed
		}
		if now > dbToken.ExpiresAt {
			return "", ErrTokenExpired
		}
		return "", errors.New("token not found")
	}

	var userID string
	err = tx.Get(&userID,
		`SELECT user_id FROM user_tokens WHERE token = $1`, token)
	if err != nil {
		return "", adjustDbError(err)
	}

	if err = tx.Commit(); err != nil {
		return "", adjustDbError(err)
	}

	return common.ID(userID), nil
}

// CleanupExpiredTokens removes expired tokens from the database
func (db *Storage) CleanupExpiredTokens() error {
	_, err := db.Exec(`DELETE FROM user_tokens WHERE expires_at < $1`,
		time.Now().Unix())
	return adjustDbError(err)
}
