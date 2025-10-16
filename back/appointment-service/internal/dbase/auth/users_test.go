package auth

import (
	"errors"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/dbase/test"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUserPassword(t *testing.T) {
	db := AuthStorage{test.InitTmpDB(t)}
	defer db.Close()

	user := "test_user"
	password := "test_password"

	_, err := db.CreateUserPassword(user, password)
	assert.NoError(t, err)
}

func TestUserPasswordEmpty(t *testing.T) {
	db := AuthStorage{test.InitTmpDB(t)}
	defer db.Close()

	user := "test_user"

	_, err := db.CreateUserPassword(user, "")
	assert.Equal(t, ErrEmptyPassword, err)

	_, err = db.CheckUserPassword(user, "")
	assert.Equal(t, errors.Is(err, common.ErrNotFound), true)

	_, err = db.CreateUserPassword(user, "12345")
	assert.NoError(t, err)
	_, err = db.CheckUserPassword(user, "")
	assert.Equal(t, ErrEmptyPassword, err)
}

func TestCheckUserPassword(t *testing.T) {
	db := AuthStorage{test.InitTmpDB(t)}
	defer db.Close()

	user := "test_user"
	password := "test_password"

	id, err := db.CreateUserPassword(user, password)
	assert.NoError(t, err)

	_, err = db.CheckUserPassword(user, password+" ")
	assert.Error(t, err)
	assert.Equal(t, common.ErrUnauthorized, err)

	_, err = db.CheckUserPassword(user, password+"123")
	assert.Error(t, err)
	assert.Equal(t, common.ErrUnauthorized, err)

	oid, err := db.CheckUserPassword(user, password)
	assert.NoError(t, err)
	assert.Equal(t, id, oid)
}

func TestUpdateUserPassword(t *testing.T) {
	db := AuthStorage{test.InitTmpDB(t)}
	defer db.Close()

	user := "test_user"
	oldPassword := "test_password"
	newPassword := "new_test_password"

	id, err := db.CreateUserPassword(user, oldPassword)
	assert.NoError(t, err)

	oid, err := db.CheckUserPassword(user, oldPassword)
	assert.NoError(t, err)
	assert.Equal(t, id, oid)

	err = db.UpdateUserPassword(id, oldPassword, newPassword)
	assert.NoError(t, err)

	oid, err = db.CheckUserPassword(user, newPassword)
	assert.NoError(t, err)
	assert.Equal(t, id, oid)

	err = db.UpdateUserPassword(id, oldPassword, newPassword)
	assert.Error(t, err)

	oid, err = db.CheckUserPassword(user, newPassword)
	assert.NoError(t, err)
	assert.Equal(t, id, oid)
}

func TestExistAndDelete(t *testing.T) {
	db := AuthStorage{test.InitTmpDB(t)}
	defer db.Close()

	user := "test_user"
	password := "test_password"

	err := db.IsExist("")
	assert.Error(t, err)
	assert.Equal(t, common.ErrNotFound, err)

	err = db.IsExist("1234")
	assert.Error(t, err)
	assert.Equal(t, common.ErrNotFound, err)

	id, err := db.CreateUserPassword(user, password)
	assert.NoError(t, err)

	err = db.IsExist(id)
	assert.NoError(t, err)

	err = db.IsExist(id + "2")
	assert.Error(t, err)
	assert.Equal(t, common.ErrNotFound, err)

	err = db.DeleteUser(id)
	assert.NoError(t, err)

	err = db.IsExist(id)
	assert.Error(t, err)
	assert.Equal(t, common.ErrNotFound, err)
}

func TestDeleteUser(t *testing.T) {
	db := AuthStorage{test.InitTmpDB(t)}
	defer db.Close()

	user := "test_user"
	password := "test_password"

	id, err := db.CreateUserPassword(user, password)
	assert.NoError(t, err)
	err = db.IsExist(id)
	assert.NoError(t, err)

	err = db.DeleteUserWithCheck(user, password+"bla")
	assert.Error(t, err)
	err = db.IsExist(id)
	assert.NoError(t, err)

	err = db.DeleteUserWithCheck(user, password)
	assert.NoError(t, err)
	err = db.IsExist(id)
	assert.Error(t, err)
	assert.Equal(t, common.ErrNotFound, err)
}
