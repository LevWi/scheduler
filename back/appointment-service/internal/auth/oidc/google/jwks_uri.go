package google

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const GoogleOpenIDDoc = "https://accounts.google.com/.well-known/openid-configuration"

func FetchGoogleJWKsUri(ctx context.Context) (string, error) {
	return fetchJWKsURI(ctx, http.DefaultClient, GoogleOpenIDDoc)
}

func fetchJWKsURI(ctx context.Context, httpClient *http.Client, openIDDocURL string) (string, error) {
	var cfg struct {
		JwksURI string `json:"jwks_uri"`
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, openIDDocURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("[FetchGoogleJWKsUri] http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("[FetchGoogleJWKsUri] unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		return "", fmt.Errorf("[FetchGoogleJWKsUri] http body decoding: %w", err)
	}

	return cfg.JwksURI, nil
}
