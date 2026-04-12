package date

import (
	"time"
)

var monthsShortKk = map[time.Month]string{
	time.January:   "Қаң",
	time.February:  "Ақп",
	time.March:     "Нау",
	time.April:     "Сәу",
	time.May:       "Мам",
	time.June:      "Мау",
	time.July:      "Шіл",
	time.August:    "Там",
	time.September: "Қыр",
	time.October:   "Қаз",
	time.November:  "Қар",
	time.December:  "Жел",
}

var weekdaysShortKk = map[time.Weekday]string{
	time.Monday:    "Дс",
	time.Tuesday:   "Сс",
	time.Wednesday: "Ср",
	time.Thursday:  "Бс",
	time.Friday:    "Жм",
	time.Saturday:  "Сб",
	time.Sunday:    "Жс",
}

type DateFormatKK struct {
}

func (DateFormatKK) MonthShort(m time.Month) string {
	return monthsShortKk[m]
}

func (DateFormatKK) WeekDayShort(d time.Weekday) string {
	return weekdaysShortKk[d]
}

func (DateFormatKK) MinShort() string {
	return "мин"
}
