package command

import (
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot"
)

func NewDefaultSlotsCommandStMachine(mmp MessageMapProvider, chat ChatOutput, businessID common.ID, connection *bot.SchedulerConnection) *SlotsCommandSMachine {
	return NewSlotsCommandStMachine(mmp, chat,
		NewWeekSlots(&HttpSlotsProvider{baseURL: connection.URL, businessID: businessID}),
		&HttpAppointment{Connection: connection})
}
