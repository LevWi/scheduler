package command

import (
	"scheduler/appointment-service/internal/bot"
	"scheduler/appointment-service/internal/bot/chat"
	"scheduler/appointment-service/internal/bot/i18n/messages"
	"sync"
	"time"
)

type DialogsStorage struct {
	mu         sync.RWMutex
	depsProto  *MenuDeps
	connection *bot.SchedulerConnection
	defaultLoc *time.Location
	dialogs    map[Customer]*DialogValue
}

type DialogValue struct {
	ChatID chat.ChatID
	Menu   *MainMenu
}

func NewDialogStorage(chat chat.Chat, l *messages.Localization, defaultTimeZone *time.Location, connection *bot.SchedulerConnection) (*DialogsStorage, error) {
	deps, err := NewMenuDeps(chat, &bot.UserSettings{
		Loc:      l,
		TimeZone: defaultTimeZone,
	})
	if err != nil {
		return nil, err
	}
	return &DialogsStorage{
		depsProto:  deps,
		connection: connection,
		defaultLoc: defaultTimeZone,
		dialogs:    make(map[Customer]*DialogValue),
	}, nil
}

func (ds *DialogsStorage) GetOrCreateMenu(ca Customer, ch chat.ChatID, language string, userTimeZone *time.Location) *MainMenu {
	dialog := ds.GetDialog(ca)
	if dialog != nil && dialog.ChatID == ch {
		return dialog.Menu
	}

	ds.mu.Lock()
	defer ds.mu.Unlock()
	dialog = ds.dialogs[ca]
	if dialog == nil || dialog.ChatID != ch {
		clone := ds.depsProto.Clone()
		if userTimeZone != nil {
			clone.UserSettings.TimeZone = userTimeZone
		} else {
			clone.UserSettings.TimeZone = ds.defaultLoc
		}
		if language != "" {
			_ = clone.SetLanguage(language)
		}
		appointments := &HttpAppointment{Connection: ds.connection}
		dialog = &DialogValue{
			Menu:   newMainMenu(clone, newDefaultSlotSelectionCommand(clone, ds.connection), appointments),
			ChatID: ch,
		}
		ds.dialogs[ca] = dialog
	}

	return dialog.Menu
}

func (ds *DialogsStorage) GetDialog(ca Customer) *DialogValue {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.dialogs[ca]
}
