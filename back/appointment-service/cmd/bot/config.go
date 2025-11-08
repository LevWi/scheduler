package main

import (
	"errors"
	"log/slog"
	"scheduler/appointment-service/internal/bot"
	"scheduler/appointment-service/internal/config"
)

type BotConfig struct {
	BotAPIConnection string                  `cfg:"bot_api_connection"`
	LogLevel         slog.Level              `cfg:"log_level"`
	SchedulerAPI     bot.SchedulerConnection `cfg:"scheduler"`
	//TODO business_id?
}

func (c *BotConfig) Validate() error {
	if c.BotAPIConnection == "" {
		return errors.New("bot_api_connection is not set")
	}
	if c.SchedulerAPI.ClientId == "" {
		return errors.New("scheduler_api client_id is not set")
	}
	if c.SchedulerAPI.Token == "" {
		return errors.New("scheduler_api token is not set")
	}
	return nil
}

func LoadBotConfig() (*BotConfig, error) {
	var cfg BotConfig
	err := config.LoadAndCheckConfig("cfg", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
