package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/dbase/test"
)

func TestUserToken(t *testing.T) {
	storage := AuthStorage{test.InitTmpDB(t)}
	defer storage.Close()

	businessID := "user1"
	clientID := "client1"
	token := "token1"
	expiresAt := time.Now().Add(1 * time.Hour)

	// Add a new token
	err := storage.AddUserToken(businessID, clientID, token, expiresAt)
	assert.NoError(t, err)

	// Try to add the same token again, should update
	err = storage.AddUserToken(businessID, clientID, "new-token", expiresAt)
	assert.NoError(t, err)

	// Exchange the token
	expected, err := storage.ExchangeToken("new-token")
	assert.NoError(t, err)
	assert.Equal(t, businessID, expected.BusinessID)
	assert.Equal(t, clientID, expected.ClientID)

	// Try to exchange the same token again
	_, err = storage.ExchangeToken("new-token")
	assert.ErrorIs(t, err, ErrTokenAlreadyUsed)

	// Try to exchange a non-existent token
	_, err = storage.ExchangeToken("non-existent-token")
	assert.ErrorIs(t, err, common.ErrNotFound)

	// Test expired token
	expiredToken := "expired-token"
	expiredExpiresAt := time.Now().Add(-1 * time.Hour)
	err = storage.AddUserToken("user2", "client2", expiredToken, expiredExpiresAt)
	assert.NoError(t, err)

	_, err = storage.ExchangeToken(expiredToken)
	assert.ErrorIs(t, err, ErrTokenExpired)

	// Cleanup expired tokens
	err = storage.CleanupExpiredTokens()
	assert.NoError(t, err)

	// Try to get the expired token again
	_, err = storage.ExchangeToken(expiredToken)
	assert.ErrorIs(t, err, common.ErrNotFound)
}
