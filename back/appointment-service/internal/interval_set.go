package common

import "errors"

type IntervalSet struct {
	Union     Intervals
	Exclusion Intervals
}

func (s *IntervalSet) IsValid() bool {
	if !s.Union.IsSorted() {
		return false
	}

	if s.Union.HasOverlaps() {
		return false
	}

	if !s.Exclusion.IsSorted() {
		return false
	}

	return true
}

func NewIntervalSetWithCopies(union Intervals, exclusion Intervals) (IntervalSet, error) {
	tmp := IntervalSet{}
	tmp.Union = make(Intervals, len(union))
	tmp.Exclusion = make(Intervals, len(exclusion))
	copy(tmp.Union, union)
	copy(tmp.Exclusion, exclusion)
	err := tmp.PreparePayload()
	if err != nil {
		return IntervalSet{}, err
	}
	return tmp, nil
}

func (s *IntervalSet) PreparePayload() error {
	s.Union.SortByStart()
	if s.Union.HasOverlaps() {
		return errors.New("bad union")
	}
	s.Exclusion.SortByStart()
	return nil
}

func (s *IntervalSet) IsFit(other Interval) bool {
	return s.Union.IsFit(other) && !s.Exclusion.IsOverlap(other)
}

func (s *IntervalSet) AddExclusion(other Interval) {
	s.Exclusion = append(s.Exclusion, other)
	s.Exclusion.SortByStart()
}

func (s *IntervalSet) AddUnion(other Interval) error {
	if s.Union.IsOverlap(other) {
		return errors.New("overlap detected")
	}
	s.Union = append(s.Union, other)
	s.Union.SortByStart()
	return nil
}
