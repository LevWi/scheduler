package oidc

import (
	"context"
	"errors"
	"fmt"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/storage"

	"github.com/MicahParks/jwkset"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

type UserAuthCheckImpl struct {
	s          storage.Storage
	jwkStorage jwkset.Storage
}

func (c *UserAuthCheckImpl) AuthCheck(ctx context.Context, token *oauth2.Token) (id common.ID, isNew bool, err error) {
	if token == nil {
		err = common.ErrInvalidArgument
		return
	}

	jwtToken, err := c.verifyOIDCToken(ctx, token)
	if err != nil {
		return
	}

	iss, err := claimsCheck(jwtToken.Claims.GetIssuer())
	if err != nil {
		return
	}
	iss = issNormalize(iss)

	sub, err := claimsCheck(jwtToken.Claims.GetSubject())
	if err != nil {
		return
	}

	authData := storage.OIDCData{
		Provider: iss,
		Subject:  sub,
	}

	dbUid, err := c.s.OIDCUserAuth(authData)
	if errors.Is(err, common.ErrNotFound) {
		isNew = true
		dbUid, err = c.s.OIDCCreateUser(storage.GenerateUsername(), authData)
	}
	if err != nil {
		return
	}

	id = common.ID(dbUid)
	return
}

// jwt library may return empty string without errors
func claimsCheck(v string, err error) (string, error) {
	if v == "" && err == nil {
		err = common.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("jwt token check fail: %w", err)
	}
	return v, nil
}

func (c *UserAuthCheckImpl) verifyOIDCToken(ctx context.Context, token *oauth2.Token) (*jwt.Token, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("%w: id_token not found", common.ErrInvalidArgument)
	}

	//Don't see how to pass WithValidMethods with correct values when we not sure.
	//So "alg" checking done in keyFunc
	jwtToken, err := jwt.Parse(rawIDToken, func(t *jwt.Token) (any, error) {
		if kid, ok := t.Header["kid"].(string); ok {
			jwk, err := c.jwkStorage.KeyRead(ctx, kid)
			if err != nil {
				return nil, err
			}

			if t.Method.Alg() != string(jwk.Marshal().ALG) {
				return nil, fmt.Errorf("%w: 'alg' %s is not expected for kid == %s", common.ErrInvalidArgument, t.Method.Alg(), kid)
			}

			return jwk.Key(), nil
		}

		return nil, fmt.Errorf("%w: kid field not found", common.ErrInvalidArgument)
	})

	if err != nil {
		return nil, err
	}

	return jwtToken, nil

}

func issNormalize(v string) string {
	if v == "https://accounts.google.com" {
		v = "accounts.google.com"
	}
	return v
}
