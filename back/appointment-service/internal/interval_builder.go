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
