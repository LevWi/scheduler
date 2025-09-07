package common

import (
	"encoding/json"
	"fmt"
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

type IntervalType string

const (
	Inclusion IntervalType = "inclusion"
	Exclusion IntervalType = "exclusion"
)

func (i IntervalType) isValid() bool {
	switch i {
	case Inclusion:
		fallthrough
	case Exclusion:
		return true
	}
	return false
}

func (i IntervalType) MarshalJSON() ([]byte, error) {
	if !i.isValid() {
		return nil, fmt.Errorf("IntervalType: wrong value %v", i)
	}
	return json.Marshal(string(i))
}

func (i *IntervalType) UnmarshalJSON(in []byte) error {
	var v string
	err := json.Unmarshal(in, &v)
	if err != nil {
		return err
	}

	if !IntervalType(v).isValid() {
		return fmt.Errorf("IntervalType: wrong value %v", v)
	}

	*i = IntervalType(v)
	return nil
}

type IntervalRRuleWithType struct {
	Rule IntervalRRule
	Type IntervalType
}

func CalculateIntervals(in []IntervalRRuleWithType) Intervals {
	var inclusion Intervals
	var exclusion Intervals

	for _, el := range in {
		switch el.Type {
		case Exclusion:
			exclusion = append(exclusion, el.Rule.GetIntervals()...)
		case Inclusion:
			inclusion = append(inclusion, el.Rule.GetIntervals()...)
		default:
			panic("Unexpected value")
		}
	}

	PrepareUnited(inclusion)
	PrepareUnited(exclusion)

	return inclusion.PassedIntervals(exclusion)
}

// TODO remove?
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
