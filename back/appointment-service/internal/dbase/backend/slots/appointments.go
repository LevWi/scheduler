package slots

import (
	common "scheduler/appointment-service/internal"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type TimeSlotsStorage struct {
	*sqlx.DB
}

type dbBusySlot struct {
	Customer  string `db:"client_id"`
	Business  string `db:"business_id"` // TODO use integer
	DateStart int64  `db:"date_start"`
	DateEnd   int64  `db:"date_end"`
}

func (slot dbBusySlot) ToSlot() common.BusySlot {
	return common.BusySlot{
		Customer: slot.Customer,
		Interval: common.Interval{
			Start: time.Unix(slot.DateStart, 0),
			End:   time.Unix(slot.DateEnd, 0),
		},
	}
}

type DBRule = string

type DbBusinessRule struct {
	Id   string
	Rule DBRule
}

func (db *TimeSlotsStorage) GetBusinessRules(business_id common.ID) ([]DbBusinessRule, error) {
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
func (db *TimeSlotsStorage) AddBusinessRule(businessID string, rule DBRule) error {
	newID := uuid.New().String()

	_, err := db.Exec(`
		INSERT INTO business_work_rule (id, business_id, rule)
		VALUES ($1, $2, $3)
	`, newID, businessID, rule)
	return err
}

// TODO is value deletion confirm needed?
// TODO Do not delete permanently
func (db *TimeSlotsStorage) DeleteBusinessRule(business_id common.ID, id string) error {
	_, err := db.Exec("DELETE FROM business_work_rule WHERE business_id = $1 AND id = $2",
		business_id, id)
	return err
}

// TODO add test
func (db *TimeSlotsStorage) GetAvailableSlotsInRange(business_id common.ID, between common.Interval) (common.Intervals, error) {
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

func (db *TimeSlotsStorage) GetBusySlotsInRange(business_id common.ID, between common.Interval) ([]common.BusySlot, error) {
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

func (db *TimeSlotsStorage) DeleteSlots(business_id common.ID, customerID common.ID, start time.Time, end time.Time) error {
	_, err := db.Exec("DELETE FROM appointments WHERE business_id = $1 AND client_id = $2 AND date_start BETWEEN $3 AND $4",
		business_id, customerID, start.Unix(), end.Unix())
	return err
}

type AddSlotsData struct {
	Business common.ID
	Customer common.ID
	Slots    common.Intervals
}

// expected that no intersections in range
func (db *TimeSlotsStorage) AddSlots(in AddSlotsData) error {
	dbSlots := make([]dbBusySlot, 0, len(in.Slots))
	for _, slot := range in.Slots {
		dbSlots = append(dbSlots, dbBusySlot{
			Customer:  in.Customer,
			Business:  in.Business,
			DateStart: slot.Start.Unix(),
			DateEnd:   slot.End.Unix(),
		})
	}
	_, err := db.NamedExec("INSERT INTO appointments (business_id, date_start, client_id, date_end) VALUES (:business, :date_start, :customer, :date_end)", dbSlots)
	return err
}
