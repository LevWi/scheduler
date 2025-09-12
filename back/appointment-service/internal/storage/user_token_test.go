package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	common "scheduler/appointment-service/internal"
)

func TestUserToken(t *testing.T) {
	storage := initDB(t)
	defer storage.Close()

	err := CreateTableUserTokens(&storage)
	assert.NoError(t, err)
	userID := common.ID("user1")
	token := "token1"
	expiresAt := time.Now().Add(1 * time.Hour)

	// Add a new token
	err = storage.AddUserToken(userID, token, expiresAt)
	assert.NoError(t, err)

	// Try to add the same token again, should update
	err = storage.AddUserToken(userID, "new-token", expiresAt)
	assert.NoError(t, err)

	// Exchange the token
	exchangedUserID, err := storage.ExchangeToken("new-token")
	assert.NoError(t, err)
	assert.Equal(t, userID, exchangedUserID)

	// Try to exchange the same token again
	_, err = storage.ExchangeToken("new-token")
	assert.ErrorIs(t, err, ErrTokenAlreadyUsed)

	// Try to exchange a non-existent token
	_, err = storage.ExchangeToken("non-existent-token")
	assert.ErrorIs(t, err, common.ErrNotFound)

	// Test expired token
	expiredToken := "expired-token"
	expiredUserID := common.ID("user2")
	expiredExpiresAt := time.Now().Add(-1 * time.Hour)
	err = storage.AddUserToken(expiredUserID, expiredToken, expiredExpiresAt)
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
