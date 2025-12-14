package command

import (
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot"
)

func newDefaultSlotSelectionCommand(lp LocalizationProvider, chat ChatSlotsOutput, businessID common.ID, connection *bot.SchedulerConnection) *SlotSelectionCommand {
	return NewSlotSelectionCommand(lp, chat,
		NewWeekSlots(&HttpSlotsProvider{baseURL: connection.URL, businessID: businessID}),
		&HttpAppointment{Connection: connection})
}

func NewDefaultMainMenu(lp LocalizationProvider, chat ChatSlotsOutput, businessID common.ID, connection *bot.SchedulerConnection) *MainMenu {
	return NewMainMenu(lp, chat, newDefaultSlotSelectionCommand(lp, chat, businessID, connection))
}
