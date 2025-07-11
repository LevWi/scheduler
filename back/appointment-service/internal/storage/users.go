package storage

import (
	"database/sql"
	"errors"
	"fmt"
	common "scheduler/appointment-service/internal"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrWrongPassword = bcrypt.ErrMismatchedHashAndPassword

const createUsersTable = `CREATE TABLE IF NOT EXISTS users_pwd (
	id INTEGER PRIMARY KEY,
	username TEXT NOT NULL UNIQUE,
	pass_hash TEXT NOT NULL
);`

func CreateUsersTable(db *Storage) error {
	_, err := db.Exec(createUsersTable)
	if err != nil {
		return err
	}
	return nil
}

// TODO add checking requirements for password symbols somewhere
func (db *Storage) CreateUserPassword(user string, password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("[CreateUser] bcrypt error: %w", err)
	}

	//TODO Handle error instead panic ?
	id := uuid.NewString()

	_, err = db.Exec("INSERT INTO users_pwd (id, username, pass_hash) VALUES ($1, $2, $3)", id, user, hashed)
	if err != nil {
		return fmt.Errorf("[CreateUser] db error: %w", err)
	}

	return nil
}

func (db *Storage) UpdateUserPassword(id common.ID, oldPword string, newPword string) error {
	_, err := db.CheckUserPassword(id, oldPword)
	if err != nil {
		return err
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("[UpdateUserPassword] bcrypt error: %w", err)
	}

	_, err = db.Exec("UPDATE users_pwd SET pass_hash = $1 WHERE id = $2", newHash, id)
	if err != nil {
		return fmt.Errorf("[UpdateUserPassword] db error: %w", err)
	}

	return nil
}

func (db *Storage) DeleteUser(id common.ID, password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("[CreateUser] bcrypt error: %w", err)
	}

	_, err = db.Exec("DELETE FROM users_pwd WHERE id = $1 AND pass_hash = $2", id, hashed)
	if err != nil {
		return fmt.Errorf("[DeleteUser] db error: %w", err)
	}
	return nil
}

func (db *Storage) IsExist(id common.ID) error {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM users_pwd WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("[IsExist] db error: %w", err)
	}
	if count == 0 {
		return common.ErrNotFound
	}
	return nil
}

func (db *Storage) CheckUserPassword(user string, password string) (common.ID, error) {
	type DBUser struct {
		id        string
		pass_hash string
	}

	var dbUser DBUser
	err := db.Get(&dbUser, "SELECT id, pass_hash FROM users_pwd WHERE username = $1", user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", common.ErrNotFound
		}
		return "", fmt.Errorf("[CheckUserPassword] db error: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.pass_hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", ErrWrongPassword
		}
		return "", fmt.Errorf("[CheckUserPassword] unexpected error: %w", err)
	}

	return dbUser.id, nil
}
