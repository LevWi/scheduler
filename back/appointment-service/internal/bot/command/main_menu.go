package command

import (
	"errors"
	"fmt"
	"log/slog"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot/i18n/messages"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type mainMenuState uint

const (
	menuStart mainMenuState = iota
	menuSlotSelection
)

type ChatOutput interface {
	Print(c *ChatContext, message string) error
	ShowMenu(c *ChatContext, message string, options []string) error
}

// TODO cancel by timeout if user not reacted. Add mutex
type MainMenu struct {
	LP           LocalizationProvider
	Chat         ChatOutput
	SlotCommands *SlotSelectionCommand
	state        mainMenuState
}

func (menu *MainMenu) ShowHelp(c *ChatContext) error {
	l, err := menu.LP.Localizer()
	if err != nil {
		return err
	}

	localized, err := l.LocalizeMessage(messages.HelpMessage)
	if err != nil {
		return err
	}
	return menu.Chat.Print(c, localized)
}

// TODO fix case sensitive input
func (menu *MainMenu) Process(r *Request) error {
	m, err := menu.LP.LocalizedMap()
	if err != nil {
		return err
	}

	l, err := menu.LP.Localizer()
	if err != nil {
		return err
	}

	c := m[r.Text]

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
			return menu.wrongUserInput(l, r.ChatContext)
		}
	case menuSlotSelection:
		result, err := menu.SlotCommands.Process(r)
		if err != nil {
			if errors.Is(err, ErrWrongUserInput) {
				//This is possible behavior for user. Not an error
				slog.DebugContext(r.Ctx, "mainMenu", "err", err.Error())
				return menu.wrongUserInput(l, r.ChatContext)
			} else {
				slog.ErrorContext(r.Ctx, "mainMenu menuSlotSelection process", "err", err.Error())
				menu.BackToStart()

				s, err2 := l.LocalizeMessage(messages.InternalErrorOccurred)
				if err2 != nil {
					s = messages.InternalErrorOccurred.One
				}
				return errors.Join(err, err2, menu.Chat.Print(r.ChatContext, s))
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
	return err
}

func (menu *MainMenu) wrongUserInput(l *i18n.Localizer, c *ChatContext) error {
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

func NewMainMenu(lp LocalizationProvider, chat ChatOutput, slotCommands *SlotSelectionCommand) *MainMenu {
	return &MainMenu{
		LP:           lp,
		Chat:         chat,
		SlotCommands: slotCommands,
		state:        menuStart,
	}
}
