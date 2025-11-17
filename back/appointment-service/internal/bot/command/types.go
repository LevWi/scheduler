package command

import (
	"context"
	"scheduler/appointment-service/internal/bot/i18n/messages"
	"time"
)

type ChatID any
type ChoiceIndex = int

type MessageMapProvider interface {
	Get() (messages.MessageMap, error)
}

// now - is current time corresponding customer local time
// Local time should be from business config or from user config
type Request struct {
	Ctx     context.Context
	Now     time.Time
	Text    string
	ChatID  ChatID
	Choices []ChoiceIndex
}
