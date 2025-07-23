package storage

import (
	common "scheduler/appointment-service/internal"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	*sqlx.DB
}

type dbBusySlot struct {
	Client    string `db:"client_id"`
	Business  string `db:"business_id"` // TODO use integer
	DateStart int64  `db:"date_start"`
	DateEnd   int64  `db:"date_end"`
}

func (slot dbBusySlot) ToSlot() common.BusySlot {
	return common.BusySlot{Client: slot.Client,
		Interval: common.Interval{
			Start: time.Unix(slot.DateStart, 0),
			End:   time.Unix(slot.DateEnd, 0),
		}}
}

type DBRule = string

type DbBusinessRule struct {
	Id   string
	Rule DBRule
}

// TODO IF NOT EXISTS
const queryCreateAppointmentsTable = `CREATE TABLE "appointments" (
	"date_start"	  INTEGER NOT NULL,
	"date_end"	  INTEGER NOT NULL,
	"business_id" TEXT NOT NULL,
	"client_id"	  TEXT NOT NULL,
	UNIQUE (business_id, date_start)
);`

// TODO add exclude type rule
const createBusinessTable = `CREATE TABLE "business_work_rule" (
	"id" TEXT NOT NULL,
	"business_id" TEXT NOT NULL,
	"rule"	  TEXT NOT NULL,
	UNIQUE (id, business_id)
);`

func CreateTableAppointments(db *Storage) error {
	_, err := db.Exec(queryCreateAppointmentsTable)
	if err != nil {
		return err
	}
	return nil
}

func CreateBusinessTable(db *Storage) error {
	_, err := db.Exec(createBusinessTable)
	if err != nil {
		return err
	}
	return nil
}

func (db *Storage) GetBusinessRules(business_id common.ID) ([]DbBusinessRule, error) {
	var rules []DbBusinessRule
	err := db.Select(&rules, "SELECT id, rule FROM business_work_rule WHERE business_id = $1",
		string(business_id))
	if err != nil {
		return nil, err
	}

	return rules, nil
}

// TODO check that rule valid ?
// TODO return business rule id ?
func (db *Storage) AddBusinessRule(businessID string, rule DBRule) error {
	newID := uuid.New().String()

	_, err := db.Exec(`
		INSERT INTO business_work_rule (id, business_id, rule)
		VALUES ($1, $2, $3)
	`, newID, businessID, rule)
	return err
}

// TODO is value deletion confirm needed?
// TODO Do not delete permanently
func (db *Storage) DeleteBusinessRule(business_id common.ID, id string) error {
	_, err := db.Exec("DELETE FROM business_work_rule WHERE business_id = $1 AND id = $2",
		business_id, id)
	return err
}

// TODO add test
func (db *Storage) GetAvailableSlotsInRange(business_id common.ID, between common.Interval) (common.Intervals, error) {
	var rules []DBRule
	err := db.Select(&rules, "SELECT rule FROM business_work_rule WHERE business_id = $1", string(business_id))
	if err != nil {
		return nil, err
	}

	intervalsRRules, err := common.ConvertToIntervalRRuleWithType(rules)
	if err != nil {
		return nil, err
	}

	// TODO Not optimal
	intervals := common.CalculateIntervals(intervalsRRules)
	intervals = intervals.UnitedBetween(between)
	if len(intervals) == 0 {
		return nil, nil
	}

	slots, err := db.GetBusySlotsInRange(business_id, between)
	if err != nil {
		return nil, err
	}

	var exclusions common.Intervals
	for _, slot := range slots {
		exclusions = append(exclusions, slot.Interval)
	}

	intervals = intervals.PassedIntervals(exclusions)
	return intervals, nil
}

func (db *Storage) GetBusySlotsInRange(business_id common.ID, between common.Interval) ([]common.BusySlot, error) {
	var dbSlots []dbBusySlot
	err := db.Select(&dbSlots, "SELECT * FROM appointments WHERE business_id = $1 AND date_start BETWEEN $2 AND $3",
		string(business_id), between.Start.Unix(), between.End.Unix())
	if err != nil {
		return nil, err
	}

	var slotsOut []common.BusySlot
	for _, dbSlot := range dbSlots {
		slotsOut = append(slotsOut, dbSlot.ToSlot())
	}
	return slotsOut, nil
}

func (db *Storage) DeleteSlots(business_id common.ID, client_id common.ID, start time.Time, end time.Time) error {
	_, err := db.Exec("DELETE FROM appointments WHERE business_id = $1 AND client_id = $2 AND date_start BETWEEN $3 AND $4",
		business_id, client_id, start.Unix(), end.Unix())
	return err
}

type AddSlotsData struct {
	Business common.ID
	Client   common.ID
	Slots    common.Intervals
}

// expected that no intersections in range
func (db *Storage) AddSlots(in AddSlotsData) error {
	dbSlots := make([]dbBusySlot, 0, len(in.Slots))
	for _, slot := range in.Slots {
		dbSlots = append(dbSlots, dbBusySlot{Client: in.Client, Business: in.Business, DateStart: slot.Start.Unix(), DateEnd: slot.End.Unix()})
	}
	_, err := db.NamedExec("INSERT INTO appointments (business_id, date_start, client_id, date_end) VALUES (:business_id, :date_start, :client_id, :date_end)", dbSlots)
	return err
}
