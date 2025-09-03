package oidc

import (
	"context"

	"github.com/MicahParks/jwkset"

	"scheduler/appointment-service/internal/auth/oidc/google"
	db "scheduler/appointment-service/internal/storage"
)

func NewOIDCAuthCheckDefault(ctx context.Context, dbase *db.Storage) (OIDCAuthCheck, error) {
	//TODO make it with periodically update?
	url, err := google.FetchGoogleJWKsUri(ctx)
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
