package storage

import (
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
	password TEXT NOT NULL
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

	_, err = db.Exec("INSERT INTO users_pwd (id, username, password) VALUES ($1, $2, $3)", id, user, hashed)
	if err != nil {
		return fmt.Errorf("[CreateUser] db error: %w", err)
	}

	return nil
}

func (db *Storage) UpdateUserPassword(user string, oldPword string, newPword string) error {
	err := db.CheckUserPassword(user, oldPword)
	if err != nil {
		return err
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("[UpdateUserPassword] bcrypt error: %w", err)
	}

	_, err = db.Exec("UPDATE users_pwd SET password = $1 WHERE username = $2", newHash, user)
	if err != nil {
		return fmt.Errorf("[UpdateUserPassword] db error: %w", err)
	}

	return nil
}

func (db *Storage) DeleteUser(user string, password string) error {
	err := db.CheckUserPassword(user, password)
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM users_pwd WHERE username = $1", user)
	if err != nil {
		return fmt.Errorf("[DeleteUser] db error: %w", err)
	}
	return nil

}

func (db *Storage) IsExist(uuid common.ID) error {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM users_pwd WHERE id = $1", uuid)
	if err != nil {
		return fmt.Errorf("[IsExist] db error: %w", err)
	}
	if count == 0 {
		return common.ErrNotFound
	}
	return nil
}

func (db *Storage) CheckUserPassword(user string, password string) error {
	var storedHash string
	err := db.Get(&storedHash, "SELECT password FROM users_pwd WHERE username = $1", user)
	if err != nil {
		return fmt.Errorf("[CheckUser] db error: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return fmt.Errorf("[CheckUser] for [%v] unexpected error: %w", user, err)
	}

	return err
}
