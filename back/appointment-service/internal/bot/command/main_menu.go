package command

import (
	"errors"
	"fmt"
	"log/slog"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot/chat"
	"scheduler/appointment-service/internal/bot/i18n/messages"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type mainMenuState uint

const (
	menuStart mainMenuState = iota
	menuSlotSelection
)

// TODO cancel by timeout if user not reacted. Add mutex
type MainMenu struct {
	Chat         *ChatAdapter
	SlotCommands *SlotSelectionCommand
	state        mainMenuState
}

func (menu *MainMenu) ShowHelp(c *chat.ChatContext) error {
	return menu.Chat.PrintMessage(c, messages.HelpMessage)
}

// TODO fix case sensitive input
func (menu *MainMenu) Process(r *Request) error {
	c := menu.Chat.IdentifyMessage(r.Text)

	if c == messages.Cancel {
		menu.BackToStart()
		//TODO print main menu
		return menu.ShowHelp(r.ChatContext)
	}

	switch menu.state {
	case menuStart:
		if c == messages.BookSlot {
			err := menu.SlotCommands.ShowRangesMenu(r.ChatContext, messages.Cancel)
			if err != nil {
				slog.ErrorContext(r.Ctx, "mainMenu menuStart", "err", err.Error())
				return err
			}
			menu.state = menuSlotSelection
		} else if c == messages.Help || r.Text == "/help" {
			return menu.ShowHelp(r.ChatContext)
		} else {
			return menu.showMessageForce(r.ChatContext, messages.WrongUserInput)
		}
	case menuSlotSelection:
		result, err := menu.SlotCommands.Process(r)
		if err != nil {
			if errors.Is(err, ErrWrongUserInput) {
				//This is possible behavior for user. Not an error
				slog.DebugContext(r.Ctx, "mainMenu", "err", err.Error())
				return menu.showMessageForce(r.ChatContext, messages.WrongUserInput)
			} else {
				slog.ErrorContext(r.Ctx, "mainMenu menuSlotSelection process", "err", err.Error())
				menu.BackToStart()

				return menu.showMessageForce(r.ChatContext, messages.InternalErrorOccurred)
			}
		}

		switch result {
		case SlotSelectionResultDone:
			menu.state = menuStart
		case SlotSelectionResultNotSet:
			slog.ErrorContext(r.Ctx, "mainMenu menuSlotSelection unexpected result")
			return common.ErrInternal
		}
	default:
		return fmt.Errorf("%w: unexpected mainMenu state %v", common.ErrInternal, menu.state)
	}
	return nil
}

func (menu *MainMenu) showMessageForce(c *chat.ChatContext, msg *i18n.Message) error {
	err := menu.Chat.PrintMessage(c, msg)
	if errors.Is(err, ErrLocalizeMessage) {
		err = menu.Chat.Print(c, msg.One)
	}
	return err
}

func (menu *MainMenu) wrongUserInput(l *i18n.Localizer, c *chat.ChatContext) error {
	s, err := l.LocalizeMessage(messages.WrongUserInput)
	if err != nil {
		s = messages.WrongUserInput.One
	}
	return errors.Join(err, menu.Chat.Print(c, s))
}

func (menu *MainMenu) BackToStart() {
	menu.state = menuStart
	menu.SlotCommands.Cancel()
}

func NewMainMenu(chat *ChatAdapter, slotCommands *SlotSelectionCommand) *MainMenu {
	return &MainMenu{
		Chat:         chat,
		SlotCommands: slotCommands,
		state:        menuStart,
	}
}
