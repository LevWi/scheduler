package common

import (
	"time"

	"github.com/google/uuid"
)

type UserTokenStorage interface {
	AddUserToken(userID ID, token string, expiresAt time.Time) error
	ExchangeToken(token string) (ID, error)
}

type UserTokenPool struct {
	storage       UserTokenStorage
	tokenLifeTime time.Duration
}

func (us *UserTokenPool) NewToken(userID ID) (string, error) {
	token := uuid.New().String()
	expiresAt := time.Now().Add(us.tokenLifeTime)
	return token, us.storage.AddUserToken(userID, token, expiresAt)
}

func (us *UserTokenPool) Exchange(token string) (ID, error) {
	return us.storage.ExchangeToken(token)
}

func NewUserTokenPool(storage UserTokenStorage, tokenLifeTime time.Duration) *UserTokenPool {
	return &UserTokenPool{
		storage:       storage,
		tokenLifeTime: tokenLifeTime,
	}
}
