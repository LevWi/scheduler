package command

import (
	"scheduler/appointment-service/internal/bot"
	"scheduler/appointment-service/internal/bot/chat"
	"scheduler/appointment-service/internal/bot/i18n/messages"
	"sync"
)

type DialogsStorage struct {
	mu         sync.RWMutex
	depsProto  *MenuDeps
	connection *bot.SchedulerConnection
	dialogs    map[Customer]*DialogValue
}

type DialogValue struct {
	ChatID chat.ChatID
	Menu   *MainMenu
}

func NewDialogStorage(chat chat.Chat, l *messages.Localization, connection *bot.SchedulerConnection) (*DialogsStorage, error) {
	deps, err := NewMenuDeps(chat, l)
	if err != nil {
		return nil, err
	}
	return &DialogsStorage{
		depsProto:  deps,
		connection: connection,
		dialogs:    make(map[Customer]*DialogValue),
	}, nil
}

func (ds *DialogsStorage) GetOrCreateMenu(ca Customer, ch chat.ChatID) *MainMenu {
	dialog := ds.GetDialog(ca)
	if dialog != nil && dialog.ChatID == ch {
		return dialog.Menu
	}

	ds.mu.Lock()
	defer ds.mu.Unlock()
	dialog = ds.dialogs[ca]
	if dialog == nil || dialog.ChatID != ch {
		clone := ds.depsProto.Clone()
		dialog = &DialogValue{
			Menu:   newMainMenu(clone, newDefaultSlotSelectionCommand(clone, ds.connection)),
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
