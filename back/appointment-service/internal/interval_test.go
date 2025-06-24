package common

import (
	"runtime"
	"slices"
	"testing"
	"time"
)

func TestInterval(t *testing.T) {
	start := time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC)
	interval := Interval{Start: start, End: start.Add(time.Hour)}

	if !interval.IsOverlap(interval) {
		t.Fatalf("interval is not overlapping itself %v", interval)
	}

	if !interval.IsFit(interval) {
		t.Fatalf("interval is not fitting itself %v", interval)
	}

	interval2 := Interval{Start: interval.End, End: interval.End.Add(time.Second)}
	if interval.IsOverlap(interval2) {
		t.Fatalf("intervals should not overlap %v %v", interval, interval2)
	}

	if interval.IsFit(interval2) {
		t.Fatalf("intervals should not fit %v %v", interval, interval2)
	}

	interval2.Start = interval2.Start.Add(-2 * time.Second)
	if !interval.IsOverlap(interval2) {
		t.Fatalf("intervals should not overlap %v %v", interval, interval2)
	}

	if interval.IsFit(interval2) {
		t.Fatalf("intervals should not fit %v %v", interval, interval2)
	}

	interval2.Start = interval.Start.Add(1 * time.Second)
	interval2.End = interval.End.Add(-1 * time.Second)
	if !interval.IsOverlap(interval2) {
		t.Fatalf("intervals is not overlapping %v %v", interval, interval2)
	}

	if !interval.IsFit(interval2) {
		t.Fatalf("intervals is not fitting %v %v", interval, interval2)
	}

	if !interval2.IsOverlap(interval) {
		t.Fatalf("intervals is not overlapping %v %v", interval, interval2)
	}

	if interval2.IsFit(interval) {
		t.Fatalf("intervals is fitting %v %v", interval, interval2)
	}

	interval2.Start = interval.Start
	interval2.End = interval.End.Add(-1 * time.Second)
	if !interval.IsOverlap(interval2) {
		t.Fatalf("intervals is not overlapping %v %v", interval, interval2)
	}

	interval = Interval{Start: start, End: start.Add(time.Hour)}
	result := interval.Subtract(interval)
	if result != nil {
		t.Fatalf("result should be nil %v", result)
	}

	interval2 = Interval{Start: start.Add(30 * time.Minute), End: start.Add(time.Hour)}
	expected := Interval{Start: start, End: start.Add(30 * time.Minute)}
	result = interval.Subtract(interval2)
	if len(result) != 1 || result[0] != expected {
		t.Fatalf("result should be %v but got %v", expected, result)
	}

	interval2 = Interval{Start: start.Add(31 * time.Minute), End: start.Add(time.Hour + time.Minute)}
	expected = Interval{Start: start, End: start.Add(31 * time.Minute)}
	result = interval.Subtract(interval2)
	if len(result) != 1 || result[0] != expected {
		t.Fatalf("result should be %v but got %v", expected, result)
	}

	interval2 = Interval{Start: start, End: start.Add(30 * time.Minute)}
	expected = Interval{Start: start.Add(30 * time.Minute), End: start.Add(time.Hour)}
	result = interval.Subtract(interval2)
	if len(result) != 1 || result[0] != expected {
		t.Fatalf("result should be %v but got %v", expected, result)
	}

	interval2 = Interval{Start: start.Add(-1 * time.Minute), End: start.Add(31 * time.Minute)}
	expected = Interval{Start: start.Add(31 * time.Minute), End: start.Add(time.Hour)}
	result = interval.Subtract(interval2)
	if len(result) != 1 || result[0] != expected {
		t.Fatalf("result should be %v but got %v", expected, result)
	}

	interval2 = Interval{Start: start.Add(10 * time.Minute), End: start.Add(time.Hour - 10*time.Minute)}
	expected2 := Intervals{
		{Start: start, End: start.Add(10 * time.Minute)},
		{Start: start.Add(time.Hour - 10*time.Minute), End: start.Add(time.Hour)},
	}
	result = interval.Subtract(interval2)
	if len(result) != 2 || result[0] != expected2[0] || result[1] != expected2[1] {
		t.Fatalf("result should be %v but got %v", expected2, result)
	}
}

func TestIntervalSlices(t *testing.T) {
	start := time.Now()
	interval := Interval{Start: start, End: start.Add(time.Minute)}

	intervals := Intervals{interval, interval}
	intervals.SortByStart()
	if !intervals.IsSorted() {
		t.Fatalf("intervals should be sorted %v", intervals)
	}

	if !intervals.HasOverlaps() {
		t.Fatalf("intervals should be overlap %v", intervals)
	}

	intervals = Intervals{}
	for i := range 10 {
		tp := start.Add(time.Duration(i) * time.Minute)
		intervals = append(intervals, Interval{Start: tp, End: tp.Add(1 * time.Minute)})
	}
	intervals.SortByStart()
	if !intervals.IsSorted() {
		t.Fatalf("intervals should be sorted %v", intervals)
	}
	if intervals.HasOverlaps() {
		t.Fatalf("intervals should has overlap %v", intervals)
	}

	intervals = append(intervals, intervals[0])
	if intervals.IsSorted() {
		t.Fatalf("intervals should not be sorted %v", intervals)
	}
	intervals.SortByStart()
	if !intervals.IsSorted() {
		t.Fatalf("intervals should not be sorted %v", intervals)
	}
	if !intervals.HasOverlaps() {
		t.Fatalf("intervals should has overlap %v", intervals)
	}
}

