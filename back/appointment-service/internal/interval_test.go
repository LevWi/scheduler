package common

import (
	"testing"
	"time"
)

func TestIntervals(t *testing.T) {
	start := time.Now()
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

func TestIntervalSetSlices(t *testing.T) {
	start := time.Date(2024, 10, 9, 9, 0, 0, 0, time.UTC)
	end := time.Date(2024, 10, 9, 18, 0, 0, 0, time.UTC)

	intervals := Intervals{Interval{start, end}}

	if intervals.HasOverlaps() {
		t.Fatalf("intervals should not has overlap %v", intervals)
	}

	set, err := NewIntervalSetWithCopies(intervals, nil)
	if err != nil {
		t.Fatal("unexpected error", set)
	}

	if !set.IsValid() {
		t.Fatalf("set should be valid %v", set)
	}

	set.AddExclusion(Interval{start.Add(3 * time.Hour), start.Add(4 * time.Hour)})
	if !set.IsValid() {
		t.Fatalf("set should be valid %v", set)
	}

	validInterval := Interval{start.Add(1 * time.Hour), start.Add(1*time.Hour + 30*time.Minute)}
	if !set.IsFit(validInterval) {
		t.Fatalf("interval should fit %v", validInterval)
	}

	invalidInterval := Interval{start.Add(1 * time.Hour), start.Add(4*time.Hour + 30*time.Minute)}
	if set.IsFit(invalidInterval) {
		t.Fatalf("interval should not fit %v", invalidInterval)
	}

	validInterval = Interval{start.Add(2 * time.Hour), start.Add(3 * time.Hour)}
	if !set.IsFit(validInterval) {
		t.Fatalf("interval should fit %v", validInterval)
	}

	invalidInterval = validInterval
	invalidInterval.End = invalidInterval.End.Add(1 * time.Minute)
	if set.IsFit(invalidInterval) {
		t.Fatalf("interval should not fit %v", invalidInterval)
	}
}
