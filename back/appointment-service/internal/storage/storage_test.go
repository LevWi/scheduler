package storage

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/mattn/go-sqlite3"

	types "scheduler/appointment-service/internal"
)

func TestStorage(t *testing.T) {
	dbPath := t.TempDir() + string(os.PathSeparator) + "test_.db"
	t.Log("db path:", dbPath)

	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = createTableAppointments(db)
	if err != nil {
		t.Fatal(err)
	}

	appointment := types.Appointment{
		Business: "b1", Slots: []types.Slot{
			{Client: "c1", Start: time.Now().Truncate(time.Minute), Len: 30},
			{Client: "c2", Start: time.Now().Truncate(time.Minute).Add(30 * time.Minute), Len: 30},
		},
	}

	err = AddSlots(db, appointment)
	if err != nil {
		t.Fatal(err)
	}

	slots, err := GetSlotInRange(db, appointment.Business, appointment.Slots[0].Start, appointment.Slots[1].Start)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(slots)

	for i, slot := range slots {
		if appointment.Slots[i].Start != slot.Start || appointment.Slots[i].Len != slot.Len || appointment.Slots[i].Client != slot.Client {
			fmt.Printf("slot mismatch : %+v != %+v", slot, appointment.Slots[i])
			t.FailNow()
		}
	}

	err = DeleteSlots(db, appointment.Business, appointment.Slots[0].Client, appointment.Slots[0].Start, appointment.Slots[0].Start)
	if err != nil {
		t.Fatal(err)
	}

	slots, err = GetSlotInRange(db, appointment.Business, appointment.Slots[0].Start, appointment.Slots[1].Start)
	if err != nil {
		t.Fatal(err)
	}

	if len(slots) != 1 {
		t.Fatal("expected 1 slot, got", len(slots))
	}

	if appointment.Slots[1].Start != slots[0].Start || appointment.Slots[1].Len != slots[0].Len || appointment.Slots[1].Client != slots[0].Client {
		fmt.Printf("slot mismatch : %+v != %+v", slots[0], appointment.Slots[1])
		t.FailNow()
	}
}
