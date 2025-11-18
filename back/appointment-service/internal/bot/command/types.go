package command

import (
	"context"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot/i18n/messages"
	"time"
)

type ChatID any
type ChoiceID = int
type ChoiceType string

type MessageMapProvider interface {
	Get() (messages.MessageMap, error)
}

const (
	ChoiceTypeNone  = ""
	ChoiceTypeDays  = "CT_Days"
	ChoiceTypeSlots = "CT_Slots"
)

// now - is current time corresponding customer local time
// Local time should be from business config or from user config
type Request struct {
	Ctx      context.Context
	Now      time.Time
	Customer common.ID
	ChatID   ChatID
	Text     string
	Choices  struct {
		Type ChoiceType
		IDs  []ChoiceID
	}
}
