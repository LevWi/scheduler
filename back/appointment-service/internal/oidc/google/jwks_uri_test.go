package google

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchGoogleJWKsUri(t *testing.T) {
	const expected = "https://www.googleapis.com/oauth2/v3/certs"
	uri, err := FetchGoogleJWKsUri(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expected, uri)
}
