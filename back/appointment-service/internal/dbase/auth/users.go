package auth

import (
	"errors"
	"fmt"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/dbase"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrWrongPassword = bcrypt.ErrMismatchedHashAndPassword
var ErrEmptyPassword = fmt.Errorf("%w: password empty", common.ErrNotAllowed)

// TODO need updated_at handling
// TODO add checking requirements for password symbols somewhere
func (db *AuthStorage) CreateUserPassword(user string, password string) (UserID, error) {
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
	return UserID(id), dbase.DbError(err)
}

func (db *AuthStorage) UpdateUserPassword(id UserID, oldPword string, newPword string) error {
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
	return dbase.DbError(err)
}

func (db *AuthStorage) CheckUserPassword(user, password string) (UserID, error) {
	u, err := db.readUser(user)
	if err != nil {
		return "", err
	}
	return UserID(u.Id), checkPassword(u.PwdHash, password)
}

func (db *AuthStorage) DeleteUser(id UserID) error {
	//TODO don't delete user. Update a status cell value
	_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
	return dbase.DbError(err)
}

func (db *AuthStorage) DeleteUserWithCheck(user string, password string) error {
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

func (db *AuthStorage) IsExist(id UserID) error {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM users WHERE id = $1", id)
	if err != nil {
		return dbase.DbError(err)
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

func (db *AuthStorage) readUser(user string) (dbUser, error) {
	var dbUser dbUser
	err := db.Get(&dbUser, "SELECT id, username, pwd_hash FROM users WHERE username = $1", user)
	return dbUser, dbase.DbError(err)
}

func (db *AuthStorage) readUserByID(id UserID) (dbUser, error) {
	var dbUser dbUser
	err := db.Get(&dbUser, "SELECT id, username, pwd_hash FROM users WHERE id = $1", id)
	return dbUser, dbase.DbError(err)
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
