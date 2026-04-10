package date

import (
	"time"
)

var monthsShortRu = map[time.Month]string{
	time.January:   "Янв",
	time.February:  "Фев",
	time.March:     "Мар",
	time.April:     "Апр",
	time.May:       "Май",
	time.June:      "Июн",
	time.July:      "Июл",
	time.August:    "Авг",
	time.September: "Сен",
	time.October:   "Окт",
	time.November:  "Ноя",
	time.December:  "Дек",
}

var weekdaysShortRu = map[time.Weekday]string{
	time.Monday:    "Пн",
	time.Tuesday:   "Вт",
	time.Wednesday: "Ср",
	time.Thursday:  "Чт",
	time.Friday:    "Пт",
	time.Saturday:  "Сб",
	time.Sunday:    "Вс",
}

type DateFormatRu struct {
}

func (DateFormatRu) MonthShort(m time.Month) string {
	return monthsShortRu[m]
}

func (DateFormatRu) WeekDayShort(d time.Weekday) string {
	return weekdaysShortRu[d]
}

func (DateFormatRu) MinShort() string {
	return "мин"
}
