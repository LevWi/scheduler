package command

import (
	"context"
	"errors"
	"fmt"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot/i18n/messages"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type ChatSlotsOutput interface {
	ChatOutput
	PrintSlots(c *ChatContext, message string, m []LabeledSlot) error
	ConfirmAppointment(c *ChatContext, m []LabeledSlot) error
}

type LabeledSlot struct {
	ID int
	common.Slot
}

func ToLabeledSlot(slots []common.Slot) []LabeledSlot {
	ls := make([]LabeledSlot, len(slots))
	for i, s := range slots {
		ls[i] = LabeledSlot{ID: i, Slot: s}
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

type Appointment interface {
	AddSlots(ctx context.Context, customer common.ID, slots []common.Slot) error
}

type commands struct {
	WeekSlots   *WeekSlots
	Appointment Appointment
}

type slotsSmDeps struct {
	LP       LocalizationProvider
	Chat     ChatSlotsOutput
	Commands *commands
}

type SlotSelectionCommand struct {
	availableSlots []LabeledSlot
	deps           *slotsSmDeps
}

type SlotSelectionResult uint

const (
	SlotSelectionResultNotSet SlotSelectionResult = iota
	SlotSelectionResultContinue
	SlotSelectionResultDone
)

// TODO write here or in MainMenu?
func (sm *SlotSelectionCommand) ShowRangesMenu(c *ChatContext, additional ...*i18n.Message) error {
	l, err := sm.deps.LP.Localizer()
	if err != nil {
		return err
	}

	options := []*i18n.Message{messages.NextWeek, messages.ThisWeek}
	options = append(options, additional...)
	localized, err := messages.LocalizeMessages(l, options)
	if err != nil {
		return err
	}
	return sm.deps.Chat.ShowMenu(c, "TBD", localized)
}

func (sm *SlotSelectionCommand) Process(r *Request) (SlotSelectionResult, error) {
	if sm.availableSlots == nil {
		m, err := sm.deps.LP.LocalizedMap()
		if err != nil {
			return SlotSelectionResultNotSet, err
		}

		c, ok := m[r.Text]
		if !ok {
			return SlotSelectionResultNotSet, errors.Join(ErrWrongUserInput, common.ErrNotFound)
		}

		var slots []common.Slot
		switch c {
		case messages.NextWeek:
			slots, err = sm.deps.Commands.WeekSlots.NextWeek(r.Ctx, r.Now)
		case messages.ThisWeek:
			slots, err = sm.deps.Commands.WeekSlots.ThisWeek(r.Ctx, r.Now)
		default:
			err = fmt.Errorf("%w: unexpected message text ID %s (%s)", ErrWrongUserInput, c.ID, r.Text)
		}

		if err != nil {
			return SlotSelectionResultNotSet, err
		}
		sm.availableSlots = ToLabeledSlot(slots)
		return SlotSelectionResultContinue, sm.deps.Chat.PrintSlots(r.ChatContext, "TBD123", sm.availableSlots)
	} else {
		if r.Text != "" {
			return SlotSelectionResultContinue, fmt.Errorf("%w: input text should be empty", ErrWrongUserInput)
		}

		if len(r.Choices.IDs) == 0 {
			return SlotSelectionResultContinue, fmt.Errorf("%w: expected user choices", ErrWrongUserInput)
		}

		switch r.Choices.Type {
		case ChoiceTypeDays:
			//For simplicity only one day now
			if len(r.Choices.IDs) != 1 {
				return SlotSelectionResultContinue, fmt.Errorf("%w: expected 1 choice", ErrWrongUserInput)
			}
			day := time.Weekday(r.Choices.IDs[0])
			if day > time.Saturday || day < time.Sunday {
				return SlotSelectionResultContinue, fmt.Errorf("%w: bad day index %v", ErrWrongUserInput, day)
			}

			tmpArray := make([]LabeledSlot, 0, 16)
			for _, slot := range sm.availableSlots {
				if slot.Start.Weekday() == time.Weekday(day) {
					tmpArray = append(tmpArray, slot)
				}
			}
			sm.availableSlots = tmpArray
			return SlotSelectionResultContinue, sm.deps.Chat.PrintSlots(r.ChatContext, "TBD151", sm.availableSlots)
		case ChoiceTypeSlots:
			//TODO need logic for multiple slot choices
			if len(r.Choices.IDs) != 1 {
				return SlotSelectionResultContinue, fmt.Errorf("%w: wrong choices len (%v)", ErrWrongUserInput, len(r.Choices.IDs))
			}
			tmpArray := make([]common.Slot, len(r.Choices.IDs))
			tmpPrint := make([]LabeledSlot, len(r.Choices.IDs))
			for _, slot := range sm.availableSlots {
				if slot.ID == r.Choices.IDs[0] {
					tmpArray = append(tmpArray, slot.Slot)
					tmpPrint = append(tmpPrint, slot)
				}
			}

			err := sm.deps.Commands.Appointment.AddSlots(r.Ctx, r.Customer, tmpArray)
			if err != nil {
				return SlotSelectionResultContinue, err
			}

			return SlotSelectionResultDone, sm.deps.Chat.ConfirmAppointment(r.ChatContext, tmpPrint)
		default:
			return SlotSelectionResultContinue, fmt.Errorf("%w: unexpected ChoiceType (%v)", common.ErrInvalidArgument, r.Choices.Type)
		}
	}
}

func (mm *SlotSelectionCommand) Cancel() {
	mm.availableSlots = nil
}

func NewSlotSelectionCommand(mmp LocalizationProvider, chat ChatSlotsOutput, weekSlots *WeekSlots, appointment Appointment) *SlotSelectionCommand {
	sm := &SlotSelectionCommand{
		deps: &slotsSmDeps{
			LP:   mmp,
			Chat: chat,
			Commands: &commands{
				WeekSlots:   weekSlots,
				Appointment: appointment,
			},
		},
	}
	return sm
}
