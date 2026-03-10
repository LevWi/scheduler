package google

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchGoogleJWKsUri(t *testing.T) {
	const expected = "https://www.googleapis.com/oauth2/v3/certs"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"jwks_uri":"` + expected + `"}`))
	}))
	defer ts.Close()

	uri, err := fetchJWKsURI(context.Background(), ts.URL, ts.Client())
	require.NoError(t, err)
	assert.Equal(t, expected, uri)
}

func TestFetchGoogleJWKsUri_ErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	_, err := fetchJWKsURI(context.Background(), ts.URL, ts.Client())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code")
}
