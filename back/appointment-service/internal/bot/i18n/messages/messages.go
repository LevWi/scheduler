package messages

import (
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var AvailableSlots = &i18n.Message{
	ID:    "AvailableSlots",
	One:   "Available slot",
	Other: "Available slots",
}

var ThisWeek = &i18n.Message{
	ID:    "ThisWeek",
	Other: "This week",
}

var NextWeek = &i18n.Message{
	ID:    "NextWeek",
	Other: "Next week",
}

var Cancel = &i18n.Message{
	ID:    "Cancel",
	Other: "Cancel",
}

var Done = &i18n.Message{
	ID:    "Done",
	Other: "Done",
}

var Help = &i18n.Message{
	ID:    "Help",
	Other: "Help",
}

var BookSlot = &i18n.Message{
	ID:    "BookSlot",
	Other: "Book a slot",
}

var WrongUserInput = &i18n.Message{
	ID:    "WrongUserInput",
	Other: "Input text is unexpected",
}

var NoSlotsAvailable = &i18n.Message{
	ID:    "NoSlotsFound",
	Other: "No slots available",
}

var InternalErrorOccurred = &i18n.Message{
	ID:    "InternalErrorOccurred",
	Other: "Internal error occurred",
}

var HelpMessage = &i18n.Message{
	ID:    "HelpMessage",
	Other: "TODO: Add big help message here",
}

type MessageConstant = i18n.Message
type MessageMap map[string]*MessageConstant

func (m MessageMap) IdentifyMessage(s string) *MessageConstant {
	return m[s]
}

func LocalizedMessageMap(l *i18n.Localizer, ms ...*i18n.Message) (MessageMap, error) {
	out := make(MessageMap, len(ms))
	for _, m := range ms {
		text, err := l.LocalizeMessage(m)
		if err != nil {
			return nil, err
		}
		text = strings.ToLower(text)
		out[text] = m
	}
	return out, nil
}

func LocalizeMessages(l *i18n.Localizer, ms []*i18n.Message) ([]string, error) {
	localized := make([]string, len(ms))
	for i, m := range ms {
		s, err := l.LocalizeMessage(m)
		if err != nil {
			return nil, err
		}
		localized[i] = s
	}
	return localized, nil
}
