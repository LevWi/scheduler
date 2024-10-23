package common

import (
	"slices"
	"time"
)

type Interval struct {
	Start time.Time
	End   time.Time
}

type Intervals []Interval

func (i Interval) IsValid() bool {
	return i.Start.Before(i.End)
}

func (i Interval) IsOverlap(other Interval) bool {
	return i.Start.Before(other.End) && i.End.After(other.Start)
}

func (i Interval) IsFit(other Interval) bool {
	return i.Start.Compare(other.Start) <= 0 && i.End.Compare(other.End) >= 0
}

func intervalCompare(a, b Interval) int {
	return a.Start.Compare(b.Start)
}

func (i Intervals) SortByStart() {
	slices.SortFunc(i, intervalCompare)
}

func (i Intervals) IsSorted() bool {
	return slices.IsSortedFunc(i, intervalCompare)
}

// Note: Expected sorted slice
// TODO check it in debug mode?
func (intervals Intervals) HasOverlaps() bool {
	for i := 0; i < len(intervals)-1; i++ {
		if intervals[i].IsOverlap(intervals[i+1]) {
			return true
		}
	}
	return false
}

func (intervals Intervals) IsFit(other Interval) bool {
	for i := 0; i < len(intervals); i++ {
		if intervals[i].IsFit(other) {
			return true
		}
	}
	return false
}

func (intervals Intervals) IsOverlap(other Interval) bool {
	for i := 0; i < len(intervals); i++ {
		if intervals[i].IsOverlap(other) {
			return true
		}
	}
	return false
}
