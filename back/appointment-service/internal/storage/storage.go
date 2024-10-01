package storage

import (
	"fmt"
	types "scheduler/appointment-service/internal"
	"time"

	"github.com/jmoiron/sqlx"
)

type dbSlot struct {
	Client string `db:"client_id"`
	//	Business string `db:"business_id"`
	Date int64 `db:"date_unx"`
	Len  int   `db:"len_sec"`
}

const createAppointmentsTable = `CREATE TABLE "appointments" (
	"date_unx"	  INTEGER NOT NULL,
	"business_id" TEXT NOT NULL,
	"client_id"	  TEXT NOT NULL,
	"len_sec"	  INTEGER NOT NULL CHECK(len_sec >= 5),
);`

func GetSlotInRange(db *sqlx.DB, business_id types.ID, start time.Time, end time.Time) ([]types.Slot, error) {
	var dbSlots []dbSlot
	err := db.Select(&dbSlots, "SELECT * FROM appointments WHERE business_id = $1 AND date_unx BETWEEN $2 AND $3", string(business_id), start.Unix(), end.Unix())
	if err != nil {
		return nil, err
	}

	var slotsOut []types.Slot
	for _, slot := range dbSlots {
		slotsOut = append(slotsOut, types.Slot{Client: types.ID(slot.Client),
			Start: time.Unix(slot.Date, 0),
			Len:   slot.Len})
	}
	return slotsOut, nil
}

func AddSlots(db *sqlx.DB, appointment types.Appointment) error {
	dbSlots := make([]dbSlot, 0, len(appointment.Slots))
	for _, slot := range appointment.Slots {
		//dbSlots = append(dbSlots, dbSlot{Client: string(slot.Client), Business: string(appointment.Business), Date: slot.Start.Unix(), Len: slot.Len})
		dbSlots = append(dbSlots, dbSlot{Client: string(slot.Client), Date: slot.Start.Unix(), Len: slot.Len})
	}

	q := fmt.Sprintf("INSERT INTO appointments (date_unx, business_id, client_id, len_sec) VALUES (:date_unx, %s, :client_id, :len_sec)", appointment.Business)
	_, err := db.NamedExec(q, dbSlots)
	return err
}
