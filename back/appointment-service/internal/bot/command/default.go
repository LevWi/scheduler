package command

import (
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot"
)

func newDefaultSlotSelectionCommand(chat *ChatAdapter, businessID common.ID, connection *bot.SchedulerConnection) *SlotSelectionCommand {
	return NewSlotSelectionCommand(chat,
		NewWeekSlots(&HttpSlotsProvider{baseURL: connection.URL, businessID: businessID}),
		&HttpAppointment{Connection: connection})
}

func NewDefaultMainMenu(chat *ChatAdapter, businessID common.ID, connection *bot.SchedulerConnection) *MainMenu {
	return NewMainMenu(chat, newDefaultSlotSelectionCommand(chat, businessID, connection))
}
