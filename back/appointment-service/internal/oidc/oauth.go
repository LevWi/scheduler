package oidc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

type OAuth2CfgProvider interface {
	Get() (*oauth2.Config, error)
}

func OAuth2HTTPRedirectHandler(s sessions.Store, cfg OAuth2CfgProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		oauthConfig, err := cfg.Get()
		if err != nil {
			slog.ErrorContext(r.Context(), "OAuth2HTTPRedirect get config", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		session, err := s.Get(r, "auth-session")
		if err != nil {
			slog.ErrorContext(r.Context(), "OAuth2HTTPRedirect session", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		state := generateState()
		session.Values["oauth_state"] = state

		err = session.Save(r, w)
		if err != nil {
			slog.ErrorContext(r.Context(), "OAuth2HTTPRedirect session save", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		url := oauthConfig.AuthCodeURL(state)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func OAuth2UserBackHandler(s sessions.Store, cfg OAuth2CfgProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := s.Get(r, "auth-session")
		if err != nil {
			slog.ErrorContext(r.Context(), "OAuth2UserBack session", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		{
			//TODO delete oauth_state
			expectedState, ok := session.Values["oauth_state"].(string)
			queryState := r.URL.Query().Get("state")
			if !ok || r.URL.Query().Get("state") != expectedState {
				slog.ErrorContext(r.Context(), "Invalid state parameter", "expected", expectedState, "actual", queryState)
				http.Error(w, "Invalid state parameter", http.StatusBadRequest)
				return
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
			http.Error(w, "Failed to exchange token:", http.StatusInternalServerError)
			return
		}

		//TODO
		// rawIDToken, ok := token.Extra("id_token").(string)
		// if !ok {
		// 	http.Error(w, "No id_token field in oauth2 token", http.StatusInternalServerError)
		// 	return
		// }

		// t, err := jwt.Parse(rawIDToken, func(token *jwt.Token) (any, error) {
		// 	if kid, ok := token.Header["kid"].(string); ok {
		// 		jwk, err := jwkStorage.KeyRead(r.Context(), kid)
		// 		if err != nil {
		// 			return nil, err
		// 		}
		// 		return jwk.Key(), nil
		// 	}

		// 	return nil, errors.New("kid field not found")
		// })
	}
}

func generateState() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Fatal(err)
	}
	return base64.URLEncoding.EncodeToString(b)
}
