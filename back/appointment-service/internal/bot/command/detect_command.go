package command

import (
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot/i18n/messages"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type MessageMapProvider interface {
	Get() (messages.MessageMap, error)
}

type Detector struct {
	mmp MessageMapProvider
}

func (d *Detector) Detect(text string) error {
	m, err := d.mmp.Get()
	if err != nil {
		return err
	}

	c, ok := m[text]
	if !ok {
		return common.ErrNotFound
	}

	switch c {
	case messages.NextWeek:
		//TODO
	case messages.ThisWeek:
		//TODO
	case messages.Cancel:
		//TODO
	}

	return common.ErrInternal
}

func CommandMap(l *i18n.Localizer) (messages.MessageMap, error) {
	return messages.LocalizedMessageMap(l,
		messages.NextWeek,
		messages.ThisWeek,
		messages.Cancel)
}
