package auth

import (
	"errors"
	common "scheduler/appointment-service/internal"
	storage "scheduler/appointment-service/internal/dbase/user_bots"
)

type TokenStorage struct {
	*storage.BotsStorage
}

func (s *TokenStorage) TokenCheck(clientID common.ID, token string) (common.ID, error) {
	businessID, err := s.ValidateBotToken(clientID, token)
	if errors.Is(err, storage.ErrTokenMismatch) {
		err = ErrWrongToken
	}
	return businessID, err
}
