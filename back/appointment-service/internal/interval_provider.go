package common

import (
	"time"

	"github.com/teambition/rrule-go"
)

type IntervalSetProvider interface {
	GetIntervalSet() (IntervalSet, error)
}

type IntervalsProvider interface {
	GetIntervals() Intervals
}

type GetIntervalSetFunc func() (IntervalSet, error)
type GetIntervalsFunc func() Intervals

func (f GetIntervalSetFunc) GetIntervalSet() (IntervalSet, error) {
	return f()
}

func (f GetIntervalsFunc) GetIntervals() Intervals {
	return f()
}

type IntervalRRule struct {
	RRule    *rrule.RRule
	Duration Seconds
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
		result = append(result, Interval{Start: start, End: start.Add(r.Duration.Duration())})
	}
}
