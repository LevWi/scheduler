package command

import (
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot"
	"scheduler/appointment-service/internal/bot/chat"
	"scheduler/appointment-service/internal/bot/i18n/messages"
)

func NewDefaultMainMenu(chat chat.Chat, l *messages.Localization, businessID common.ID, connection *bot.SchedulerConnection) (*MainMenu, error) {
	md, err := NewMenuDeps(chat, l)
	if err != nil {
		return nil, err
	}
	return newMainMenu(md, newDefaultSlotSelectionCommand(md, businessID, connection)), nil
}

func newDefaultSlotSelectionCommand(md *MenuDeps, businessID common.ID, connection *bot.SchedulerConnection) *SlotSelectionCommand {
	return newSlotSelectionCommand(md,
		NewWeekSlots(&HttpSlotsProvider{baseURL: connection.URL, businessID: businessID}),
		&HttpAppointment{Connection: connection})
}
