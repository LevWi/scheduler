package auth

import (
	"errors"
	"time"

	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/dbase"
)

var (
	ErrTokenAlreadyUsed = errors.New("token already used")
	ErrTokenExpired     = errors.New("token expired")
)

type ClientID = common.ID
type BusinessID = common.ID

type TokenEntry struct {
	Token      string `db:"token"`
	BusinessID string `db:"business_id"`
	ClientID   string `db:"client_id"`
	ExpiresAt  int64  `db:"expires_at"`
	IsUsed     bool   `db:"is_used"`
}

func (db *AuthStorage) AddUserToken(businessID BusinessID, clientID ClientID,
	token string, expiresAt time.Time) error {
	_, err := db.Exec(
		`INSERT INTO user_tokens (business_id, client_id, token, expires_at, is_used)
			VALUES ($1, $2, $3, $4, FALSE)
		ON CONFLICT(business_id, client_id) DO UPDATE SET
			token = $3,
			expires_at = $4,
			is_used = FALSE`, businessID, clientID, token, expiresAt.Unix())

	return dbase.DbError(err)
}

func (db *AuthStorage) ExchangeToken(token string) (*TokenEntry, error) {
	tx, err := db.Beginx()
	if err != nil {
		return nil, dbase.DbError(err)
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
		return nil, dbase.DbError(err)
	}

	rows, _ := res.RowsAffected()
	var dbToken TokenEntry
	if rows == 0 {
		//TODO add tests
		err = tx.Get(&dbToken,
			`SELECT business_id, client_id, is_used, expires_at
			FROM user_tokens
			WHERE token = $1`, token)
		if err != nil {
			return nil, dbase.DbError(err)
		}

		if dbToken.IsUsed {
			return nil, ErrTokenAlreadyUsed
		}
		if now > dbToken.ExpiresAt {
			return nil, ErrTokenExpired
		}

		return nil, errors.New("[ExchangeToken]: unexpected result")
	}

	err = tx.Get(&dbToken,
		`SELECT business_id, client_id, is_used, expires_at FROM user_tokens WHERE token = $1`, token)
	if err != nil {
		return nil, dbase.DbError(err)
	}

	if err = tx.Commit(); err != nil {
		return nil, dbase.DbError(err)
	}

	return &dbToken, nil
}

func (db *AuthStorage) CleanupExpiredTokens() error {
	_, err := db.Exec(`DELETE FROM user_tokens WHERE expires_at < $1`,
		time.Now().Unix())
	return dbase.DbError(err)
}
