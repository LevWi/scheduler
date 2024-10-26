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

type SliceIndex = int

func (i Interval) IsValid() bool {
	return i.Start.Before(i.End)
}

func (i Interval) IsOverlap(other Interval) bool {
	return i.Start.Before(other.End) && i.End.After(other.Start)
}

func (i Interval) Before(other Interval) bool {
	return i.End.Compare(other.Start) <= 0
}

func (i Interval) IsFit(other Interval) bool {
	return i.Start.Compare(other.Start) <= 0 && i.End.Compare(other.End) >= 0
}

func (i Interval) Subtract(other Interval) Intervals {
	if !i.IsOverlap(other) {
		return Intervals{i}
	}

	if other.IsFit(i) {
		return nil
	}

	if i.Start.Before(other.Start) {
		// right hand crossing
		if i.End.Compare(other.End) <= 0 {
			return Intervals{{Start: i.Start, End: other.Start}}
		}
		// middle crossing
		return Intervals{
			{Start: i.Start, End: other.Start},
			{Start: other.End, End: i.End},
		}
	}
	// left hand crossing
	return Intervals{{Start: other.End, End: i.End}}
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

func (intervals Intervals) FirstOverlapped(other Interval) SliceIndex {
	for i := 0; i < len(intervals); i++ {
		if intervals[i].IsOverlap(other) {
			return i
		}
	}
	return -1
}

func (intervals Intervals) Copy() Intervals {
	out := make(Intervals, len(intervals))
	copy(out, intervals)
	return out
}

// Note: Expected sorted slice
func (intervals *Intervals) Unite() {
	*intervals = unions(*intervals)
}

func unions(intervals Intervals) Intervals {
	i, j := 0, 1
	for ; j < len(intervals); j++ {
		if intervals[i].IsOverlap(intervals[j]) {
			if intervals[i].End.Compare(intervals[j].End) < 0 {
				intervals[i].End = intervals[j].End
			}
		} else {
			i++
		}
	}
	return intervals[: len(intervals)-i : cap(intervals)]
}
