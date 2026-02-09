package command

import (
	"errors"
	"fmt"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot/chat"
	"scheduler/appointment-service/internal/bot/i18n/messages"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

//TODO add here localizer?

type ChatAdapter struct {
	chat.Chat
	Loc *messages.Localization
}

func NewChatAdapter(chat chat.Chat, loc *messages.Localization) *ChatAdapter {
	return &ChatAdapter{
		Chat: chat,
		Loc:  loc,
	}
}

const (
	DateMarker = "bookDateOption_"
	SlotMarker = "bookSlotOption_"
)

var ErrLocalizeMessage = errors.New("localize message error")

func (ca *ChatAdapter) PrintMessage(c *chat.ChatContext, m *i18n.Message) error {
	l := ca.Loc.Localizer()
	localized, err := l.LocalizeMessage(m)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrLocalizeMessage, err)
	}
	return ca.Print(c, localized)
}

func (ca *ChatAdapter) ShowMenuMessages(c *chat.ChatContext, m *i18n.Message, ops []*i18n.Message) error {
	l := ca.Loc.Localizer()
	localized, err := messages.LocalizeMessages(l, ops)
	if err != nil {
		return err
	}

	mess, err := l.LocalizeMessage(m)
	if err != nil {
		return err
	}
	return ca.ShowMenu(c, mess, localized)
}

func (ca *ChatAdapter) ShowAsOptions(c *chat.ChatContext, me *i18n.Message, ops []LabeledSlot) error {
	if len(ops) < 2 {
		return fmt.Errorf("%w: slots array too small =%d", common.ErrInvalidArgument, len(ops))
	}

	localized, err := ca.Loc.Localizer().LocalizeMessage(me)
	if err != nil {
		return err
	}

	loc := ops[0].Start.Location()
	m := make(map[string]struct{}, len(ops))
	for _, v := range ops {
		key := v.Start.In(loc).Format(time.RFC822)
		m[key] = struct{}{}
	}

	var chatOptions []chat.ChatOption
	if len(m) > 1 {
		chatOptions = make([]chat.ChatOption, len(m))
		i := 0
		for k := range m {
			chatOptions[i].ID = DateMarker + k
			chatOptions[i].Text = k //TODO need Localization here
			i++
		}
	} else {
		chatOptions = make([]chat.ChatOption, len(ops))
		for i, v := range ops {
			chatOptions[i].ID = fmt.Sprintf("%s%d", SlotMarker, v.ID)
			chatOptions[i].Text = v.Start.Format(time.DateTime) //TODO need Localization here
		}
	}

	return ca.Chat.ShowOptions(c, localized, chatOptions)
}

// // TODO move to struct? to bot adapter?
// func TimeShortFormat(df DateFormatter) TimeFormatFunc {
// 	return func(t time.Time) string {
// 		return fmt.Sprintf(
// 			"%d %s (%s) %02d:%02d",
// 			t.Day(),
// 			df.MonthShort(t.Month()),
// 			df.WeekDayShort(t.Weekday()),
// 			t.Hour(),
// 			t.Minute(),
// 		)
// 	}
// }

// type TimeFormatFunc func(time.Time) string
