package bot

import (
	"scheduler/appointment-service/internal/bot/i18n/messages"
	"time"
)

type UserSettings struct {
	Loc      *messages.Localization
	TimeZone *time.Location
}
