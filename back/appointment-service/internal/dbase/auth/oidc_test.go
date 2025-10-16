package auth

import (
	"errors"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/dbase/test"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOIDCCreateUser(t *testing.T) {
	db := AuthStorage{test.InitTmpDB(t)}
	defer db.Close()

	in := OIDCData{
		Provider: "google",
		Subject:  "1234567890",
	}
	assert.True(t, in.IsValid())

	err := db.OIDCPairWithUser("", in)
	require.Equal(t, common.ErrInvalidArgument, err)
	err = db.OIDCPairWithUser("asdsad", OIDCData{})
	require.Equal(t, common.ErrInvalidArgument, err)

	_, err = db.OIDCCreateUser("", in)
	require.Equal(t, common.ErrInvalidArgument, err)
	_, err = db.OIDCCreateUser("asdsad", OIDCData{})
	require.Equal(t, common.ErrInvalidArgument, err)

	userID1, err := db.OIDCCreateUser("test_user1", in)
	require.NoError(t, err)
	require.NotEmpty(t, userID1)

	userID2, err := db.CreateUserPassword("test_user2", "password")
	require.NoError(t, err)

	err = db.OIDCPairWithUser(userID2, in)
	require.Error(t, err)

	in2 := OIDCData{
		Provider: "google",
		Subject:  "23232323",
	}
	err = db.OIDCPairWithUser(userID2, in2)
	require.NoError(t, err)

	err = db.OIDCUnPairUser(userID1, in)
	require.NoError(t, err)
	err = db.OIDCPairWithUser(userID2, in)
	require.NoError(t, err)

	tmp, err := db.OIDCUserAuth(in)
	assert.Equal(t, userID2, tmp)
	require.NoError(t, err)

	tmp, err = db.OIDCUserAuth(in2)
	assert.Equal(t, userID2, tmp)
	require.NoError(t, err)

	_, err = db.OIDCUserAuth(OIDCData{
		Provider: "google",
		Subject:  "55555",
	})
	require.True(t, errors.Is(err, common.ErrNotFound))
}
