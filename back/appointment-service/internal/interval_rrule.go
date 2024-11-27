package common

import (
	"encoding/json"
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

type intervalRRuleJsonAdapter struct {
	RRule string
	Len   int64
}

func (r IntervalRRule) MarshalJSON() ([]byte, error) {
	var tmp intervalRRuleJsonAdapter
	tmp.RRule = r.RRule.String()
	tmp.Len = int64(r.Len)
	return json.Marshal(tmp)
}

// TODO check utc time
func (r *IntervalRRule) UnmarshalJSON(b []byte) error {
	var tmp intervalRRuleJsonAdapter
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	rule, err := rrule.StrToRRule(tmp.RRule)
	if err != nil {
		return err
	}

	r.RRule = rule
	r.Len = Seconds(tmp.Len)
	return nil
}
