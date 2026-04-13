package command

import (
	"errors"
	"fmt"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot"
	"scheduler/appointment-service/internal/bot/chat"
	"scheduler/appointment-service/internal/bot/i18n/messages"
	"slices"
	"strings"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type ChatAdapter struct {
	chat.Chat
	settings *bot.UserSettings
}

func NewChatAdapter(chat chat.Chat, settings *bot.UserSettings) *ChatAdapter {
	return &ChatAdapter{
		Chat:     chat,
		settings: settings,
	}
}

const (
	DayMarker  = "bookDayOption_"
	SlotMarker = "bookSlotOption_"
)

var ErrLocalizeMessage = errors.New("localize message error")

func (ca *ChatAdapter) PrintMessage(c *chat.ChatContext, m *i18n.Message) error {
	l := ca.settings.Loc.Localizer()
	localized, err := l.LocalizeMessage(m)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrLocalizeMessage, err)
	}
	return ca.Print(c, localized)
}

func (ca *ChatAdapter) ShowMenuMessages(c *chat.ChatContext, m *i18n.Message, ops []*i18n.Message) error {
	l := ca.settings.Loc.Localizer()
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

func LocationToGMTString(loc *time.Location, t time.Time) string {
	_, offset := t.In(loc).Zone()

	sign := "+"
	if offset < 0 {
		sign = "-"
		offset = -offset
	}

	h := offset / 3600
	m := (offset % 3600) / 60

	return fmt.Sprintf("GMT%s%02d:%02d", sign, h, m)
}

func (ca *ChatAdapter) ShowAsOptions(c *chat.ChatContext, me *i18n.Message, ops []LabeledSlot) error {
	if len(ops) == 0 {
		return fmt.Errorf("%w: slots array too small =%d", common.ErrInvalidArgument, len(ops))
	}

	localized, err := ca.settings.Loc.Localizer().LocalizeMessage(me)
	if err != nil {
		return err
	}

	localizedTimeZone, err := ca.settings.Loc.Localizer().LocalizeMessage(messages.DialogTimeZone)
	if err != nil {
		return err
	}

	location := ca.settings.TimeZone
	dateFormatter := ca.settings.Loc.DF

	formatDateShort := func(tp time.Time) string {
		_, month, day := tp.Date()
		return fmt.Sprintf("%s %02d %s", dateFormatter.WeekDayShort(tp.Weekday()), day, dateFormatter.MonthShort(month))
	}

	m := make(map[time.Time]struct{}, len(ops))
	dstCheckM := make(map[int]struct{}, len(ops))
	for _, v := range ops {
		t := v.Start.In(location)
		key := common.DayBeginning(t)
		m[key] = struct{}{}
		_, offset := t.Zone()
		dstCheckM[offset] = struct{}{}
	}

	var chatOptions []chat.ChatOption
	if len(m) > 1 {
		chatOptions = make([]chat.ChatOption, len(m))
		i := 0
		for tp := range m {
			chatOptions[i].ID = fmt.Sprintf("%s%s", DayMarker, tp.Format(time.DateOnly))
			chatOptions[i].Text = formatDateShort(tp)
			i++
		}
		l := len(DayMarker)
		slices.SortFunc(chatOptions, func(a, b chat.ChatOption) int {
			return strings.Compare(a.ID[l:], b.ID[l:])
		})
	} else {
		chatOptions = make([]chat.ChatOption, len(ops))
		localized = fmt.Sprintf("%s\n%s:", localized, formatDateShort(ops[0].Start.In(location)))
		for i, v := range ops {
			chatOptions[i].ID = fmt.Sprintf("%s%d", SlotMarker, v.ID)
			hour, min, _ := v.Start.In(location).Clock()
			chatOptions[i].Text = fmt.Sprintf("%02d:%02d (%02d %s)", hour, min, int(v.Dur.Minutes()), dateFormatter.MinShort())
		}
	}

	if len(dstCheckM) == 1 {
		localized = fmt.Sprintf("%s\n%s: %s (%s)", localized, localizedTimeZone, location.String(), LocationToGMTString(location, ops[0].Start))
	} else {
		localized = fmt.Sprintf("%s\n%s: %s", localized, localizedTimeZone, location.String())
	}

	return ca.Chat.ShowOptions(c, localized, chatOptions)
}
