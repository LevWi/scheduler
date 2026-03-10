package google

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchJWKsURI(t *testing.T) {
	t.Run("fetches jwks uri from openid document", func(t *testing.T) {
		const expected = "https://www.googleapis.com/oauth2/v3/certs"

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			_, err := fmt.Fprintf(w, `{"jwks_uri":"%s"}`, expected)
			require.NoError(t, err)
		}))
		defer s.Close()

		actual, err := fetchJWKsURI(context.Background(), s.Client(), s.URL)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("returns explicit error on non-200 response", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "upstream error", http.StatusInternalServerError)
		}))
		defer s.Close()

		_, err := fetchJWKsURI(context.Background(), s.Client(), s.URL)
		require.Error(t, err)
		assert.EqualError(t, err, "[FetchGoogleJWKsUri] unexpected status code: 500")
	})
}
