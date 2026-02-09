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

type MenuDeps struct {
	LangTag string
	Chat    *ChatAdapter
	Loc     *messages.Localization
	MM      messages.MessageMap
}

func (md *MenuDeps) SetLanguage(l string) error {
	if md.LangTag != l {
		md.Loc.SetLanguage(l)
		m, err := commandMap(md.Loc.Localizer())
		if err != nil {
			return err
		}
		md.MM = m
		md.LangTag = l
		md.Chat.Loc = md.Loc
	}
	return nil
}

func NewMenuDeps(ch chat.Chat, loc *messages.Localization) (*MenuDeps, error) {
	ca := &ChatAdapter{
		Chat: ch,
		Loc:  loc,
	}
	m, err := commandMap(loc.Localizer())
	if err != nil {
		return nil, err
	}
	return &MenuDeps{
			LangTag: loc.LangTag(),
			Chat:    ca,
			Loc:     loc,
			MM:      m,
		},
		nil
}

func commandMap(l *i18n.Localizer) (messages.MessageMap, error) {
	return messages.LocalizedMessageMap(l,
		messages.BookSlot,
		messages.Help,
		messages.NextWeek,
		messages.ThisWeek,
		messages.Cancel,
		messages.Done)
}

// TODO cancel by timeout if user not reacted. Add mutex
type MainMenu struct {
	menuDeps     *MenuDeps
	slotCommands *SlotSelectionCommand
	state        mainMenuState
}

func (menu *MainMenu) SetLanguage(l string) error {
	return menu.SetLanguage(l)
}

func (menu *MainMenu) ShowHelp(c *chat.ChatContext) error {
	return menu.menuDeps.Chat.PrintMessage(c, messages.HelpMessage)
}

// TODO fix case sensitive input
func (menu *MainMenu) Process(r *Request) error {
	c := menu.menuDeps.MM.IdentifyMessage(r.Text)

	if c == messages.Cancel {
		menu.BackToStart()
		//TODO print main menu
		return menu.ShowHelp(r.ChatContext)
	}

	switch menu.state {
	case menuStart:
		if c == messages.BookSlot {
			err := menu.slotCommands.ShowRangesMenu(r.ChatContext, messages.Cancel)
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
		result, err := menu.slotCommands.Process(r)
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
	err := menu.menuDeps.Chat.PrintMessage(c, msg)
	if errors.Is(err, ErrLocalizeMessage) {
		err = menu.menuDeps.Chat.Print(c, msg.One)
	}
	return err
}

func (menu *MainMenu) wrongUserInput(l *i18n.Localizer, c *chat.ChatContext) error {
	s, err := l.LocalizeMessage(messages.WrongUserInput)
	if err != nil {
		s = messages.WrongUserInput.One
	}
	return errors.Join(err, menu.menuDeps.Chat.Print(c, s))
}

func (menu *MainMenu) BackToStart() {
	menu.state = menuStart
	menu.slotCommands.Cancel()
}

func newMainMenu(md *MenuDeps, slotCommands *SlotSelectionCommand) *MainMenu {
	return &MainMenu{
		menuDeps:     md,
		slotCommands: slotCommands,
		state:        menuStart,
	}
}
