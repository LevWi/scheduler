package command

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot"
	"scheduler/appointment-service/internal/bot/chat"
	"scheduler/appointment-service/internal/bot/i18n/messages"
	"sort"
	"strings"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type mainMenuState uint

const (
	menuStart mainMenuState = iota
	menuSlotSelection
	menuSettings
)

type MenuDeps struct {
	chat         chat.Chat
	UserSettings *bot.UserSettings
	MM           messages.MessageMap
}

func (md *MenuDeps) Chat() *ChatAdapter {
	return NewChatAdapter(md.chat, md.UserSettings)
}

func (md *MenuDeps) Clone() *MenuDeps {
	l := *md.UserSettings.Loc
	return &MenuDeps{
		chat: md.chat,
		UserSettings: &bot.UserSettings{
			Loc:      &l,
			TimeZone: md.UserSettings.TimeZone,
		},
		MM: md.MM,
	}
}

// TODO LangTag . seems it need to rework
func (md *MenuDeps) SetLanguage(l string) error {
	if md.UserSettings.Loc.Language() != l {
		m, err := commandMap(md.UserSettings.Loc.LocalizerFor(l))
		if err != nil {
			return err
		}
		md.MM = m
		md.UserSettings.Loc.SetLanguage(l)
	}
	return nil
}

func NewMenuDeps(ch chat.Chat, userSettings *bot.UserSettings) (*MenuDeps, error) {
	m, err := commandMap(userSettings.Loc.Localizer())
	if err != nil {
		return nil, err
	}
	return &MenuDeps{
			chat:         ch,
			UserSettings: userSettings,
			MM:           m,
		},
		nil
}

func commandMap(l *i18n.Localizer) (messages.MessageMap, error) {
	return messages.LocalizedMessageMap(l,
		messages.BookSlot,
		messages.Appointments,
		messages.Settings,
		messages.SetLanguage,
		messages.SetTimeZone,
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
	settingsMenu *SettingsMenu
	appointments AppointmentsProvider
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
		} else if c == messages.Appointments {
			return menu.ShowAppointments(r.Ctx, r.ChatContext, common.ID(r.Customer), time.Now().In(menu.menuDeps.UserSettings.TimeZone))
		} else if c == messages.Settings {
			menu.state = menuSettings
			return menu.settingsMenu.ShowMenu(r.ChatContext)
		} else if c == messages.Help || r.Text == "/help" {
			return menu.ShowHelp(r.ChatContext)
		} else {
			return menu.showMessageForce(r.ChatContext, messages.WrongUserInput)
		}
	case menuSettings:
		result, err := menu.settingsMenu.Process(r)
		if err != nil {
			if errors.Is(err, ErrWrongUserInput) {
				return menu.showMessageForce(r.ChatContext, messages.WrongUserInput)
			}
			return err
		}
		switch result {
		case SettingsResultDone:
			return menu.BackToStart(r.ChatContext)
		case SettingsResultContinue:
			return nil
		default:
			return common.ErrInternal
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
	options := []*i18n.Message{messages.BookSlot, messages.Appointments, messages.Settings, messages.Help, messages.Cancel}
	err := menu.menuDeps.Chat().ShowMenuMessages(c, messages.CommandRequestMessage, options)
	if err != nil {
		slog.ErrorContext(c.Ctx, "mainMenu.BackToStart", "err", err.Error())
		return err
	}
	return nil
}

func (menu *MainMenu) ShowAppointments(ctx context.Context, c *chat.ChatContext, customer common.ID, now time.Time) error {
	appointments, err := menu.appointments.CustomerAppointmentsInRange(ctx, customer, common.Interval{Start: now})
	if err != nil {
		return err
	}

	msg := formatAppointmentsMessage(menu.menuDeps.UserSettings.Loc.Localizer(), appointments, menu.menuDeps.UserSettings.TimeZone)
	return menu.menuDeps.Chat().PrintMessage(c, &i18n.Message{ID: "AppointmentsList", Other: msg})
}

func formatAppointmentsMessage(l *i18n.Localizer, appointments []common.Slot, loc *time.Location) string {
	header, err := l.LocalizeMessage(messages.AppointmentsListHeader)
	if err != nil {
		header = messages.AppointmentsListHeader.Other
	}
	noUpcoming, err := l.LocalizeMessage(messages.NoUpcomingAppointments)
	if err != nil {
		noUpcoming = messages.NoUpcomingAppointments.Other
	}

	if len(appointments) == 0 {
		return noUpcoming
	}

	apptSorted := make([]common.Slot, len(appointments))
	copy(apptSorted, appointments)
	sort.Slice(apptSorted, func(i, j int) bool {
		return apptSorted[i].Start.Before(apptSorted[j].Start)
	})

	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")
	for _, appt := range apptSorted {
		b.WriteString("- ")
		b.WriteString(appt.Start.In(loc).Format(time.DateTime))
		b.WriteString(" (")
		b.WriteString(appt.Dur.String())
		b.WriteString(")\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func newMainMenu(md *MenuDeps, slotCommands *SlotSelectionCommand, appointments AppointmentsProvider) *MainMenu {
	return &MainMenu{
		menuDeps:     md,
		slotCommands: slotCommands,
		settingsMenu: newSettingsMenu(md),
		appointments: appointments,
		state:        menuStart,
	}
}
