package google

import (
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func ReadOAuth2Config(file string, scope ...string) (*oauth2.Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	config, err := google.ConfigFromJSON(data, scope...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OAuth config: %w", err)
	}

	return config, nil
}
