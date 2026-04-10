package bot

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

// TODO For slots GET request is no auth required. So business_id used in request only here.
// Ether we need to set BusinessID in config (but it strongly should corresponding
// bot's client id or error can occurs) or we need change HTTP API
// TODO add default business time zone
type SchedulerConnection struct {
	URL        string `cfg:"url"`
	BusinessID string `cfg:"business_id"`
	ClientId   string `cfg:"client_id"`
	Token      string `cfg:"token"`
}

type DefaultUserSettings struct {
	Language string `cfg:"language"`
	TimeZone string `cfg:"time_zone"`
}

func (s *SchedulerConnection) Validate() error {
	if s.URL == "" {
		return errors.New("scheduler url is not set")
	} else if _, err := url.Parse(s.URL); err != nil {
		return fmt.Errorf("config check error: %w", err)
	}
	if s.BusinessID == "" {
		return errors.New("scheduler business_id is not set")
	}
	if s.ClientId == "" {
		return errors.New("scheduler client_id is not set")
	}
	if s.Token == "" {
		return errors.New("scheduler token is not set")
	}
	return nil
}

func (s *DefaultUserSettings) Validate() error {
	if s.Language == "" {
		return errors.New("default language is not set")
	}
	if s.TimeZone == "" {
		return errors.New("default time_zone is not set")
	}
	_, err := time.LoadLocation(s.TimeZone)
	if err != nil {
		return fmt.Errorf("default time_zone check error: %w", err)
	}
	return nil
}
