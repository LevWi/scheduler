package common

import "errors"

type IntervalSet struct {
	Intervals Intervals
	Exclusion Intervals
}

func (s *IntervalSet) IsValid() bool {
	if !s.Intervals.IsSorted() {
		return false
	}

	if s.Intervals.HasOverlaps() {
		return false
	}

	if !s.Exclusion.IsSorted() {
		return false
	}

	// TODO check
	return true
}

func NewIntervalSetWithCopies(union Intervals, exclusion Intervals) (IntervalSet, error) {
	tmp := IntervalSet{}
	tmp.Intervals = make(Intervals, len(union))
	tmp.Exclusion = make(Intervals, len(exclusion))
	copy(tmp.Intervals, union)
	copy(tmp.Exclusion, exclusion)
	err := tmp.PreparePayload()
	if err != nil {
		return IntervalSet{}, err
	}
	return tmp, nil
}

func (s *IntervalSet) PreparePayload() error {
	s.Intervals.SortByStart()
	if s.Intervals.HasOverlaps() {
		return errors.New("bad Intervals")
	}
	s.Exclusion.SortByStart()
	if s.Exclusion.HasOverlaps() {
		return errors.New("bad Exclusion")
	}
	return nil
}

func (s *IntervalSet) IsFit(other Interval) bool {
	return s.Intervals.IsFit(other) && !s.Exclusion.IsOverlap(other)
}

func (s *IntervalSet) AddExclusion(other Interval) {
	s.Exclusion = append(s.Exclusion, other)
	s.Exclusion.SortByStart()
}

func (s *IntervalSet) Add(other Interval) error {
	if s.Intervals.IsOverlap(other) {
		return errors.New("overlap detected")
	}
	s.Intervals = append(s.Intervals, other)
	s.Intervals.SortByStart()
	return nil
}

// Expected valid IntervalSet and Exclusion.Unite()
func (s *IntervalSet) PassedIntervals() Intervals {
	intervals := s.Intervals
	if len(intervals) == 0 {
		return Intervals{}
	}

	if len(s.Exclusion) == 0 {
		return intervals.Copy()
	}

	out := make(Intervals, 0, len(intervals))
	tmp := Interval{}
	exclusionIndex := 0
	nextIntervalIndex := 0
	for (nextIntervalIndex < len(intervals) || tmp.IsValid()) && exclusionIndex < len(s.Exclusion) {
		if !tmp.IsValid() {
			tmp = intervals[nextIntervalIndex]
			nextIntervalIndex++
		}
		if tmp.Before(s.Exclusion[exclusionIndex]) {
			out = append(out, tmp)
			tmp = Interval{}
		} else if s.Exclusion[exclusionIndex].Before(tmp) {
			exclusionIndex++
		} else if tmp.IsOverlap(s.Exclusion[exclusionIndex]) {
			result := tmp.Subtract(s.Exclusion[exclusionIndex])
			switch len(result) {
			case 2:
				out = append(out, result[0])
				tmp = result[1]
			case 1:
				if result[0].Before(s.Exclusion[exclusionIndex]) {
					out = append(out, result[0])
					tmp = Interval{}
				} else {
					tmp = result[0]
				}
			case 0:
				tmp = Interval{}
			default:
				panic("Subtract: Unexpected behavior")
			}
		} else {
			panic("IsOverlap: Unexpected behavior")
		}
	}

	if tmp.IsValid() {
		out = append(out, tmp)
	}

	if nextIntervalIndex < len(intervals) {
		out = append(out, intervals[nextIntervalIndex:]...)
	}
	return out
}
