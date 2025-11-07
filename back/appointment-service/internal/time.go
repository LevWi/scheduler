package common

import "time"

func DayBeginning(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func NextMonday(t time.Time) time.Time {
	daysUntilMonday := int(time.Monday - t.Weekday())
	if daysUntilMonday <= 0 {
		daysUntilMonday += 7
	}
	return DayBeginning(t.AddDate(0, 0, daysUntilMonday))
}
