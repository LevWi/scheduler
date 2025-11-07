package command

import (
	"context"
	common "scheduler/appointment-service/internal"
	"time"
)

type SlotsProvider interface {
	AvailableSlotsInRange(ctx context.Context, business_id common.ID, interval common.Interval) (common.Intervals, error)
}

type CurrentWeek struct {
	businessID common.ID
	storage    SlotsProvider
}

func NewCurrentWeek(businessID common.ID, storage SlotsProvider) *CurrentWeek {
	return &CurrentWeek{
		businessID: businessID,
		storage:    storage,
	}
}

func (cw *CurrentWeek) Slots(ctx context.Context, now time.Time) (common.Intervals, error) {
	interval := common.Interval{
		Start: now,
		End:   common.NextMonday(now),
	}

	return cw.storage.AvailableSlotsInRange(ctx, cw.businessID, interval)
}

type NextWeek struct {
	businessID common.ID
	storage    SlotsProvider
}

func NewNextWeek(businessID common.ID, storage SlotsProvider) *NextWeek {
	return &NextWeek{
		businessID: businessID,
		storage:    storage,
	}
}

func (cw *NextWeek) Slots(ctx context.Context, now time.Time) (common.Intervals, error) {
	interval := common.Interval{
		Start: common.NextMonday(now),
	}
	interval.End = common.NextMonday(interval.Start)

	return cw.storage.AvailableSlotsInRange(ctx, cw.businessID, interval)
}
