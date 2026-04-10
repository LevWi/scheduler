package command

import (
	"context"
	"errors"
	"fmt"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot/chat"
	"scheduler/appointment-service/internal/bot/i18n/messages"
	"strconv"
	"strings"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

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

type Appointment interface {
	AddSlots(ctx context.Context, customer common.ID, slots []common.Slot) error
}

type commands struct {
	WeekSlots   *WeekSlots
	Appointment Appointment
}

type IdentifyMessageFunc func(string) *messages.MessageConstant

type slotsSmDeps struct {
	MD       *MenuDeps
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
func (sm *SlotSelectionCommand) ShowRangesMenu(c *chat.ChatContext, additional ...*i18n.Message) error {
	options := []*i18n.Message{messages.NextWeek, messages.ThisWeek}
	options = append(options, additional...)
	return sm.deps.MD.Chat().ShowMenuMessages(c, messages.SelectRequestMessage, options)
}

func (sm *SlotSelectionCommand) Process(r *Request) (SlotSelectionResult, error) {
	if sm.availableSlots == nil {
		c := sm.deps.MD.MM.IdentifyMessage(r.Text)
		if c == nil {
			return SlotSelectionResultNotSet, errors.Join(ErrWrongUserInput, common.ErrNotFound)
		}

		var err error
		var slots []common.Slot
		switch c {
		case messages.NextWeek:
			slots, err = sm.deps.Commands.WeekSlots.NextWeek(r.Ctx, r.Time.In(sm.deps.MD.UserSettings.TimeZone))
		case messages.ThisWeek:
			slots, err = sm.deps.Commands.WeekSlots.ThisWeek(r.Ctx, r.Time.In(sm.deps.MD.UserSettings.TimeZone))
		default:
			err = fmt.Errorf("%w: unexpected message text ID %s (%s)", ErrWrongUserInput, c.ID, r.Text)
		}

		if err != nil {
			return SlotSelectionResultNotSet, err
		}

		if len(slots) == 0 {
			return SlotSelectionResultContinue, sm.deps.MD.Chat().PrintMessage(r.ChatContext, messages.NoSlotsAvailable)
		}

		sm.availableSlots = ToLabeledSlot(slots)
		return SlotSelectionResultContinue, sm.deps.MD.Chat().ShowAsOptions(r.ChatContext,
			messages.SelectRequestMessage, sm.availableSlots)
	} else {
		if r.Text != "" {
			return SlotSelectionResultContinue, fmt.Errorf("%w: input text should be empty", ErrWrongUserInput)
		}

		if len(r.Choices) == 0 {
			return SlotSelectionResultContinue, fmt.Errorf("%w: expected user choices", ErrWrongUserInput)
		}

		switch {
		case strings.HasPrefix(r.Choices[0], DayMarker):
			//For simplicity only one day now
			if len(r.Choices) != 1 {
				return SlotSelectionResultContinue, fmt.Errorf("%w: expected 1 choice", ErrWrongUserInput)
			}

			timeSrt := r.Choices[0][len(DayMarker):]
			//TODO time zone?
			date, err := time.ParseInLocation(time.DateOnly, timeSrt, sm.deps.MD.UserSettings.TimeZone)
			if err != nil {
				return SlotSelectionResultNotSet, fmt.Errorf("%w: for %v", err, timeSrt)
			}

			day := date.Weekday()
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
			return SlotSelectionResultContinue, sm.deps.MD.Chat().ShowAsOptions(r.ChatContext, messages.SelectRequestMessage, sm.availableSlots)
		case strings.HasPrefix(r.Choices[0], SlotMarker):
			//TODO need logic for multiple slot choices
			if len(r.Choices) != 1 {
				return SlotSelectionResultContinue, fmt.Errorf("%w: wrong choices len (%v)", ErrWrongUserInput, len(r.Choices))
			}

			idxStr := r.Choices[0][len(SlotMarker):]
			id, err := strconv.Atoi(idxStr)
			if err != nil {
				return SlotSelectionResultNotSet, fmt.Errorf("%w: for %v", err, idxStr)
			}

			tmpArray := make([]common.Slot, 0, len(r.Choices))
			//tmpPrint := make([]LabeledSlot, 0, len(r.Choices))
			for _, slot := range sm.availableSlots {
				if slot.ID == id {
					tmpArray = append(tmpArray, slot.Slot)
					//tmpPrint = append(tmpPrint, slot)
				}
			}

			err = sm.deps.Commands.Appointment.AddSlots(r.Ctx, common.ID(r.Customer), tmpArray)
			if err != nil {
				return SlotSelectionResultContinue, err
			}

			return SlotSelectionResultDone, sm.deps.MD.Chat().PrintMessage(r.ChatContext, messages.Done)
		default:
			return SlotSelectionResultContinue, fmt.Errorf("%w: unexpected ChoiceType (%v)",
				common.ErrInvalidArgument, r.Choices[0])
		}
	}
}

func (mm *SlotSelectionCommand) Cancel() {
	mm.availableSlots = nil
}

func newSlotSelectionCommand(md *MenuDeps, weekSlots *WeekSlots, appointment Appointment) *SlotSelectionCommand {
	sm := &SlotSelectionCommand{
		deps: &slotsSmDeps{
			MD: md,
			Commands: &commands{
				WeekSlots:   weekSlots,
				Appointment: appointment,
			},
		},
	}
	return sm
}
