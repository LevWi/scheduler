package common

import (
	"time"

	"github.com/teambition/rrule-go"
)

type IntervalsProducer interface {
	GetIntervals() Intervals
}

type GetIntervalsFunc func() Intervals

func (f GetIntervalsFunc) GetIntervals() Intervals {
	return f()
}

type IntervalRRule struct {
	RRule *rrule.RRule
	Len   Seconds
}

type Seconds int64

func (d Seconds) Duration() time.Duration {
	return time.Duration(d) * time.Second
}

func (r IntervalRRule) GetIntervals() Intervals {
	next := r.RRule.Iterator()
	result := Intervals{}
	for {
		start, ok := next()
		if !ok {
			return result
		}
		result = append(result, Interval{Start: start, End: start.Add(r.Len.Duration())})
	}
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
		Union:     unions,
		Exclusion: exclusions,
	}
	err := out.PreparePayload()
	if err != nil {
		return IntervalSet{}, err
	}
	return out, nil
}
