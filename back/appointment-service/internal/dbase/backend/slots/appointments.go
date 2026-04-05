package slots

import (
	"encoding/json"
	"fmt"
	common "scheduler/appointment-service/internal"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type TimeSlotsStorage struct {
	*sqlx.DB
}

type BusinessSlotSettings struct {
	DefaultChunk time.Duration
	MaxChunk     time.Duration
}

func validateBusinessSlotSettings(settings BusinessSlotSettings) error {
	minChunk := common.MinBookingSlotChunk
	if settings.DefaultChunk < minChunk || settings.MaxChunk < minChunk {
		return fmt.Errorf("slot chunk is less than minimum allowed")
	}
	if settings.DefaultChunk > settings.MaxChunk {
		return fmt.Errorf("default chunk is greater than max chunk")
	}
	return nil
}

type dbBusinessSlotSettings struct {
	DefaultChunkMinutes int `db:"default_chunk_minutes"`
	MaxChunkMinutes     int `db:"max_chunk_minutes"`
}

func (db *TimeSlotsStorage) GetBusinessSlotSettings(businessID common.ID) (BusinessSlotSettings, error) {
	var row dbBusinessSlotSettings
	err := db.Get(&row, `SELECT default_chunk_minutes, max_chunk_minutes FROM business_slot_settings WHERE business_id = $1`, string(businessID))
	if err != nil {
		return BusinessSlotSettings{}, err
	}

	settings := BusinessSlotSettings{
		DefaultChunk: time.Duration(row.DefaultChunkMinutes) * time.Minute,
		MaxChunk:     time.Duration(row.MaxChunkMinutes) * time.Minute,
	}

	if err := validateBusinessSlotSettings(settings); err != nil {
		return BusinessSlotSettings{}, err
	}

	return settings, nil
}

func (db *TimeSlotsStorage) SetBusinessSlotSettings(businessID common.ID, settings BusinessSlotSettings) error {
	if err := validateBusinessSlotSettings(settings); err != nil {
		return err
	}

	_, err := db.Exec(`
		INSERT INTO business_slot_settings (business_id, default_chunk_minutes, max_chunk_minutes)
		VALUES ($1, $2, $3)
		ON CONFLICT (business_id) DO UPDATE
		SET default_chunk_minutes = EXCLUDED.default_chunk_minutes,
		    max_chunk_minutes = EXCLUDED.max_chunk_minutes`,
		string(businessID),
		int(settings.DefaultChunk.Minutes()),
		int(settings.MaxChunk.Minutes()),
	)
	return err
}

type dbBusySlot struct {
	Customer  string `db:"customer_id"`
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

type RuleID = string

type DbBusinessRule struct {
	Id   RuleID
	Rule common.IntervalRRuleWithType
}

type dbJsonBusinessRule struct {
	Id   RuleID
	Rule string
}

func ConvertSlice(in []dbJsonBusinessRule) ([]DbBusinessRule, error) {
	var intervalsRRules []DbBusinessRule
	for _, el := range in {
		var tmp DbBusinessRule
		err := json.Unmarshal([]byte(el.Rule), &tmp.Rule)
		if err != nil {
			return nil, err
		}
		tmp.Id = el.Id
		intervalsRRules = append(intervalsRRules, tmp)
	}
	return intervalsRRules, nil
}

func (db *TimeSlotsStorage) GetBusinessRules(business_id common.ID) ([]DbBusinessRule, error) {
	var rules []dbJsonBusinessRule
	err := db.Select(&rules, "SELECT id, rule FROM business_work_rule WHERE business_id = $1", string(business_id))
	if err != nil {
		return nil, err
	}

	return common.MapE(rules, func(in dbJsonBusinessRule) (DbBusinessRule, error) {
		var tmp DbBusinessRule
		err := json.Unmarshal([]byte(in.Rule), &tmp.Rule)
		if err != nil {
			return DbBusinessRule{}, err
		}
		tmp.Id = in.Id
		return tmp, nil
	})
}

func (db *TimeSlotsStorage) AddBusinessRule(businessID string, rule common.IntervalRRuleWithType) (RuleID, error) {
	b, err := json.Marshal(rule)
	if err != nil {
		return "", err
	}

	newID := uuid.New().String()

	_, err = db.Exec(`
		INSERT INTO business_work_rule (id, business_id, rule)
		VALUES ($1, $2, $3)
	`, newID, businessID, string(b))
	return newID, err
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
	var jsonRules []string
	err := db.Select(&jsonRules, "SELECT rule FROM business_work_rule WHERE business_id = $1", string(business_id))
	if err != nil {
		return nil, err
	}

	intervalsRRules, err := common.MapE(jsonRules, func(in string) (common.IntervalRRuleWithType, error) {
		var tmp common.IntervalRRuleWithType
		err := json.Unmarshal([]byte(in), &tmp)
		if err != nil {
			return common.IntervalRRuleWithType{}, err
		}
		return tmp, nil
	})
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

// GetCustomerAppointmentsInRange returns customer's appointments for the given business
// that overlap the requested interval.
//
// Acceptable arguments:
//   - businessID and customerID must identify the exact business/customer pair to filter by.
//   - between.Start is required and used as the lower bound.
//   - between.End is optional:
//   - if between.End.IsZero() then all appointments with date_end >= between.Start are returned
//     (including appointments that started before between.Start but are still in progress).
//   - if between.End is set then appointments are returned only when they overlap
//     [between.Start, between.End], i.e. date_end >= between.Start AND date_start <= between.End.
//
// Returned appointments are ordered by date_start ascending.
func (db *TimeSlotsStorage) GetCustomerAppointmentsInRange(businessID common.ID, customerID common.ID, between common.Interval) ([]common.BusySlot, error) {
	var (
		dbSlots []dbBusySlot
		err     error
	)

	if between.End.IsZero() {
		err = db.Select(
			&dbSlots,
			`SELECT * FROM appointments
			 WHERE business_id = $1 AND customer_id = $2 AND date_end >= $3
			 ORDER BY date_start`,
			string(businessID), string(customerID), between.Start.Unix(),
		)
	} else {
		err = db.Select(
			&dbSlots,
			`SELECT * FROM appointments
			 WHERE business_id = $1 AND customer_id = $2 AND date_end >= $3 AND date_start <= $4
			 ORDER BY date_start`,
			string(businessID), string(customerID), between.Start.Unix(), between.End.Unix(),
		)
	}
	if err != nil {
		return nil, err
	}

	slotsOut := make([]common.BusySlot, 0, len(dbSlots))
	for _, dbSlot := range dbSlots {
		slotsOut = append(slotsOut, dbSlot.ToSlot())
	}
	return slotsOut, nil
}

func (db *TimeSlotsStorage) DeleteSlots(business_id common.ID, customerID common.ID, start time.Time, end time.Time) error {
	_, err := db.Exec("DELETE FROM appointments WHERE business_id = $1 AND customer_id = $2 AND date_start BETWEEN $3 AND $4",
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
	_, err := db.NamedExec("INSERT INTO appointments (business_id, date_start, customer_id, date_end) VALUES (:business_id, :date_start, :customer_id, :date_end)", dbSlots)
	return err
}
