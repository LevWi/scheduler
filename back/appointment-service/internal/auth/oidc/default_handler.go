package oidc

import (
	"context"
	"net/http"

	"github.com/MicahParks/jwkset"

	"scheduler/appointment-service/internal/auth/oidc/google"
	"scheduler/appointment-service/internal/dbase/auth"
)

func NewOIDCAuthCheckDefault(ctx context.Context, dbase *auth.AuthStorage) (OIDCAuthCheck, error) {
	//TODO make it with periodically update?
	//TODO make timeout
	url, err := google.FetchGoogleJWKsUri(ctx, http.DefaultClient)
	if err != nil {
		return nil, err
	}

	j, err := jwkset.NewDefaultHTTPClient([]string{url})
	if err != nil {
		return nil, err
	}

	return &OIDCAuthCheckImpl{
		s:          dbase,
		jwkStorage: j,
	}, nil

}
