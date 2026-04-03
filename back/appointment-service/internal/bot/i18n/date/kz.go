package date

import (
	"time"
)

var monthsShortKz = map[time.Month]string{
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

var weekdaysShortKz = map[time.Weekday]string{
	time.Monday:    "Дс",
	time.Tuesday:   "Сс",
	time.Wednesday: "Ср",
	time.Thursday:  "Бс",
	time.Friday:    "Жм",
	time.Saturday:  "Сб",
	time.Sunday:    "Жс",
}

type DateFormatKz struct {
}

func (DateFormatKz) MonthShort(m time.Month) string {
	return monthsShortKz[m]
}

func (DateFormatKz) WeekDayShort(d time.Weekday) string {
	return weekdaysShortKz[d]
}
