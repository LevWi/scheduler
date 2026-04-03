package auth

import (
	"errors"
	"fmt"
	"net/http"
	common "scheduler/appointment-service/internal"
	"strings"
)

func getAuthorizationTokenByScheme(r *http.Request, expectedScheme string) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("%w: 'Authorization' header", common.ErrNotFound)
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], expectedScheme) {
		return "", errors.New("invalid Authorization header format")
	}

	return parts[1], nil
}

func ClientIDFromHeader(r *http.Request) (common.ID, error) {
	clientID := r.Header.Get("X-Client-ID")
	if clientID == "" {
		return "", fmt.Errorf("%w: 'X-Client-ID' header", common.ErrNotFound)
	}
	return common.ID(clientID), nil
}
