package storage

import (
	"fmt"
	"math/rand"
	common "scheduler/appointment-service/internal"
	"time"

	"github.com/google/uuid"
)

type UserID = common.ID

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
	Provider string
	Subject  string
}

func (d OIDCData) IsValid() bool {
	return d.Provider != "" && d.Subject != ""

}

// TODO move it
func GenerateUsername() string {
	return fmt.Sprintf("user_%x_%d", rand.Uint64(), time.Now().Unix())
}

// TODO need algorithm for names collision
func (db *Storage) OIDCCreateUser(userName string, in OIDCData) (UserID, error) {
	if userName == "" || !in.IsValid() {
		return "", common.ErrInvalidArgument
	}

	id := uuid.NewString()
	tx, err := db.Begin()
	if err != nil {
		return "", adjustDbError(err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO users (id, username) VALUES ($1, $2)", id, userName)
	if err != nil {
		return "", adjustDbError(err)
	}

	_, err = tx.Exec("INSERT INTO user_oidc (user_id, provider, subject) VALUES ($1, $2, $3)", id, in.Provider, in.Subject)
	if err != nil {
		return "", adjustDbError(err)
	}

	if err := tx.Commit(); err != nil {
		return "", adjustDbError(err)
	}

	return UserID(id), nil
}

func (db *Storage) OIDCPairWithUser(uid UserID, in OIDCData) error {
	if uid == "" || !in.IsValid() {
		return common.ErrInvalidArgument
	}

	_, err := db.Exec("INSERT INTO user_oidc (user_id, provider, subject) VALUES ($1, $2, $3)", uid, in.Provider, in.Subject)
	if err != nil {
		return adjustDbError(err)
	}

	return nil
}

func (db *Storage) OIDCUnPairUser(uid UserID, in OIDCData) error {
	if uid == "" || !in.IsValid() {
		return common.ErrInvalidArgument
	}

	_, err := db.Exec("DELETE FROM user_oidc WHERE user_id = $1 AND provider = $2 AND subject = $3", uid, in.Provider, in.Subject)
	if err != nil {
		return adjustDbError(err)
	}

	return nil
}

func (db *Storage) OIDCUserAuth(in OIDCData) (UserID, error) {
	if !in.IsValid() {
		return "", common.ErrInvalidArgument
	}

	var userID UserID
	err := db.Get(&userID, "SELECT user_id FROM user_oidc WHERE provider = $1 AND subject = $2", in.Provider, in.Subject)
	if err != nil {
		return "", adjustDbError(err)
	}

	return userID, nil

}
