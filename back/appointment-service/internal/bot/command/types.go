package command

import (
	"context"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot/i18n/messages"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type ChatID any
type ChoiceID = int
type ChoiceType string

type LocalizationProvider interface {
	LocalizedMap() (messages.MessageMap, error)
	Localizer() (*i18n.Localizer, error)
}

const (
	ChoiceTypeNone  ChoiceType = ""
	ChoiceTypeDays  ChoiceType = "CT_Days"
	ChoiceTypeSlots ChoiceType = "CT_Slots"
)

type ChatContext struct {
	Ctx      context.Context
	ChatID   ChatID
	Customer common.ID
}

// now - is current time corresponding customer local time
// Local time should be from business config or from user config
type Request struct {
	*ChatContext
	Now     time.Time
	Text    string
	Choices struct {
		Type ChoiceType
		IDs  []ChoiceID
	}
}
