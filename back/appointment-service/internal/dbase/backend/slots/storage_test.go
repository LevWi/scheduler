package slots

import (
	"fmt"
	"math/rand"
	"slices"
	"testing"
	"time"

	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/dbase/test"

	"github.com/teambition/rrule-go"
)

func toInterval(start time.Time, end time.Time) common.Interval {
	return common.Interval{Start: start, End: end}
}

func TestStorage(t *testing.T) {
	storage := TimeSlotsStorage{test.InitTmpDB(t)}
	defer storage.Close()
	var appointments []AddSlotsData
	{
		tm := time.Now().Truncate(time.Minute)
		appointments = append(appointments, AddSlotsData{
			Business: "b1",
			Customer: "c1",
			Slots: []common.Interval{
				{Start: tm, End: tm.Add(30 * time.Minute)}},
		}, AddSlotsData{
			Business: "b1",
			Customer: "c2",
			Slots: []common.Interval{
				{Start: tm.Add(30 * time.Minute), End: tm.Add(60 * time.Minute)}},
		})
	}

	for _, el := range appointments {
		err := storage.AddSlots(el)
		if err != nil {
			t.Fatal(err)
		}

		//Check duplicates
		err = storage.AddSlots(el)
		if err == nil {
			t.Fatal(err)
		}
	}

	slots, err := storage.GetBusySlotsInRange(appointments[0].Business, toInterval(appointments[0].Slots[0].Start, appointments[1].Slots[0].Start))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(slots)

	for i, slot := range slots {
		if appointments[i].Slots[0].Start != slot.Start || appointments[i].Slots[0].End != slot.End || appointments[i].Customer != slot.Customer {
			fmt.Printf("slot mismatch : %+v != %+v", slot, appointments[i].Slots[0])
			t.FailNow()
		}
	}

	err = storage.DeleteSlots(appointments[0].Business, appointments[0].Customer, appointments[0].Slots[0].Start, appointments[0].Slots[0].Start)
	if err != nil {
		t.Fatal(err)
	}

	slots, err = storage.GetBusySlotsInRange(appointments[0].Business, toInterval(appointments[0].Slots[0].Start, appointments[1].Slots[0].Start))
	if err != nil {
		t.Fatal(err)
	}

	if len(slots) != 1 {
		t.Fatal("expected 1 slot, got", len(slots))
	}

	if appointments[1].Slots[0].Start != slots[0].Start || appointments[1].Slots[0].End != slots[0].End || appointments[1].Customer != slots[0].Customer {
		fmt.Printf("slot mismatch : %+v != %+v", slots, appointments[1])
		t.FailNow()
	}
}

func TestStorageBusinessRule(t *testing.T) {
	storage := TimeSlotsStorage{test.InitTmpDB(t)}
	defer storage.Close()

	const businessName = "Business Name"
	const rulesCount = 100

	rules := make([]DbBusinessRule, rulesCount)

	for x := range rulesCount {
		randI := rand.Int()
		ruleType := common.Exclusion
		if randI%2 == 1 {
			ruleType = common.Inclusion
		}

		rr, err := rrule.NewRRule(
			rrule.ROption{
				Dtstart: time.Now(),
				Freq:    rrule.DAILY,
				Count:   randI % 100,
			})

		if err != nil {
			t.Fatal(err)
		}

		r := DbBusinessRule{
			Rule: common.IntervalRRuleWithType{
				Rule: common.IntervalRRule{
					RRule: rr,
					Len:   common.Seconds(randI % (60 * 60)),
				},
				Type: ruleType,
			},
		}

		ruleID, err := storage.AddBusinessRule(businessName, r.Rule)
		if err != nil {
			t.Fatal(err)
		}
		r.Id = ruleID

		rules[x] = r
	}

	var count int
	err := storage.Get(&count, "SELECT COUNT(*) FROM business_work_rule")
	if err != nil {
		t.Fatal(err)
	}
	if count != rulesCount {
		t.Fatal("unexpect count: ", count)
	}

	rulesGet, err := storage.GetBusinessRules(businessName)
	if err != nil {
		t.Fatal(err)
	}
	if len(rulesGet) != rulesCount {
		t.Fatal("unexpect rules count: ", len(rulesGet))
	}

	for i, r := range rules {
		if r.Id != rulesGet[i].Id || !r.Rule.Equal(rulesGet[i].Rule) {
			t.Fatalf("unexpect rule[%d]: %v", i, rulesGet[i])
		}
	}

	const index = 42
	deletedRule := rules[index]
	err = storage.DeleteBusinessRule(businessName, deletedRule.Id)
	if err != nil {
		t.Fatal(err)
	}

	rulesGet, err = storage.GetBusinessRules(businessName)
	if err != nil {
		t.Fatal(err)
	}
	if len(rulesGet) != rulesCount-1 {
		t.Fatal("unexpect rules count: ", len(rules))
	}

	if slices.ContainsFunc(rulesGet, func(r DbBusinessRule) bool {
		return r.Rule == deletedRule.Rule || r.Id == deletedRule.Id
	}) {
		t.Fatal("rule should be deleted")
	}
}

func TestGetBusinessSlotSettings(t *testing.T) {
	storage := TimeSlotsStorage{test.InitTmpDB(t)}
	defer storage.Close()

	_, err := storage.Exec(`INSERT INTO business_slot_settings (business_id, default_chunk_minutes, max_chunk_minutes) VALUES ($1, $2, $3)`, "b1", 25, 50)
	if err != nil {
		t.Fatal(err)
	}

	settings, err := storage.GetBusinessSlotSettings("b1")
	if err != nil {
		t.Fatal(err)
	}

	if settings.DefaultChunk != 25*time.Minute {
		t.Fatalf("unexpected default chunk: %v", settings.DefaultChunk)
	}
	if settings.MaxChunk != 50*time.Minute {
		t.Fatalf("unexpected max chunk: %v", settings.MaxChunk)
	}
}
