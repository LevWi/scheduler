package storage

import (
	"errors"
	common "scheduler/appointment-service/internal"

	"github.com/google/uuid"
)

const createOIDCTable = `CREATE TABLE IF NOT EXISTS user_oidc (
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    subject TEXT NOT NULL,
    PRIMARY KEY (provider, subject)
);`

func CreateOIDCTable(db *Storage) error {
	_, err := db.Exec(createOIDCTable)
	if err != nil {
		return err
	}
	return nil
}

type OIDCData struct {
	UserName string
	Provider string
	Subject  string
}

func (db *Storage) OIDCCreateUser(in OIDCData) (UserID, error) {
	//TODO
	id := uuid.NewString()
	if in.UserName == "" { //Can UserName be null in "user" table?
		in.UserName = id
	}

	if in.Provider == "" || in.Subject == "" {
		return "", common.ErrInvalidArgument
	}

	tx, err := db.Begin()
	if err != nil {
		return "", adjustDbError(err)
	}

	_, err = tx.Exec("INSERT INTO users (id, username) VALUES ($1, $2)", id, in.UserName)
	if err != nil {
		goto Rollback
	}

	_, err = db.Exec("INSERT INTO user_oidc (user_id, provider, subject) VALUES ($1, $2, $3)", id, in.Provider, in.Subject)
	if err != nil {
		goto Rollback
	}

	if err := tx.Commit(); err != nil {
		return "", adjustDbError(err)
	}

	return UserID(id), nil

Rollback:
	re := tx.Rollback()
	if re != nil {
		err = errors.Join(err, tx.Rollback())
	}
	return "", adjustDbError(err)
}
