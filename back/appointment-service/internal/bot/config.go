package bot

import "errors"

// TODO For slots GET request is no auth required. So business_id used in request only here.
// Ether we need to set BusinessID in config (but it strongly should corresponding
// bot's client id or errors can occure) or we need change HTTP API
type SchedulerConnection struct {
	URL        string `cfg:"url"`
	BusinessID string `cfg:"business_id"`
	ClientId   string `cfg:"client_id"`
	Token      string `cfg:"token"`
}

func (s *SchedulerConnection) Validate() error {
	if s.URL == "" {
		return errors.New("scheduler url is not set")
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
