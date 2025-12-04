package command

import (
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot"
)

func NewDefaultSlotSelectionCommand(mmp LocalizationProvider, chat ChatSlotsOutput, businessID common.ID, connection *bot.SchedulerConnection) *SlotSelectionCommand {
	return NewSlotSelectionCommand(mmp, chat,
		NewWeekSlots(&HttpSlotsProvider{baseURL: connection.URL, businessID: businessID}),
		&HttpAppointment{Connection: connection})
}
