package command

import (
	"context"
	"errors"
	"fmt"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot/chat"
	"scheduler/appointment-service/internal/bot/i18n/messages"
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

type slotsSmDeps struct {
	Chat     *ChatAdapter
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

// TODO remove it. Only for skeleton
var todo = &i18n.Message{
	ID:  "TBD",
	One: "TBD",
}

// TODO write here or in MainMenu?
func (sm *SlotSelectionCommand) ShowRangesMenu(c *chat.ChatContext, additional ...*i18n.Message) error {
	options := []*i18n.Message{messages.NextWeek, messages.ThisWeek}
	options = append(options, additional...)
	return sm.deps.Chat.ShowMenuMessages(c, todo, options)
}

func (sm *SlotSelectionCommand) Process(r *Request) (SlotSelectionResult, error) {
	if sm.availableSlots == nil {
		c := sm.deps.Chat.IdentifyMessage(r.Text)
		if c == nil {
			return SlotSelectionResultNotSet, errors.Join(ErrWrongUserInput, common.ErrNotFound)
		}

		var err error
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
		return SlotSelectionResultContinue, sm.deps.Chat.ShowAsOptions(r.ChatContext, todo, sm.availableSlots)
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
			return SlotSelectionResultContinue, sm.deps.Chat.ShowAsOptions(r.ChatContext, todo, sm.availableSlots)
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

			return SlotSelectionResultDone, sm.deps.Chat.PrintMessage(r.ChatContext, messages.Done)
		default:
			return SlotSelectionResultContinue, fmt.Errorf("%w: unexpected ChoiceType (%v)", common.ErrInvalidArgument, r.Choices.Type)
		}
	}
}

func (mm *SlotSelectionCommand) Cancel() {
	mm.availableSlots = nil
}

func NewSlotSelectionCommand(chat *ChatAdapter, weekSlots *WeekSlots, appointment Appointment) *SlotSelectionCommand {
	sm := &SlotSelectionCommand{
		deps: &slotsSmDeps{
			Chat: chat,
			Commands: &commands{
				WeekSlots:   weekSlots,
				Appointment: appointment,
			},
		},
	}
	return sm
}
