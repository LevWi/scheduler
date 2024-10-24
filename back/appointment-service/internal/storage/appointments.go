package storage

import (
	common "scheduler/appointment-service/internal"
	"time"

	"github.com/jmoiron/sqlx"
)

type Storage struct {
	*sqlx.DB
}

type dbSlot struct {
	Client    string `db:"client_id"`
	Business  string `db:"business_id"` // TODO use integer
	DateStart int64  `db:"date_start"`
	DateEnd   int64  `db:"date_end"`
}

const queryCreateAppointmentsTable = `CREATE TABLE "appointments" (
	"date_start"	  INTEGER NOT NULL,
	"date_end"	  INTEGER NOT NULL,
	"business_id" TEXT NOT NULL,
	"client_id"	  TEXT NOT NULL,
	UNIQUE (business_id, date_start)
);`

func CreateTableAppointments(db *Storage) error {
	_, err := db.Exec(queryCreateAppointmentsTable)
	if err != nil {
		return err
	}
	return nil
}

// func (db *Storage) GetAvailableSlotsInRange(business_id common.ID, start time.Time, end time.Time) ([]common.Slot, error) {
// 	panic("Not implemented")
// }

func (db *Storage) GetBusySlotsInRange(business_id common.ID, start time.Time, end time.Time) ([]common.Slot, error) {
	var dbSlots []dbSlot
	err := db.Select(&dbSlots, "SELECT * FROM appointments WHERE business_id = $1 AND date_start BETWEEN $2 AND $3",
		string(business_id), start.Unix(), end.Unix())
	if err != nil {
		return nil, err
	}

	var slotsOut []common.Slot
	for _, slot := range dbSlots {
		slotsOut = append(slotsOut,
			common.Slot{Client: slot.Client,
				Interval: common.Interval{
					Start: time.Unix(slot.DateStart, 0),
					End:   time.Unix(slot.DateEnd, 0),
				}})
	}
	return slotsOut, nil
}

func (db *Storage) DeleteSlots(business_id common.ID, client_id common.ID, start time.Time, end time.Time) error {
	_, err := db.Exec("DELETE FROM appointments WHERE business_id = $1 AND client_id = $2 AND date_start BETWEEN $3 AND $4",
		business_id, client_id, start.Unix(), end.Unix())
	return err
}

// TODO check intersections in range

func (db *Storage) AddSlots(appointment common.Appointment) error {
	dbSlots := make([]dbSlot, 0, len(appointment.Slots))
	for _, slot := range appointment.Slots {
		dbSlots = append(dbSlots, dbSlot{Client: slot.Client, Business: appointment.Business, DateStart: slot.Interval.Start.Unix(), DateEnd: slot.Interval.End.Unix()})
	}
	_, err := db.NamedExec("INSERT INTO appointments (business_id, date_start, client_id, date_end) VALUES (:business_id, :date_start, :client_id, :date_end)", dbSlots)
	return err
}
