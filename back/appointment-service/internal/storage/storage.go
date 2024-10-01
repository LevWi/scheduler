package storage

import (
	types "scheduler/appointment-service/internal"
	"time"

	"github.com/jmoiron/sqlx"
)

type dbSlot struct {
	Client string `db:"client_id"`
	Date   int64  `db:"date_unx"`
	Len    int    `db:"len_sec"`
}

const createAppointmentsTable = `CREATE TABLE "appointments" (
	"date_unx"	INTEGER NOT NULL,
	"business_id"	TEXT NOT NULL,
	"client_id"	TEXT NOT NULL,
	"len_sec"	INTEGER NOT NULL
);`

func GetSlotInRange(db *sqlx.DB, business_id types.ID, start time.Time, end time.Time) ([]types.Slot, error) {
	var dbSlots []dbSlot
	err := db.Select(&dbSlots, "SELECT * FROM appointments WHERE business_id = ? AND date_unx BETWEEN ? AND ?", string(business_id), start.Unix(), end.Unix())
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
