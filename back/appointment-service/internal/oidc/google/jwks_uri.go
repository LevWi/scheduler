package google

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const GoogleOpenIDDoc = "https://accounts.google.com/.well-known/openid-configuration"

func FetchGoogleJWKsUri(ctx context.Context) (string, error) {
	var cfg struct {
		JwksURI string `json:"jwks_uri"`
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, GoogleOpenIDDoc, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request err: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		return "", fmt.Errorf("http body decoding err: %w", err)
	}

	return cfg.JwksURI, nil
}
