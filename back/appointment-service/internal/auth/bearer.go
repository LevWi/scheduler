package auth

import (
	"errors"
	"fmt"
	"net/http"
	common "scheduler/appointment-service/internal"
	"strings"
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
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing 'Authorization' header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid Authorization header format")
	}

	return parts[1], nil
}

func (ba *BearerAuth) Authorization(r *http.Request) (common.ID, error) {
	token, err := getToken(r)
	if err != nil {
		return "", fmt.Errorf("[Bearer Authorization]: %w", err)
	}

	clientID := r.Header.Get("X-Client-ID")
	if clientID == "" {
		return "", errors.New("missing 'X-Client-ID' header")
	}

	businessID, err := ba.TC.TokenCheck(clientID, token)
	if err != nil {
		err = errors.Join(common.ErrUnauthorized, err)
		return "", fmt.Errorf("[Bearer Authorization]: %w", err)
	}

	return common.ID(businessID), nil
}
