package messages

import "github.com/nicksnyder/go-i18n/v2/i18n"

var AvailableSlots = &i18n.Message{
	ID:    "AvailableSlots",
	One:   "Available slot",
	Other: "Available slots",
}

var ThisWeek = &i18n.Message{
	ID:  "ThisWeek",
	One: "This week",
}

var NextWeek = &i18n.Message{
	ID:  "NextWeek",
	One: "Next week",
}

var Cancel = &i18n.Message{
	ID:  "Cancel",
	One: "Cancel",
}

type MessageMap map[string]*i18n.Message

func LocalizedMessageMap(l *i18n.Localizer, ms ...*i18n.Message) (MessageMap, error) {
	out := make(MessageMap, len(ms))
	for _, m := range ms {
		text, err := l.LocalizeMessage(m)
		if err != nil {
			return nil, err
		}
		out[text] = m
	}
	return out, nil
}
