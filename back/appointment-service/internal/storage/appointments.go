package storage

import (
	types "scheduler/appointment-service/internal"
	"time"

	"github.com/jmoiron/sqlx"
)

type Storage struct {
	*sqlx.DB
}

type dbSlot struct {
	Client   string `db:"client_id"`
	Business string `db:"business_id"` // TODO use integer
	Date     int64  `db:"date_unx"`
	Len      int    `db:"len_sec"`
}

const queryCreateAppointmentsTable = `CREATE TABLE "appointments" (
	"date_unx"	  INTEGER NOT NULL,
	"business_id" TEXT NOT NULL,
	"client_id"	  TEXT NOT NULL,
	"len_sec"	  INTEGER NOT NULL CHECK(len_sec >= 5)
);`

func CreateTableAppointments(db *Storage) error {
	_, err := db.Exec(queryCreateAppointmentsTable)
	if err != nil {
		return err
	}
	return nil
}

func (db *Storage) GetSlotInRange(business_id types.ID, start time.Time, end time.Time) ([]types.Slot, error) {
	var dbSlots []dbSlot
	err := db.Select(&dbSlots, "SELECT * FROM appointments WHERE business_id = $1 AND date_unx BETWEEN $2 AND $3",
		string(business_id), start.Unix(), end.Unix())
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

func (db *Storage) DeleteSlots(business_id types.ID, client_id types.ID, start time.Time, end time.Time) error {
	_, err := db.Exec("DELETE FROM appointments WHERE business_id = $1 AND client_id = $2 AND date_unx BETWEEN $3 AND $4",
		string(business_id), string(client_id), start.Unix(), end.Unix())
	return err
}

// TODO check intersections in range

func (db *Storage) AddSlots(appointment types.Appointment) error {
	dbSlots := make([]dbSlot, 0, len(appointment.Slots))
	for _, slot := range appointment.Slots {
		dbSlots = append(dbSlots, dbSlot{Client: string(slot.Client), Business: string(appointment.Business), Date: slot.Start.Unix(), Len: slot.Len})
	}

	_, err := db.NamedExec("INSERT INTO appointments (business_id, date_unx, client_id, len_sec) VALUES (:business_id, :date_unx, :client_id, :len_sec)", dbSlots)
	return err
}
