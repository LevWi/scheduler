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

	if !intervals.IsOverlap() {
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
	if intervals.IsOverlap() {
		t.Fatalf("intervals should not be overlap %v", intervals)
	}

	intervals = append(intervals, intervals[0])
	if intervals.IsSorted() {
		t.Fatalf("intervals should not be sorted %v", intervals)
	}
	intervals.SortByStart()
	if !intervals.IsSorted() {
		t.Fatalf("intervals should not be sorted %v", intervals)
	}
	if !intervals.IsOverlap() {
		t.Fatalf("intervals should be overlap %v", intervals)
	}
}
