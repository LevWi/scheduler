package oidc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	common "scheduler/appointment-service/internal"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

type OAuth2CfgProvider interface {
	Get() (*oauth2.Config, error)
}

type OAuth2SignIn interface {
	Do(token *oauth2.Token) (common.ID, error)
}

type OAuth2ValidateState struct {
	S sessions.Store
}

func (v *OAuth2ValidateState) PrepareState(w http.ResponseWriter, r *http.Request) (string, error) {
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

func (v *OAuth2ValidateState) Validate(w http.ResponseWriter, r *http.Request) error {
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

func OAuth2HTTPRedirectHandler(v *OAuth2ValidateState, cfg OAuth2CfgProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		oauthConfig, err := cfg.Get()
		if err != nil {
			slog.ErrorContext(r.Context(), "OAuth2HTTPRedirect get config", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		state, err := v.PrepareState(w, r)
		if err != nil {
			slog.ErrorContext(r.Context(), "OAuth2HTTPRedirect", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, oauthConfig.AuthCodeURL(state), http.StatusTemporaryRedirect)
	}
}

func OAuth2HTTPUserBackHandler(v *OAuth2ValidateState, cfg OAuth2CfgProvider, singIn OAuth2SignIn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := v.Validate(w, r)
		if err != nil {
			slog.WarnContext(r.Context(), "OAuth2UserBack", "err", err.Error())
			if errors.Is(err, common.ErrInvalidArgument) {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}

		oauthConfig, err := cfg.Get()
		if err != nil {
			slog.ErrorContext(r.Context(), "OAuth2UserBack get config", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		code := r.URL.Query().Get("code")
		ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
		defer cancel()
		token, err := oauthConfig.Exchange(ctx, code)
		if err != nil {
			slog.WarnContext(r.Context(), "OAuth2UserBack get config", "err", err.Error())
			http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
			return
		}

		uid, err := singIn.Do(token)
		if err != nil {
			slog.WarnContext(r.Context(), "OAuth2UserBack", "err", err.Error())
			http.Error(w, "Failed sing in", http.StatusInternalServerError)
			return
		}

		//TODO write uid to
	}
}

func generateState() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err) //Not expected
	}
	return base64.URLEncoding.EncodeToString(b)
}