func TestIntervalsUnite(t *testing.T) {
	type Expected = Intervals
	checkCase := func(intervals Intervals, expected Expected) {
		_, filename, line, _ := runtime.Caller(1)
		intervals.Unite()
		if !slices.Equal(intervals, expected) {
			t.Fatalf("%s:%d: expected %v, got %v", filename, line, expected, intervals)
		}
	}

	// Test empty intervals
	checkCase(Intervals{}, Expected{})

	// Test single interval
	singleInterval := Intervals{
		{Start: time.Date(2024, 10, 9, 9, 0, 0, 0, time.UTC), End: time.Date(2024, 10, 9, 10, 0, 0, 0, time.UTC)},
	}
	checkCase(singleInterval, singleInterval.Copy())

	// Test non-overlapping intervals
	nonOverlapping := Intervals{
		{Start: time.Date(2024, 10, 9, 9, 0, 0, 0, time.UTC), End: time.Date(2024, 10, 9, 10, 0, 0, 0, time.UTC)},
		{Start: time.Date(2024, 10, 9, 11, 0, 0, 0, time.UTC), End: time.Date(2024, 10, 9, 12, 0, 0, 0, time.UTC)},
		{Start: time.Date(2024, 10, 9, 13, 0, 0, 0, time.UTC), End: time.Date(2024, 10, 9, 14, 0, 0, 0, time.UTC)},
	}
	checkCase(nonOverlapping, nonOverlapping.Copy())

	// Test overlapping intervals
	overlapping := Intervals{
		{Start: time.Date(2024, 10, 9, 9, 0, 0, 0, time.UTC), End: time.Date(2024, 10, 9, 11, 0, 0, 0, time.UTC)},
		{Start: time.Date(2024, 10, 9, 10, 30, 0, 0, time.UTC), End: time.Date(2024, 10, 9, 12, 1, 0, 0, time.UTC)},
		{Start: time.Date(2024, 10, 9, 12, 0, 0, 0, time.UTC), End: time.Date(2024, 10, 9, 13, 0, 0, 0, time.UTC)},
	}
	expectedOverlapping := Intervals{
		{Start: time.Date(2024, 10, 9, 9, 0, 0, 0, time.UTC), End: time.Date(2024, 10, 9, 13, 0, 0, 0, time.UTC)},
	}
	checkCase(overlapping, expectedOverlapping)

	// Test multiple overlapping intervals
	multipleOverlapping := Intervals{
		{Start: time.Date(2024, 10, 9, 9, 0, 0, 0, time.UTC), End: time.Date(2024, 10, 9, 11, 0, 0, 0, time.UTC)},
		{Start: time.Date(2024, 10, 9, 10, 30, 0, 0, time.UTC), End: time.Date(2024, 10, 9, 12, 0, 0, 0, time.UTC)},
		{Start: time.Date(2024, 10, 9, 12, 0, 0, 0, time.UTC), End: time.Date(2024, 10, 9, 13, 0, 0, 0, time.UTC)},
		{Start: time.Date(2024, 10, 9, 13, 0, 0, 0, time.UTC), End: time.Date(2024, 10, 9, 14, 1, 0, 0, time.UTC)},
		{Start: time.Date(2024, 10, 9, 14, 0, 0, 0, time.UTC), End: time.Date(2024, 10, 9, 15, 0, 0, 0, time.UTC)},
		{Start: time.Date(2024, 10, 9, 14, 0, 0, 0, time.UTC), End: time.Date(2024, 10, 9, 15, 0, 0, 0, time.UTC)},
		{Start: time.Date(2024, 10, 9, 15, 0, 0, 0, time.UTC), End: time.Date(2024, 10, 9, 16, 0, 0, 0, time.UTC)},
	}
	expectedMultipleOverlapping := Intervals{
		{Start: time.Date(2024, 10, 9, 9, 0, 0, 0, time.UTC), End: time.Date(2024, 10, 9, 12, 0, 0, 0, time.UTC)},
		{Start: time.Date(2024, 10, 9, 12, 0, 0, 0, time.UTC), End: time.Date(2024, 10, 9, 13, 0, 0, 0, time.UTC)},
		{Start: time.Date(2024, 10, 9, 13, 0, 0, 0, time.UTC), End: time.Date(2024, 10, 9, 15, 0, 0, 0, time.UTC)},
		{Start: time.Date(2024, 10, 9, 15, 0, 0, 0, time.UTC), End: time.Date(2024, 10, 9, 16, 0, 0, 0, time.UTC)},
	}
	checkCase(multipleOverlapping, expectedMultipleOverlapping)
}

