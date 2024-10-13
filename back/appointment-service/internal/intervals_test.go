package common

import (
	"testing"
	"time"
)

func TestIntervals(t *testing.T) {
	start := time.Now()
	interval := Interval{Start: start, End: start.Add(time.Hour)}

	if !interval.IsOverlaps(interval) {
		t.Fatalf("interval is not overlapping itself %v", interval)
	}

	if !interval.IsFit(interval) {
		t.Fatalf("interval is not fitting itself %v", interval)
	}

	interval2 := Interval{Start: interval.End, End: interval.End.Add(time.Second)}
	if interval.IsOverlaps(interval2) {
		t.Fatalf("intervals should not overlap %v %v", interval, interval2)
	}

	if interval.IsFit(interval2) {
		t.Fatalf("intervals should not fit %v %v", interval, interval2)
	}

	interval2.Start = interval2.Start.Add(-2 * time.Second)
	if !interval.IsOverlaps(interval2) {
		t.Fatalf("intervals should not overlap %v %v", interval, interval2)
	}

	if interval.IsFit(interval2) {
		t.Fatalf("intervals should not fit %v %v", interval, interval2)
	}

	interval2.Start = interval.Start.Add(1 * time.Second)
	interval2.End = interval.End.Add(-1 * time.Second)
	if !interval.IsOverlaps(interval2) {
		t.Fatalf("intervals is not overlapping %v %v", interval, interval2)
	}

	if !interval.IsFit(interval2) {
		t.Fatalf("intervals is not fitting %v %v", interval, interval2)
	}

	if !interval2.IsOverlaps(interval) {
		t.Fatalf("intervals is not overlapping %v %v", interval, interval2)
	}

	if interval2.IsFit(interval) {
		t.Fatalf("intervals is fitting %v %v", interval, interval2)
	}
}
