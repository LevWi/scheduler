package storage

import (
	"time"

	common "scheduler/appointment-service/internal"

	"github.com/teambition/rrule-go"
)

const countLimit = 100
const slotIntervalMin = 15 // Hardcoded. Fix it by reading from db.

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

func (*Storage) GetPossibleSlotsInRange(business_id common.ID, start time.Time, end time.Time) ([]common.Slot, error) {
	start = start.Truncate(time.Hour * 24).Add(time.Hour * 9) //TODO Fix it. hardcoded to 9am
	const workingHours = time.Hour * 8                        // TODO Fix it. Hardcoded 8h
	weekdays, err := generateWeekdays(start, end)

	if err != nil {
		return nil, err
	}

	var times []time.Time
	for _, weekday := range weekdays {
		tmp, err := generateTimeSlotsForDay(weekday, workingHours)
		if err != nil {
			return nil, err
		}
		times = append(times, tmp...)
	}

	out := make([]common.Slot, len(times))
	for inx, t := range times {
		out[inx] = common.Slot{Start: t, Len: slotIntervalMin}
	}

	return out, nil
}

// TODO not optimal. Result is unordered
func (db *Storage) GetAvailableSlotsInRange(business_id common.ID, start time.Time, end time.Time) ([]common.Slot, error) {
	slots, err := db.GetPossibleSlotsInRange(business_id, start, end)
	if err != nil {
		return nil, err
	}

	//TODO check for time order
	busySlots, err := db.GetBusySlotsInRange(business_id, start, end)
	if err != nil {
		return nil, err
	}

	if len(busySlots) == 0 {
		return slots, nil
	}

	availableSlotsMap := make(map[int64]common.Slot, len(slots))
	for _, slot := range slots {
		availableSlotsMap[slot.Start.Unix()] = slot
	}

	for _, slot := range busySlots {
		delete(availableSlotsMap, slot.Start.Unix())
	}

	availableSlots := slots[:len(availableSlotsMap)]
	for inx, slot := range availableSlotsMap {
		availableSlots[inx] = slot
	}

	return availableSlots, nil
}
