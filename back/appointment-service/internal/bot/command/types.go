package command

import (
	"scheduler/appointment-service/internal/bot/chat"
	"scheduler/appointment-service/internal/bot/i18n/messages"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
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

type Localization struct {
	Map messages.MessageMap
	*i18n.Localizer
	DateFormatter
}

type DateFormatter interface {
	MonthShort(m time.Month) string
	WeekDayShort(d time.Weekday) string
}
