package main

import (
	"errors"
	"fmt"
	"scheduler/appointment-service/internal/config"

	"github.com/knadh/koanf/v2"
)

// TODO add log level
type ServiceConfig struct {
	SessionsKey string `cfg:"sessions_key"`
	Addr        string `cfg:"addr"`
	DB          struct {
		Driver     string `cfg:"driver"`
		Connection string `cfg:"connection"`
	} `cfg:"db"`
	Auth struct {
		OAuthGoogleConfig string `cfg:"oauth_google_config"`
	} `cfg:"auth"`
	FrontPath string `cfg:"front_path"`
}

func (c *ServiceConfig) Validate() error {
	if c.SessionsKey == "" {
		return errors.New("sessions_key is required")
	}
	if c.Addr == "" {
		return errors.New("addr is required")
	}
	if c.Auth.OAuthGoogleConfig == "" {
		return errors.New("google_config is required")
	}
	if c.DB.Driver == "" {
		return errors.New("db.driver is required")
	}
	if c.DB.Connection == "" {
		return errors.New("db.connection is required")
	}
	if c.FrontPath == "" {
		return errors.New("front_path is required")
	}
	return nil
}

func configErr(err error) error {
	return fmt.Errorf("config err: %w", err)
}

func LoadServiceConfig() (*ServiceConfig, error) {
	k, err := config.LoadConfig()
	if err != nil {
		return nil, configErr(err)
	}

	var config ServiceConfig
	err = k.UnmarshalWithConf("", &config, koanf.UnmarshalConf{Tag: "cfg", FlatPaths: false})
	if err != nil {
		return nil, configErr(err)
	}

	if err = config.Validate(); err != nil {
		return nil, configErr(err)
	}
	return &config, nil
}
