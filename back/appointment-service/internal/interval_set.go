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
		return errors.New("bad union")
	}
	s.Exclusion.SortByStart()
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
