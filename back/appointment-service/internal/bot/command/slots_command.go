package command

import (
	"fmt"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot/i18n/messages"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type ChatOutput interface {
	PrintSlots(r *Request, m []LabeledSlot) error
	ConfirmAppointment(r *Request, m []LabeledSlot) error
	ConfirmCancel(r *Request) error
}

type LabeledSlot struct {
	ID int
	common.Interval
}

func ToLabeledSlot(in common.Intervals) []LabeledSlot {
	ls := make([]LabeledSlot, len(in))
	for i, interval := range in {
		ls[i] = LabeledSlot{ID: i, Interval: interval}
	}
	return ls
}

// type WeekDaysSlotsMap map[time.Weekday][]LabeledSlot

// func SlotsToWeekMap(in common.Intervals) WeekDaysSlotsMap {
// 	m := make(WeekDaysSlotsMap)
// 	for i, interval := range in {
// 		wd := interval.Start.Weekday()
// 		m[wd] = append(m[wd], LabeledSlot{ID: i, Interval: interval})
// 	}
// 	return m
// }

// TODO move it
func CommandMap(l *i18n.Localizer) (messages.MessageMap, error) {
	return messages.LocalizedMessageMap(l,
		messages.NextWeek,
		messages.ThisWeek,
		messages.Cancel,
		messages.Done)
}

type SlotsCommandStMachine struct {
	availableSlots []LabeledSlot

	mmp      MessageMapProvider
	chat     ChatOutput
	commands struct {
		WeekSlots   *WeekSlots
		Appointment *Appointment
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

	//TODO Need handle "Cancel" first
	if sm.availableSlots == nil {
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
			return err
		}
		sm.availableSlots = ToLabeledSlot(slots)
		return sm.chat.PrintSlots(r, sm.availableSlots)
	} else {
		if r.Text != "" {
			return fmt.Errorf("%w: input text should be empty", common.ErrInvalidArgument)
		}

		if len(r.Choices.IDs) == 0 {
			return fmt.Errorf("%w: expected user choices", common.ErrInvalidArgument)
		}

		switch r.Choices.Type {
		case ChoiceTypeDays:
			//For simplicity only one day now
			if len(r.Choices.IDs) != 1 {
				return fmt.Errorf("%w: expected 1 choice", common.ErrInvalidArgument)
			}
			day := time.Weekday(r.Choices.IDs[0])
			if day > time.Saturday || day < time.Sunday {
				return fmt.Errorf("%w: bad day index %v", common.ErrInvalidArgument, day)
			}

			tmpArray := make([]LabeledSlot, 0, 16)
			for _, slot := range sm.availableSlots {
				if slot.Start.Weekday() == time.Weekday(day) {
					tmpArray = append(tmpArray, slot)
				}
			}
			sm.availableSlots = tmpArray
			return sm.chat.PrintSlots(r, sm.availableSlots)
		case ChoiceTypeSlots:
			//TODO need logic for multiple slot choices
			if len(r.Choices.IDs) != 1 {
				return fmt.Errorf("%w: wrong choices len (%v)", common.ErrInvalidArgument, len(r.Choices.IDs))
			}
			tmpArray := make([]common.Slot, len(r.Choices.IDs))
			tmpPrint := make([]LabeledSlot, len(r.Choices.IDs))
			for _, slot := range sm.availableSlots {
				if slot.ID == r.Choices.IDs[0] {
					tmpArray = append(tmpArray, slot.Interval.ToSlot())
					tmpPrint = append(tmpPrint, slot)
				}
			}

			err = sm.commands.Appointment.AddSlots(r.Ctx, r.Customer, tmpArray)
			if err != nil {
				return err
			}

			return sm.chat.ConfirmAppointment(r, tmpPrint)
		case ChoiceTypeNone:
			return fmt.Errorf("%w: ChoiceType is not set", common.ErrInvalidArgument)
		default:
			return fmt.Errorf("%w: unexpected ChoiceType", common.ErrInternal)
		}
	}
}
