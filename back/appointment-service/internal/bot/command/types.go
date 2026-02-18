package command

import (
	"scheduler/appointment-service/internal/bot/chat"
	"time"
)

type ChoiceID = string
type Customer string

// Time - time the message was sent
// Local user time should be from business config or from user
type Request struct {
	*chat.ChatContext
	Time     time.Time
	Text     string
	Customer Customer
	Choices  []ChoiceID
}
