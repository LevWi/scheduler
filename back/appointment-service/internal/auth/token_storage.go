package auth

import (
	"errors"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/storage"
)

type TokenStorage storage.Storage

func (s *TokenStorage) TokenCheck(clientID common.ID, token string) (common.ID, error) {
	businessID, err := (*storage.Storage)(s).ValidateBotToken(clientID, token)
	if errors.Is(err, storage.ErrTokenMismatch) {
		err = ErrWrongToken
	}
	return businessID, err
}
