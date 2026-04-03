package date

import (
	"time"
)

var monthsShortEn = map[time.Month]string{
	time.January:   "Jan",
	time.February:  "Feb",
	time.March:     "Mar",
	time.April:     "Apr",
	time.May:       "May",
	time.June:      "Jun",
	time.July:      "Jul",
	time.August:    "Aug",
	time.September: "Sep",
	time.October:   "Oct",
	time.November:  "Nov",
	time.December:  "Dec",
}

var weekdaysShortEn = map[time.Weekday]string{
	time.Monday:    "Mon",
	time.Tuesday:   "Tue",
	time.Wednesday: "Wed",
	time.Thursday:  "Thu",
	time.Friday:    "Fri",
	time.Saturday:  "Sat",
	time.Sunday:    "Sun",
}

type DateFormatEn struct {
}

func (DateFormatEn) MonthShort(m time.Month) string {
	return monthsShortEn[m]
}

func (DateFormatEn) WeekDayShort(d time.Weekday) string {
	return weekdaysShortEn[d]
}
