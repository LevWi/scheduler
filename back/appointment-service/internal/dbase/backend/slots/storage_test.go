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

func TestSetBusinessSlotSettings(t *testing.T) {
	storage := TimeSlotsStorage{test.InitTmpDB(t)}
	defer storage.Close()

	err := storage.SetBusinessSlotSettings("b1", BusinessSlotSettings{DefaultChunk: 20 * time.Minute, MaxChunk: 40 * time.Minute})
	if err != nil {
		t.Fatal(err)
	}

	settings, err := storage.GetBusinessSlotSettings("b1")
	if err != nil {
		t.Fatal(err)
	}
	if settings.DefaultChunk != 20*time.Minute || settings.MaxChunk != 40*time.Minute {
		t.Fatalf("unexpected settings: %+v", settings)
	}

	err = storage.SetBusinessSlotSettings("b1", BusinessSlotSettings{DefaultChunk: 30 * time.Minute, MaxChunk: 60 * time.Minute})
	if err != nil {
		t.Fatal(err)
	}

	settings, err = storage.GetBusinessSlotSettings("b1")
	if err != nil {
		t.Fatal(err)
	}
	if settings.DefaultChunk != 30*time.Minute || settings.MaxChunk != 60*time.Minute {
		t.Fatalf("unexpected updated settings: %+v", settings)
	}
}

func TestBusinessSlotSettingsValidation(t *testing.T) {
	storage := TimeSlotsStorage{test.InitTmpDB(t)}
	defer storage.Close()

	if err := storage.SetBusinessSlotSettings("b1", BusinessSlotSettings{DefaultChunk: 4 * time.Minute, MaxChunk: 10 * time.Minute}); err == nil {
		t.Fatal("expected validation error for too small default chunk")
	}
	if err := storage.SetBusinessSlotSettings("b1", BusinessSlotSettings{DefaultChunk: 10 * time.Minute, MaxChunk: 4 * time.Minute}); err == nil {
		t.Fatal("expected validation error for too small max chunk")
	}
	if err := storage.SetBusinessSlotSettings("b1", BusinessSlotSettings{DefaultChunk: 30 * time.Minute, MaxChunk: 20 * time.Minute}); err == nil {
		t.Fatal("expected validation error for default > max")
	}

	_, err := storage.Exec(`INSERT INTO business_slot_settings (business_id, default_chunk_minutes, max_chunk_minutes) VALUES ($1, $2, $3)`, "b2", 3, 10)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := storage.GetBusinessSlotSettings("b2"); err == nil {
		t.Fatal("expected validation error on read")
	}
}

func TestGetCustomerAppointmentsInRange(t *testing.T) {
	storage := TimeSlotsStorage{test.InitTmpDB(t)}
	defer storage.Close()

	base := time.Now().Truncate(time.Minute)
	slotInProgress := common.Interval{Start: base.Add(-30 * time.Minute), End: base.Add(30 * time.Minute)}
	slotAfterStart := common.Interval{Start: base.Add(60 * time.Minute), End: base.Add(90 * time.Minute)}
	slotAfterAllRanges := common.Interval{Start: base.Add(180 * time.Minute), End: base.Add(210 * time.Minute)}

	err := storage.AddSlots(AddSlotsData{
		Business: "b1",
		Customer: "c1",
		Slots: common.Intervals{
			slotInProgress,
			slotAfterStart,
			slotAfterAllRanges,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = storage.AddSlots(AddSlotsData{
		Business: "b1",
		Customer: "c2",
		Slots: common.Intervals{
			{Start: base.Add(120 * time.Minute), End: base.Add(150 * time.Minute)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	type tc struct {
		name       string
		between    common.Interval
		want       common.Intervals
		wantCustID common.ID
	}

	tests := []tc{
		{
			name:       "open interval from now includes in-progress and future slots",
			between:    common.Interval{Start: base},
			want:       common.Intervals{slotInProgress, slotAfterStart, slotAfterAllRanges},
			wantCustID: "c1",
		},
		{
			name:       "bounded interval intersects only first slot",
			between:    common.Interval{Start: base, End: base.Add(45 * time.Minute)},
			want:       common.Intervals{slotInProgress},
			wantCustID: "c1",
		},
		{
			name:       "bounded interval intersects middle slot only",
			between:    common.Interval{Start: base.Add(45 * time.Minute), End: base.Add(100 * time.Minute)},
			want:       common.Intervals{slotAfterStart},
			wantCustID: "c1",
		},
		{
			name:       "bounded interval with no overlap returns empty",
			between:    common.Interval{Start: base.Add(100 * time.Minute), End: base.Add(120 * time.Minute)},
			want:       common.Intervals{},
			wantCustID: "c1",
		},
		{
			name:       "different customer is filtered out",
			between:    common.Interval{Start: base},
			want:       common.Intervals{},
			wantCustID: "unknown-customer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := storage.GetCustomerAppointmentsInRange("b1", tt.wantCustID, tt.between)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("unexpected slots count: got %d want %d", len(got), len(tt.want))
			}

			for i := range tt.want {
				if got[i].Customer != tt.wantCustID {
					t.Fatalf("unexpected customer id in result: got %q want %q", got[i].Customer, tt.wantCustID)
				}
				if got[i].Start != tt.want[i].Start || got[i].End != tt.want[i].End {
					t.Fatalf(
						"unexpected slot[%d]: got [%s - %s], want [%s - %s]",
						i,
						got[i].Start.Format(time.RFC3339),
						got[i].End.Format(time.RFC3339),
						tt.want[i].Start.Format(time.RFC3339),
						tt.want[i].End.Format(time.RFC3339),
					)
				}
			}
		})
	}
}
