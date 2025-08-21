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
var ErrEmptyPassword = fmt.Errorf("%w: password empty", common.ErrNotAllowed)

const createUsersTable = `CREATE TABLE IF NOT EXISTS users (
	id TEXT PRIMARY KEY,
	username TEXT NOT NULL UNIQUE,
	pwd_hash TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`

//TODO need updated_at handling

type UserID common.ID

func CreateUsersTable(db *Storage) error {
	_, err := db.Exec(createUsersTable)
	if err != nil {
		return err
	}
	return nil
}

// TODO add checking requirements for password symbols somewhere
func (db *Storage) CreateUserPassword(user string, password string) (UserID, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("[CreateUser] bcrypt error: %w", err)
	}

	//TODO Handle error instead panic ?
	id := uuid.NewString()

	_, err = db.Exec("INSERT INTO users (id, username, pwd_hash) VALUES ($1, $2, $3)", id, user, hashed)
	return UserID(id), adjustDbError(err)
}

func (db *Storage) UpdateUserPassword(id UserID, oldPword string, newPword string) error {
	u, err := db.readUserByID(id)
	if err != nil {
		return err
	}
	err = checkPassword(u.PwdHash, oldPword)
	if err != nil {
		return err
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("[UpdateUserPassword] bcrypt error: %w", err)
	}

	_, err = db.Exec("UPDATE users SET pwd_hash = $1 WHERE id = $2", newHash, id)
	return adjustDbError(err)
}

func (db *Storage) CheckUserPassword(user, password string) (UserID, error) {
	u, err := db.readUser(user)
	if err != nil {
		return "", err
	}
	return UserID(u.Id), checkPassword(u.PwdHash, password)
}

func (db *Storage) DeleteUser(id UserID) error {
	//TODO don't delete user. Update a status cell value
	_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
	return adjustDbError(err)
}

func (db *Storage) DeleteUserWithCheck(user string, password string) error {
	u, err := db.readUser(user)
	if err != nil {
		return err
	}
	err = checkPassword(u.PwdHash, password)
	if err != nil {
		return err
	}
	return db.DeleteUser(UserID(u.Id))
}

func (db *Storage) IsExist(id UserID) error {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM users WHERE id = $1", id)
	if err != nil {
		return adjustDbError(err)
	}
	if count == 0 {
		return common.ErrNotFound
	}
	return nil
}

type dbUser struct {
	Id       string `db:"id"`
	Username string `db:"username"`
	PwdHash  string `db:"pwd_hash"`
}

func (u dbUser) User() User {
	return User{
		Username: u.Username,
		Id:       UserID(u.Id),
		PwdHash:  u.PwdHash,
	}
}

type User struct {
	Username string
	Id       UserID
	PwdHash  string
}

// TODO add context?
func adjustDbError(e error) error {
	if e == nil {
		return nil
	}
	if errors.Is(e, sql.ErrNoRows) {
		e = common.ErrNotFound
	}
	return fmt.Errorf("db error: %w", e)
}

func (db *Storage) readUser(user string) (dbUser, error) {
	var dbUser dbUser
	err := db.Get(&dbUser, "SELECT id, username, pwd_hash FROM users WHERE username = $1", user)
	return dbUser, adjustDbError(err)
}

func (db *Storage) readUserByID(id UserID) (dbUser, error) {
	var dbUser dbUser
	err := db.Get(&dbUser, "SELECT id, username, pwd_hash FROM users WHERE id = $1", id)
	return dbUser, adjustDbError(err)
}

func checkPassword(hash, password string) error {
	if password == "" {
		return ErrEmptyPassword
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		err = common.ErrUnauthorized
	}
	return err
}
