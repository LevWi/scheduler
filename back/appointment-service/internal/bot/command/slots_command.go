package command

import (
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot/i18n/messages"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type SlotsResultChatOutput interface {
	Print(r *Request, m WeekDaysSlotsMap) error
}

type WeekDaysSlotsMap map[time.Weekday]common.Intervals

func SlotsToWeekMap(in common.Intervals) WeekDaysSlotsMap {
	m := make(WeekDaysSlotsMap)
	for _, interval := range in {
		wd := interval.Start.Weekday()
		m[wd] = append(m[wd], interval)
	}
	return m
}

// TODO move it
func CommandMap(l *i18n.Localizer) (messages.MessageMap, error) {
	return messages.LocalizedMessageMap(l,
		messages.NextWeek,
		messages.ThisWeek,
		messages.Cancel,
		messages.Done)
}

type SlotsCommandStMachine struct {
	result      WeekDaysSlotsMap
	mmp         MessageMapProvider
	chatPrinter SlotsResultChatOutput
	commands    struct {
		WeekSlots *WeekSlots
	}
}

func (sm *SlotsCommandStMachine) Process(r *Request) error {
	m, err := sm.mmp.Get()
	if err != nil {
		return err
	}

	c, ok := m[r.Text]
	if !ok {
		return common.ErrNotFound
	}

	if sm.result == nil {
		var slots common.Intervals
		switch c {
		case messages.NextWeek:
			slots, err = sm.commands.WeekSlots.NextWeek(r.Ctx, r.Now)
		case messages.ThisWeek:
			slots, err = sm.commands.WeekSlots.ThisWeek(r.Ctx, r.Now)
		case messages.Cancel:
			//TODO
		default:
			err = common.ErrInvalidArgument
		}

		if err != nil {
			return nil
		}

		sm.result = SlotsToWeekMap(slots)
		err = sm.chatPrinter.Print(r, sm.result)
		if err != nil {
			return err
		}
	} else {

	}
}
