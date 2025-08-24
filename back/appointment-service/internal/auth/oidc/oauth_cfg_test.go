package oidc

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

const test_value = `{"web":{"client_id":"1123-blablabla.apps.googleusercontent.com","project_id":"some_project","auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://oauth2.googleapis.com/token","auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs","client_secret":"some_secret_bla_bla","redirect_uris":["http://localhost:8080/callback"]}}`

func TestNewOAuth2CfgProviderFromFile(t *testing.T) {
	expected := &oauth2.Config{
		ClientID:     "1123-blablabla.apps.googleusercontent.com",
		ClientSecret: "some_secret_bla_bla",
		Scopes:       []string{"openid"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
		RedirectURL: "http://localhost:8080/callback",
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "input.json")

	err := os.WriteFile(filePath, []byte(test_value), 0644)
	require.NoError(t, err)

	actual, err := NewOAuth2CfgProviderFromFile(filePath, "openid")
	require.NoError(t, err)

	assert.Equal(t, expected, actual.cache)
}
