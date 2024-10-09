package storage

import (
	"time"

	"github.com/teambition/rrule-go"
)

const countLimit = 100
const slotIntervalMin = 15

// TODO add days off
// Note "start" and "until" includes in the range
func generateWeekdays(start time.Time, until time.Time) ([]time.Time, error) {
	rrule, err := rrule.NewRRule(rrule.ROption{
		Freq:     rrule.DAILY,
		Interval: 1,
		Count:    countLimit,
		Byweekday: []rrule.Weekday{
			rrule.MO,
			rrule.TU,
			rrule.WE,
			rrule.TH,
			rrule.FR,
		},
		Dtstart: start,
		Until:   until,
	})

	if err != nil {
		return nil, err
	}

	return rrule.All(), nil
}

// TODO It is not consider time breaks
func generateTimeSlotsForDay(dtStart time.Time, duration time.Duration) ([]time.Time, error) {
	dtStart = dtStart.Truncate(time.Minute)
	maxCount := int(duration.Minutes() / slotIntervalMin)
	end := dtStart.Add(duration)
	r, err := rrule.NewRRule(rrule.ROption{
		Freq:     rrule.MINUTELY,
		Interval: slotIntervalMin,
		Count:    maxCount,
		Dtstart:  dtStart,
		Until:    end,
	})

	if err != nil {
		return nil, err
	}

	return r.All(), nil
}

// func GetAvailableSlotsInRange(db *Storage, business_id common.ID, start time.Time, end time.Time) ([]common.Slot, error) {
// 	var slots []common.Slot

// }
