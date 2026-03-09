package command

import (
	"errors"
	"fmt"
	"log/slog"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot/chat"
	"scheduler/appointment-service/internal/bot/i18n/messages"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type mainMenuState uint

const (
	menuStart mainMenuState = iota
	menuSlotSelection
)

type MenuDeps struct {
	chat chat.Chat
	Loc  *messages.Localization
	MM   messages.MessageMap
}

func (md *MenuDeps) Chat() *ChatAdapter {
	return NewChatAdapter(md.chat, md.Loc)
}

func (md *MenuDeps) Clone() *MenuDeps {
	l := *md.Loc
	return &MenuDeps{
		chat: md.chat,
		Loc:  &l,
		MM:   md.MM,
	}
}

// TODO LangTag . seems it need to rework
func (md *MenuDeps) SetLanguage(l string) error {
	if md.Loc.Language() != l {
		m, err := commandMap(md.Loc.LocalizerFor(l))
		if err != nil {
			return err
		}
		md.MM = m
		md.Loc.SetLanguage(l)
	}
	return nil
}

func NewMenuDeps(ch chat.Chat, loc *messages.Localization) (*MenuDeps, error) {
	m, err := commandMap(loc.Localizer())
	if err != nil {
		return nil, err
	}
	return &MenuDeps{
			chat: ch,
			Loc:  loc,
			MM:   m,
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
	return menu.menuDeps.SetLanguage(l)
}

func (menu *MainMenu) ShowHelp(c *chat.ChatContext) error {
	return menu.menuDeps.Chat().PrintMessage(c, messages.HelpMessage)
}

// TODO need user time zone setting (https://github.com/LevWi/scheduler/issues/17)
func (menu *MainMenu) Process(r *Request) error {
	r.Text = strings.ToLower(r.Text)
	c := menu.menuDeps.MM.IdentifyMessage(r.Text)

	if r.Text == "/start" {
		c = messages.Cancel
	}

	if c == messages.Cancel {
		menu.BackToStart(r.ChatContext)
		return menu.ShowHelp(r.ChatContext)
	}

	switch menu.state {
	case menuStart:
		if c == messages.BookSlot {
			err := menu.slotCommands.ShowRangesMenu(r.ChatContext, messages.Cancel)
			if err != nil {
				slog.ErrorContext(r.Ctx, "mainMenu.menuStart", "err", err.Error())
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
				return errors.Join(menu.showMessageForce(r.ChatContext, messages.InternalErrorOccurred),
					menu.BackToStart(r.ChatContext))
			}
		}

		switch result {
		case SlotSelectionResultDone:
			return menu.BackToStart(r.ChatContext)
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
	ch := menu.menuDeps.Chat()
	err := ch.PrintMessage(c, msg)
	if errors.Is(err, ErrLocalizeMessage) {
		err = ch.Print(c, msg.One)
	}
	return err
}

func (menu *MainMenu) BackToStart(c *chat.ChatContext) error {
	menu.state = menuStart
	menu.slotCommands.Cancel()
	options := []*i18n.Message{messages.BookSlot, messages.Help, messages.Cancel}
	err := menu.menuDeps.Chat().ShowMenuMessages(c, messages.CommandRequestMessage, options)
	if err != nil {
		slog.ErrorContext(c.Ctx, "mainMenu.BackToStart", "err", err.Error())
		return err
	}
	return nil
}

func newMainMenu(md *MenuDeps, slotCommands *SlotSelectionCommand) *MainMenu {
	return &MainMenu{
		menuDeps:     md,
		slotCommands: slotCommands,
		state:        menuStart,
	}
}
