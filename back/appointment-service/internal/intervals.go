package common

import "time"

func (s Slot) Interval() Interval {
	return Interval{Start: s.Start, End: s.Start.Add(time.Duration(s.Len) * time.Second)}
}

func (i Interval) IsOverlaps(other Interval) bool {
	return i.Start.Before(other.End) && i.End.After(other.Start)
}

func (i Interval) IsFit(other Interval) bool {
	return i.Start.Compare(other.Start) <= 0 && i.End.Compare(other.End) >= 0
}
