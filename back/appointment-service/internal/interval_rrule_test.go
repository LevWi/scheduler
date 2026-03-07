package common

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/teambition/rrule-go"
)

func getRule(t *testing.T) IntervalRRule {
	const rruleStr = "DTSTART=20060101T150405Z;FREQ=DAILY;COUNT=5"
	var rule IntervalRRule
	{
		r, e := rrule.StrToRRule(rruleStr)
		if e != nil {
			t.Fatal(e)
		}
		rule.RRule = r
		rule.Len = 5
	}
	return rule
}

func TestIntervalRRuleMarshal(t *testing.T) {
	rule := getRule(t)
	ruleJson, err := rule.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	rule2 := IntervalRRule{}
	err = rule2.UnmarshalJSON(ruleJson)
	if err != nil {
		t.Fatal(err)
	}

	inrevals := rule.GetIntervals()
	inrevals2 := rule2.GetIntervals()
	if len(inrevals) != len(inrevals2) || len(inrevals) != 5 {
		t.Fatal("unexpected intervals count")
	}

	for i := 0; i < len(inrevals); i++ {
		if inrevals[i].Start != inrevals2[i].Start || inrevals[i].End != inrevals2[i].End {
			fmt.Printf("%v != %v\n", inrevals[i], inrevals2[i])
			t.FailNow()
		}
	}

}

func TestIntervalRRuleWithTypeMarshal(t *testing.T) {
	rule := getRule(t)
	in := IntervalRRuleWithType{Rule: rule, Type: Exclusion}
	out, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	var tmp IntervalRRuleWithType
	err = json.Unmarshal(out, &tmp)
	if err != nil {
		t.Fatal(err)
	}

	if !in.Equal(tmp) {
		t.Fatalf("%v != %v\n", in, tmp)
	}
}

func mustRRule(t *testing.T, rule string) *rrule.RRule {
	r, err := rrule.StrToRRule(rule)
	if err != nil {
		t.Fatalf("failed to parse rrule: %v", err)
	}
	return r
}

func TestIntervalRRuleEqual(t *testing.T) {
	rule1 := mustRRule(t, "FREQ=DAILY")
	rule2 := mustRRule(t, "FREQ=DAILY")
	rule3 := mustRRule(t, "FREQ=WEEKLY")

	tests := []struct {
		name string
		a    IntervalRRule
		b    IntervalRRule
		want bool
	}{
		{
			name: "both nil rules equal",
			a:    IntervalRRule{RRule: nil, Len: 60},
			b:    IntervalRRule{RRule: nil, Len: 60},
			want: true,
		},
		{
			name: "one nil rule not equal",
			a:    IntervalRRule{RRule: nil, Len: 60},
			b:    IntervalRRule{RRule: rule1, Len: 60},
			want: false,
		},
		{
			name: "same rules equal",
			a:    IntervalRRule{RRule: rule1, Len: 60},
			b:    IntervalRRule{RRule: rule2, Len: 60},
			want: true,
		},
		{
			name: "different rules not equal",
			a:    IntervalRRule{RRule: rule1, Len: 60},
			b:    IntervalRRule{RRule: rule3, Len: 60},
			want: false,
		},
		{
			name: "different length not equal",
			a:    IntervalRRule{RRule: rule1, Len: 60},
			b:    IntervalRRule{RRule: rule1, Len: 120},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.Equal(tt.b)
			if got != tt.want {
				t.Fatalf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}
