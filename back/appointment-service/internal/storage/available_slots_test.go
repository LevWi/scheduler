package storage

import (
	"fmt"
	"testing"
	"time"
)

func printTimeSlice(ts []time.Time) {
	fmt.Println(len(ts))
	for _, t := range ts {
		fmt.Println(t)
	}
}

func TestRrule(t *testing.T) {
	const daysCount = 2
	start := time.Date(2024, 10, 9, 9, 0, 0, 0, time.UTC)
	until := start.Add((daysCount - 1) * 24 * time.Hour)
	fmt.Println("generateWeekdays")
	res, err := generateWeekdays(start, until)
	if err != nil {
		t.Fatal(err)
	}
	printTimeSlice(res)
	if len(res) != daysCount {
		t.Fatal("wrong days count", len(res))
	}

	fmt.Println("generateTimeSlotsForDay")
	expectedTimeSLotsCount := len(res) * 8 * 60 / slotIntervalMin
	timeSlots := make([]time.Time, 0, expectedTimeSLotsCount)
	for _, el := range res {
		tmp, err := generateTimeSlotsForDay(el, 8*time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		timeSlots = append(timeSlots, tmp...)
	}
	printTimeSlice(timeSlots)
	if len(timeSlots) != expectedTimeSLotsCount {
		t.Fatal("wrong time slots count ", len(timeSlots), "!=", expectedTimeSLotsCount)
	}
}
