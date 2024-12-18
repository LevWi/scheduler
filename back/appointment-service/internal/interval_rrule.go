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

type IntervalType int

const (
	Inclusion IntervalType = iota
	Exclusion
)

type IntervalRRuleWithType struct {
	Rule IntervalRRule
	Type IntervalType
}

func CalculateIntervals(in []IntervalRRuleWithType) Intervals {
	var inclusion Intervals
	var exclusion Intervals

	for _, el := range in {
		if el.Type == Exclusion {
			exclusion = append(exclusion, el.Rule.GetIntervals()...)
		} else if el.Type == Inclusion {
			inclusion = append(inclusion, el.Rule.GetIntervals()...)
		} else {
			panic("Unexpected value")
		}
	}

	PrepareUnited(inclusion)
	PrepareUnited(exclusion)

	return inclusion.PassedIntervals(exclusion)
}

func ConvertToIntervalRRuleWithType(jsonStrings []string) ([]IntervalRRuleWithType, error) {
	var intervalsRRules []IntervalRRuleWithType
	for _, el := range jsonStrings {
		var tmp IntervalRRuleWithType
		err := json.Unmarshal([]byte(el), &tmp)
		if err != nil {
			return nil, err
		}
		intervalsRRules = append(intervalsRRules, tmp)
	}
	return intervalsRRules, nil
}
