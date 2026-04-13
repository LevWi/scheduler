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
	Other: "book a slot",
}

var Appointments = &i18n.Message{
	ID:    "Appointments",
	Other: "appointments",
}

var Settings = &i18n.Message{
	ID:    "Settings",
	Other: "settings",
}

var SetLanguage = &i18n.Message{
	ID:    "SetLanguage",
	Other: "set language",
}

var SetTimeZone = &i18n.Message{
	ID:    "SetTimeZone",
	Other: "set time zone",
}

var SelectLanguageMessage = &i18n.Message{
	ID:    "SelectLanguageMessage",
	Other: "Please select language",
}

var SelectSettingsOptionMessage = &i18n.Message{
	ID:    "SelectSettingsOptionMessage",
	Other: "Please select settings option",
}

var EnterTimeZoneMessage = &i18n.Message{
	ID:    "EnterTimeZoneMessage",
	Other: "Please type time zone (IANA), for example: Europe/Berlin",
}

var InvalidTimeZone = &i18n.Message{
	ID:    "InvalidTimeZone",
	Other: "Invalid time zone",
}

var LanguageUpdated = &i18n.Message{
	ID:    "LanguageUpdated",
	Other: "Language updated",
}

var TimeZoneUpdated = &i18n.Message{
	ID:    "TimeZoneUpdated",
	Other: "Time zone updated",
}

var DialogTimeZone = &i18n.Message{
	ID:    "DialogTimeZone",
	Other: "Time zone",
}

var AppointmentsListHeader = &i18n.Message{
	ID:    "AppointmentsListHeader",
	Other: "Your upcoming appointments:",
}

var NoUpcomingAppointments = &i18n.Message{
	ID:    "NoUpcomingAppointments",
	Other: "No upcoming appointments",
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
	ID: "HelpMessage",
	Other: `Available commands:
"help" - Show this help message
"book a slot"   - Book a slot for an appointment
"appointments" - Show your upcoming appointments
"settings" - Open language and time zone settings
"cancel" - Cancel current operation or booking
Use special button or type it
`,
}

// TODO remove it. Only for skeleton
var SelectRequestMessage = &i18n.Message{
	ID:    "SelectRequestMessage",
	Other: "Please select an option",
}

var CommandRequestMessage = &i18n.Message{
	ID: "CommandRequestMessage",
	Other: `Please type or select an command
	(try "help" for more info)`,
}

type MessageConstant = i18n.Message
type MessageMap map[string]*MessageConstant

func (m MessageMap) IdentifyMessage(s string) *MessageConstant {
	return m[strings.ToLower(s)]
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
