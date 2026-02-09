package command

import (
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot"
	"scheduler/appointment-service/internal/bot/i18n/messages"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func newDefaultSlotSelectionCommand(chat *ChatAdapter, businessID common.ID, connection *bot.SchedulerConnection) *SlotSelectionCommand {
	return NewSlotSelectionCommand(chat,
		NewWeekSlots(&HttpSlotsProvider{baseURL: connection.URL, businessID: businessID}),
		&HttpAppointment{Connection: connection})
}

func NewDefaultMainMenu(chat *ChatAdapter, businessID common.ID, connection *bot.SchedulerConnection) *MainMenu {
	return NewMainMenu(chat, newDefaultSlotSelectionCommand(chat, businessID, connection))
}

func CommandMap(l *i18n.Localizer) (messages.MessageMap, error) {
	return messages.LocalizedMessageMap(l,
		messages.BookSlot,
		messages.Help,
		messages.NextWeek,
		messages.ThisWeek,
		messages.Cancel,
		messages.Done)
}
