package main

import (
	"errors"
	"log/slog"
	"scheduler/appointment-service/internal/config"
)

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
	LogLevel  slog.Level `cfg:"log_level"`
	FrontPath string     `cfg:"front_path"`
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

func LoadServiceConfig() (*ServiceConfig, error) {
	var cfg ServiceConfig
	err := config.LoadAndCheckConfig("cfg", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
