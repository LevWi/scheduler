package command

import (
	"scheduler/appointment-service/internal/bot/chat"
	"time"
)

type ChoiceID = int
type ChoiceType string

const (
	ChoiceTypeNone  ChoiceType = ""
	ChoiceTypeDays  ChoiceType = "CT_Days"
	ChoiceTypeSlots ChoiceType = "CT_Slots"
)

// now - is current time corresponding customer local time
// Local time should be from business config or from user config
type Request struct {
	*chat.ChatContext
	Now      time.Time
	Text     string
	Customer string
	Choices  struct {
		Type ChoiceType
		IDs  []ChoiceID
	}
}
