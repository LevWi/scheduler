package oidc

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	common "scheduler/appointment-service/internal"

	"github.com/gorilla/sessions"
)

type OAuth2SessionsValidator struct {
	S sessions.Store //session need to be short live
}

func generateState() string {
	var b [16]byte
	rand.Read(b[:]) //Error not expected
	return base64.URLEncoding.EncodeToString(b[:])
}

// TODO accept only one of request with same cookie ?
func (v *OAuth2SessionsValidator) PrepareState(w http.ResponseWriter, r *http.Request) (string, error) {
	session, err := v.S.Get(r, "auth-session")
	if err != nil {
		return "", fmt.Errorf("session creation fail: %w", err)
	}

	state := generateState()
	session.Values["oauth_state"] = state

	err = session.Save(r, w)
	if err != nil {
		return "", fmt.Errorf("session saving fail: %w", err)
	}
	return state, nil
}

func (v *OAuth2SessionsValidator) ValidateCallback(w http.ResponseWriter, r *http.Request) error {
	session, err := v.S.Get(r, "auth-session")
	if err != nil {
		return fmt.Errorf("%w: session creation fail: %w", common.ErrInternal, err)
	}

	//TODO delete oauth_state
	expectedState, ok := session.Values["oauth_state"].(string)
	queryState := r.URL.Query().Get("state")
	if !ok || r.URL.Query().Get("state") != expectedState {
		return fmt.Errorf("%w : state parameter expected %s, but got %s", common.ErrInvalidArgument, expectedState, queryState)
	}
	return nil
}
