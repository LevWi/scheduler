package common

import (
	"time"

	"github.com/teambition/rrule-go"
)

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
