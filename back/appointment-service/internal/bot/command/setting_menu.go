package command

import (
	"fmt"
	"scheduler/appointment-service/internal/bot/chat"
	"scheduler/appointment-service/internal/bot/i18n/messages"
	"strings"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type settingsState uint

const (
	settingsStart settingsState = iota
	settingsWaitLanguage
	settingsWaitTimeZone
)

type SettingsMenu struct {
	deps  *MenuDeps
	state settingsState
}

type SettingsResult uint

const (
	SettingsResultNotSet SettingsResult = iota
	SettingsResultContinue
	SettingsResultDone
)

func newSettingsMenu(md *MenuDeps) *SettingsMenu {
	return &SettingsMenu{
		deps:  md,
		state: settingsStart,
	}
}

func (sm *SettingsMenu) ShowMenu(c *chat.ChatContext) error {
	sm.state = settingsStart
	return sm.deps.Chat().ShowMenuMessages(c, messages.SelectSettingsOptionMessage,
		[]*i18n.Message{messages.SetLanguage, messages.SetTimeZone, messages.Cancel})
}

func (sm *SettingsMenu) Process(r *Request) (SettingsResult, error) {
	switch sm.state {
	case settingsStart:
		c := sm.deps.MM.IdentifyMessage(r.Text)
		switch c {
		case messages.SetLanguage:
			sm.state = settingsWaitLanguage
			return SettingsResultContinue, sm.deps.Chat().ShowMenuMessages(r.ChatContext, messages.SelectLanguageMessage,
				[]*i18n.Message{
					{ID: "LangRu", Other: "ru"},
					{ID: "LangEn", Other: "en"},
					{ID: "LangKz", Other: "kz"},
					messages.Cancel,
				})
		case messages.SetTimeZone:
			sm.state = settingsWaitTimeZone
			return SettingsResultContinue, sm.deps.Chat().PrintMessage(r.ChatContext, messages.EnterTimeZoneMessage)
		case messages.Cancel:
			sm.state = settingsStart
			return SettingsResultDone, nil
		default:
			return SettingsResultNotSet, ErrWrongUserInput
		}
	case settingsWaitLanguage:
		lang := strings.TrimSpace(strings.ToLower(r.Text))
		switch lang {
		case "ru", "en", "kz":
			if err := sm.deps.SetLanguage(lang); err != nil {
				return SettingsResultNotSet, err
			}
			sm.state = settingsStart
			if err := sm.deps.Chat().PrintMessage(r.ChatContext, messages.LanguageUpdated); err != nil {
				return SettingsResultNotSet, err
			}
			return SettingsResultContinue, sm.ShowMenu(r.ChatContext)
		default:
			return SettingsResultNotSet, ErrWrongUserInput
		}
	case settingsWaitTimeZone:
		loc, err := time.LoadLocation(strings.TrimSpace(r.Text))
		if err != nil {
			return SettingsResultContinue, sm.deps.Chat().PrintMessage(r.ChatContext, messages.InvalidTimeZone)
		}
		sm.deps.UserSettings.TimeZone = loc
		sm.state = settingsStart
		if err := sm.deps.Chat().PrintMessage(r.ChatContext, messages.TimeZoneUpdated); err != nil {
			return SettingsResultNotSet, err
		}
		return SettingsResultContinue, sm.ShowMenu(r.ChatContext)
	default:
		return SettingsResultNotSet, fmt.Errorf("unexpected settings state: %v", sm.state)
	}
}
