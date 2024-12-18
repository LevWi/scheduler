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
	out, err := json.Marshal(IntervalRRuleWithType{Rule: rule, Type: Exclusion})
	if err != nil {
		t.Fatal(err)
	}

	jsonStrings := []string{string(out)}
	out2, err := ConvertToIntervalRRuleWithType(jsonStrings)
	if err != nil {
		t.Fatal(err)
	}

	if in.Rule.RRule.String() == out2[0].Rule.RRule.String() ||
		in.Type != out2[0].Type ||
		in.Rule.Len != out2[0].Rule.Len {
		fmt.Printf("%v != %v\n", in, out2)
		t.FailNow()
	}

	fmt.Printf("%v", out2)
}
