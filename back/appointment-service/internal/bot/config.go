package bot

import "errors"

type SchedulerConnection struct {
	URL      string `cfg:"url"`
	ClientId string `cfg:"client_id"`
	Token    string `cfg:"token"`
}

func (s *SchedulerConnection) Validate() error {
	if s.URL == "" {
		return errors.New("scheduler url is not set")
	}
	if s.ClientId == "" {
		return errors.New("scheduler client_id is not set")
	}
	if s.Token == "" {
		return errors.New("scheduler token is not set")
	}
	return nil
}
