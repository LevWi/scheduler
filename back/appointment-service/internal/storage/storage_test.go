package storage

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/mattn/go-sqlite3"

	types "scheduler/appointment-service/internal"
)

func initDB(t *testing.T) Storage {
	dbPath := t.TempDir() + string(os.PathSeparator) + "test_.db"
	t.Log("db path:", dbPath)

	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	return Storage{DB: db}
}

func toInterval(start time.Time, end time.Time) types.Interval {
	return types.Interval{Start: start, End: end}
}

func TestStorage(t *testing.T) {
	storage := initDB(t)
	defer storage.Close()

	err := CreateTableAppointments(&storage)
	if err != nil {
		t.Fatal(err)
	}

	var appointments []AddSlotsData
	{
		tm := time.Now().Truncate(time.Minute)
		appointments = append(appointments, AddSlotsData{
			Business: "b1",
			Client:   "c1",
			Slots: []types.Interval{
				{Start: tm, End: tm.Add(30 * time.Minute)}},
		}, AddSlotsData{
			Business: "b1",
			Client:   "c2",
			Slots: []types.Interval{
				{Start: tm.Add(30 * time.Minute), End: tm.Add(60 * time.Minute)}},
		})
	}

	for _, el := range appointments {
		err = storage.AddSlots(el)
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
		if appointments[i].Slots[0].Start != slot.Start || appointments[i].Slots[0].End != slot.End || appointments[i].Client != slot.Client {
			fmt.Printf("slot mismatch : %+v != %+v", slot, appointments[i].Slots[0])
			t.FailNow()
		}
	}

	err = storage.DeleteSlots(appointments[0].Business, appointments[0].Client, appointments[0].Slots[0].Start, appointments[0].Slots[0].Start)
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

	if appointments[1].Slots[0].Start != slots[0].Start || appointments[1].Slots[0].End != slots[0].End || appointments[1].Client != slots[0].Client {
		fmt.Printf("slot mismatch : %+v != %+v", slots, appointments[1])
		t.FailNow()
	}
}

func TestStorageBusinessRule(t *testing.T) {
	storage := initDB(t)
	defer storage.Close()

	err := CreateBusinessTable(&storage)
	if err != nil {
		t.Fatal(err)
	}

	const businessName = "Business Name"
	const rulesCount = 100
	for x := range rulesCount {
		err = storage.AddBusinessRule(businessName, "some_rule"+strconv.Itoa(x))
		if err != nil {
			t.Fatal(err)
		}
	}

	var count int
	err = storage.Get(&count, "SELECT COUNT(*) FROM business_work_rule WHERE rule = \"some_rule42\"")
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("unexpect count: ", count)
	}

	rules, err := storage.GetBusinessRules(businessName)
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != rulesCount {
		t.Fatal("unexpect rules count: ", len(rules))
	}

	for i, r := range rules {
		if r.Rule != "some_rule"+strconv.Itoa(i) {
			t.Fatal("unexpect rule: ", r.Rule)
		}
	}

	err = storage.DeleteBusinessRule(businessName, rules[42].Id)
	if err != nil {
		t.Fatal(err)
	}

	rules, err = storage.GetBusinessRules(businessName)
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != rulesCount-1 {
		t.Fatal("unexpect rules count: ", len(rules))
	}

	if slices.ContainsFunc(rules, func(r DbBusinessRule) bool { return r.Rule == "some_rule42" }) {
		t.Fatal("rule should be deleted")
	}
}
