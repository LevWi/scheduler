package messages

import (
	"scheduler/appointment-service/internal/bot/i18n/date"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// TODO race condition?
type Localization struct {
	bundle  *i18n.Bundle
	langTag string
	DF      DateFormatter //TODO not used now
}

func NewLocalization(bundle *i18n.Bundle, langTag string) *Localization {
	return &Localization{
		bundle:  bundle,
		langTag: langTag,
		DF:      date.DateFormatRu{},
	}
}

func (l *Localization) Language() string {
	return l.langTag
}

func (l *Localization) SetLanguage(langTag string) {
	l.langTag = langTag
	//TODO set DateFormatter
}

func (l *Localization) Localizer() *i18n.Localizer {
	return i18n.NewLocalizer(l.bundle, l.langTag)
}

func (l *Localization) LocalizerFor(lang string) *i18n.Localizer {
	return i18n.NewLocalizer(l.bundle, lang)
}

type DateFormatter interface {
	MonthShort(m time.Month) string
	WeekDayShort(d time.Weekday) string
}
