package storage

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var ErrWrongPassword = bcrypt.ErrMismatchedHashAndPassword

const createUsersTable = `CREATE TABLE IF NOT EXISTS users (
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
func (db *Storage) CreateUser(user string, password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("[CreateUser] bcrypt error: %w", err)
	}

	_, err = db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", user, hashed)
	if err != nil {
		return fmt.Errorf("[CreateUser] db error: %w", err)
	}

	return nil
}

func (db *Storage) CheckUser(user string, password string) error {
	var storedHash string
	err := db.Get(&storedHash, "SELECT password FROM users WHERE username = $1", user)
	if err != nil {
		return fmt.Errorf("[CheckUser] db error: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return fmt.Errorf("[CheckUser] for [%v] unexpected error: %w", user, err)
	}

	return err
}