func TestSetPassedIntervals(t *testing.T) {
	type Expected = Intervals
	checkCase := func(i Intervals, e Intervals, expected Expected) {
		result := i.PassedIntervals(e)
		if !slices.Equal(result, expected) {
			t.Fatalf("expected %v, got %v", expected, result)
		}
	}

	checkCase(
		Intervals{
			{
				time.Date(2024, 10, 9, 9, 0, 0, 0, time.UTC),
				time.Date(2024, 10, 9, 18, 0, 0, 0, time.UTC),
			},
		},
		Intervals{
			{
				time.Date(2024, 10, 9, 12, 0, 0, 0, time.UTC),
				time.Date(2024, 10, 9, 13, 0, 0, 0, time.UTC)},
		},

		Expected{
			{
				time.Date(2024, 10, 9, 9, 0, 0, 0, time.UTC),
				time.Date(2024, 10, 9, 12, 0, 0, 0, time.UTC),
			}, {
				time.Date(2024, 10, 9, 13, 0, 0, 0, time.UTC),
				time.Date(2024, 10, 9, 18, 0, 0, 0, time.UTC),
			},
		},
	)

	checkCase(
		Intervals{
			{
				time.Date(2024, 10, 9, 9, 0, 0, 0, time.UTC),
				time.Date(2024, 10, 9, 18, 0, 0, 0, time.UTC),
			},
		},
		Intervals{
			{
				time.Date(2024, 10, 9, 12, 0, 0, 0, time.UTC),
				time.Date(2024, 10, 9, 18, 0, 0, 0, time.UTC),
			},
		},

		Expected{
			{
				Start: time.Date(2024, 10, 9, 9, 0, 0, 0, time.UTC),
				End:   time.Date(2024, 10, 9, 12, 0, 0, 0, time.UTC),
			},
		},
	)

	checkCase(
		Intervals{
			{
				time.Date(2024, 10, 9, 9, 0, 0, 0, time.UTC),
				time.Date(2024, 10, 9, 18, 0, 0, 0, time.UTC),
			},
		},
		Intervals{
			{
				time.Date(2024, 10, 9, 8, 0, 0, 0, time.UTC),
				time.Date(2024, 10, 9, 18, 0, 0, 0, time.UTC),
			},
		},

		Expected{},
	)

	checkCase(
		Intervals{},
		Intervals{
			{
				time.Date(2024, 10, 9, 8, 0, 0, 0, time.UTC),
				time.Date(2024, 10, 9, 18, 0, 0, 0, time.UTC),
			},
		},

		Expected{},
	)

	checkCase(
		Intervals{
			{
				time.Date(2024, 10, 9, 9, 0, 0, 0, time.UTC),
				time.Date(2024, 10, 9, 18, 0, 0, 0, time.UTC),
			},

			{
				time.Date(2024, 10, 9, 20, 0, 0, 0, time.UTC),
				time.Date(2024, 10, 9, 21, 0, 0, 0, time.UTC),
			},
		},
		Intervals{
			{
				time.Date(2024, 10, 9, 10, 0, 0, 0, time.UTC),
				time.Date(2024, 10, 9, 11, 0, 0, 0, time.UTC),
			},

			{
				time.Date(2024, 10, 9, 13, 0, 0, 0, time.UTC),
				time.Date(2024, 10, 9, 14, 0, 0, 0, time.UTC),
			},

			{
				time.Date(2024, 10, 9, 15, 0, 0, 0, time.UTC),
				time.Date(2024, 10, 9, 16, 0, 0, 0, time.UTC),
			},
		},

		Expected{
			{
				time.Date(2024, 10, 9, 9, 0, 0, 0, time.UTC),
				time.Date(2024, 10, 9, 10, 0, 0, 0, time.UTC),
			},
			{
				time.Date(2024, 10, 9, 11, 0, 0, 0, time.UTC),
				time.Date(2024, 10, 9, 13, 0, 0, 0, time.UTC),
			},
			{
				time.Date(2024, 10, 9, 14, 0, 0, 0, time.UTC),
				time.Date(2024, 10, 9, 15, 0, 0, 0, time.UTC),
			},
			{
				time.Date(2024, 10, 9, 16, 0, 0, 0, time.UTC),
				time.Date(2024, 10, 9, 18, 0, 0, 0, time.UTC),
			},
			{
				time.Date(2024, 10, 9, 20, 0, 0, 0, time.UTC),
				time.Date(2024, 10, 9, 21, 0, 0, 0, time.UTC),
			},
		},
	)

}
