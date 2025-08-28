package oidc

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	common "scheduler/appointment-service/internal"
	"time"

	"golang.org/x/oauth2"
)

type OAuth2CfgProvider interface {
	GetOAuth2Cfg() (*oauth2.Config, error)
}

type OIDCAuthCheck interface {
	AuthCheck(ctx context.Context, token *oauth2.Token) (id common.ID, isNew bool, err error)
}

type OAuth2ValidateState interface {
	PrepareState(w http.ResponseWriter, r *http.Request) (string, error)
	ValidateCallback(w http.ResponseWriter, r *http.Request) error
}

type SaveUserCookie interface {
	Authenticate(id common.ID, w http.ResponseWriter, r *http.Request) error
}

type UserSignIn struct {
	OAuth2ValidateState
	OAuth2CfgProvider
	OIDCAuthCheck
	SaveUserCookie
}

func OAuth2HTTPRedirectHandler(s *UserSignIn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		oauthConfig, err := s.GetOAuth2Cfg()
		if err != nil {
			slog.ErrorContext(r.Context(), "OAuth2HTTPRedirect get config", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		state, err := s.PrepareState(w, r)
		if err != nil {
			slog.ErrorContext(r.Context(), "OAuth2HTTPRedirect", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, oauthConfig.AuthCodeURL(state), http.StatusTemporaryRedirect)
	}
}

// TODO How to handle new user? see AuthCheck()
func OAuth2HTTPUserBackHandler(s *UserSignIn, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := s.ValidateCallback(w, r)
		if err != nil {
			slog.WarnContext(r.Context(), "OAuth2UserBack", "err", err.Error())
			if errors.Is(err, common.ErrInvalidArgument) {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}

		oauthConfig, err := s.GetOAuth2Cfg()
		if err != nil {
			slog.ErrorContext(r.Context(), "OAuth2UserBack", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		code := r.URL.Query().Get("code")
		ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
		defer cancel()
		token, err := oauthConfig.Exchange(ctx, code)
		if err != nil {
			slog.WarnContext(r.Context(), "OAuth2UserBack", "err", err.Error())
			http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
			return
		}

		uid, _, err := s.AuthCheck(r.Context(), token)
		if err != nil {
			slog.WarnContext(r.Context(), "OAuth2UserBack", "err", err.Error())
			http.Error(w, "Failed sing in", http.StatusInternalServerError)
			return
		}

		err = s.Authenticate(uid, w, r)
		if err != nil {
			slog.WarnContext(r.Context(), "OAuth2UserBack", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if next != nil {
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}
