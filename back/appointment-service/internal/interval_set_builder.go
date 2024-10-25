package common

type IntervalsProducer interface {
	GetIntervals() Intervals
}

type GetIntervalsFunc func() Intervals

func (f GetIntervalsFunc) GetIntervals() Intervals {
	return f()
}

type IntervalsProducers []IntervalsProducer

func (p IntervalsProducers) GetIntervals() Intervals {
	var intervals Intervals
	for _, u := range p {
		intervals = append(intervals, u.GetIntervals()...)
	}
	return intervals
}

type IntervalSetBuilder struct {
	UnionGetters     []IntervalsProducer
	ExclusionGetters []IntervalsProducer
}

func (b IntervalSetBuilder) GetIntervalSet() (IntervalSet, error) {
	var unions Intervals
	for _, u := range b.UnionGetters {
		unions = append(unions, u.GetIntervals()...)
	}
	var exclusions Intervals
	for _, e := range b.ExclusionGetters {
		exclusions = append(exclusions, e.GetIntervals()...)
	}
	out := IntervalSet{
		Intervals: unions,
		Exclusion: exclusions,
	}
	err := out.PreparePayload()
	if err != nil {
		return IntervalSet{}, err
	}
	return out, nil
}
