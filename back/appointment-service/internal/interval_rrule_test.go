package common

import (
	"fmt"
	"testing"

	"github.com/teambition/rrule-go"
)

func TestIntervalRRuleMarshalStorage(t *testing.T) {
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
