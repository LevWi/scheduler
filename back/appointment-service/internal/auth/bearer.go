package auth

import (
	"errors"
	"fmt"
	"net/http"
	common "scheduler/appointment-service/internal"
)

var ErrWrongToken = errors.New("wrong token")

type TokenChecker interface {
	// Expected common.ErrNotFound ErrWrongToken common.ErrInvalidArgument
	TokenCheck(clientID common.ID, token string) (common.ID, error)
}

type BearerAuth struct {
	TC TokenChecker
}

func getToken(r *http.Request) (string, error) {
	return getAuthorizationTokenByScheme(r, "Bearer")
}

func (ba *BearerAuth) Authorization(r *http.Request) (common.ID, error) {
	token, err := getToken(r)
	if err != nil {
		return "", err
	}

	clientID, err := getClientIDFromHeader(r)
	if err != nil {
		return "", err
	}

	businessID, err := ba.TC.TokenCheck(clientID, token)
	if err != nil {
		err = errors.Join(common.ErrUnauthorized, err)
		return "", fmt.Errorf("[Bearer Authorization]: %w", err)
	}

	return common.ID(businessID), nil
}
