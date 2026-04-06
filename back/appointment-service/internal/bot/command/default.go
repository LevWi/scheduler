package command

import (
	"scheduler/appointment-service/internal/bot"
	"scheduler/appointment-service/internal/bot/chat"
	"scheduler/appointment-service/internal/bot/i18n/messages"
)

func NewDefaultMainMenu(chat chat.Chat, l *messages.Localization, connection *bot.SchedulerConnection) (*MainMenu, error) {
	md, err := NewMenuDeps(chat, l)
	if err != nil {
		return nil, err
	}
	appointments := &HttpAppointment{Connection: connection}
	return newMainMenu(md, newDefaultSlotSelectionCommand(md, connection), appointments), nil
}

func newDefaultSlotSelectionCommand(md *MenuDeps, connection *bot.SchedulerConnection) *SlotSelectionCommand {
	ha := &HttpAppointment{Connection: connection}
	return newSlotSelectionCommand(md, NewWeekSlots(ha), ha)
}
