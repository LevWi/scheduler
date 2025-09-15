package oidc

import (
	"fmt"
	"net/http"
	common "scheduler/appointment-service/internal"

	"github.com/gorilla/sessions"
)

const oauthSessionName = "auth-session"
const oauthStateKey = "oauth_state"

type OAuth2SessionsValidator struct {
	sessions.Store
}

// TODO accept only one of request with same cookie ?
// TODO session need to be short live
func (v *OAuth2SessionsValidator) PrepareState(w http.ResponseWriter, r *http.Request) (string, error) {
	session, err := v.Get(r, oauthSessionName)
	if err != nil {
		return "", fmt.Errorf("session creation fail: %w", err)
	}

	const stateLength = 16
	state := common.GenerateSecretKey(stateLength)
	//TODO make session short lived?
	session.Values[oauthStateKey] = state

	err = session.Save(r, w)
	if err != nil {
		return "", fmt.Errorf("session saving fail: %w", err)
	}
	return state, nil
}

func (v *OAuth2SessionsValidator) ValidateCallback(w http.ResponseWriter, r *http.Request) error {
	session, err := v.Get(r, oauthSessionName)
	if err != nil {
		return fmt.Errorf("%w: session creation fail: %w", common.ErrInternal, err)
	}

	expectedState, ok := session.Values[oauthStateKey].(string)
	queryState := r.URL.Query().Get("state")

	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		return fmt.Errorf("deleting auth session: %w", err)
	}

	if !ok || queryState != expectedState {
		return fmt.Errorf("%w : state parameter expected %s, but got %s", common.ErrInvalidArgument, expectedState, queryState)
	}
	return nil
}
