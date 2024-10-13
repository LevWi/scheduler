package common

type IntervalSet struct {
	Union     Intervals
	Exclusion Intervals
}

func (s *IntervalSet) IsValid() bool {
	if !s.Union.IsSorted() {
		return false
	}

	if !s.Exclusion.IsSorted() {
		return false
	}

	return true
}
