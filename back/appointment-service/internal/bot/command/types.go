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

// Time - time the message was sent
// Local user time should be from business config or from user
type Request struct {
	*chat.ChatContext
	Time     time.Time
	Text     string
	Customer string
	Choices  struct {
		Type ChoiceType
		IDs  []ChoiceID
	}
}
