package oidc

import (
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuth2CfgCacheProvider struct {
	cache *oauth2.Config
}

func NewOAuth2CfgProviderFromFile(file string, scope ...string) (OAuth2CfgCacheProvider, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return OAuth2CfgCacheProvider{}, fmt.Errorf("read oauth2 cfg file: %w", err)
	}

	if len(scope) == 0 {
		scope = []string{"openid"}
	}
	cfg, err := google.ConfigFromJSON(raw, scope...)
	if err != nil {
		return OAuth2CfgCacheProvider{}, fmt.Errorf("parse oauth2 cfg file: %w", err)
	}

	return OAuth2CfgCacheProvider{cache: cfg}, nil
}

func (p *OAuth2CfgCacheProvider) GetOAuth2Cfg() (*oauth2.Config, error) {
	return p.cache, nil
}
