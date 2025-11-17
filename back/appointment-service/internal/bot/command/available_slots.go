package command

import (
	"context"
	common "scheduler/appointment-service/internal"
	"time"
)

type SlotsProvider interface {
	AvailableSlotsInRange(ctx context.Context, business_id common.ID, interval common.Interval) (common.Intervals, error)
}

type WeekSlots struct {
	businessID common.ID
	storage    SlotsProvider
}

func NewWeekSlots(businessID common.ID, storage SlotsProvider) *WeekSlots {
	return &WeekSlots{
		businessID: businessID,
		storage:    storage,
	}
}

func (ws *WeekSlots) ThisWeek(ctx context.Context, now time.Time) (common.Intervals, error) {
	interval := common.Interval{
		Start: now,
		End:   common.NextMonday(now),
	}

	return ws.storage.AvailableSlotsInRange(ctx, ws.businessID, interval)
}

func (ws *WeekSlots) NextWeek(ctx context.Context, now time.Time) (common.Intervals, error) {
	interval := common.Interval{
		Start: common.NextMonday(now),
	}
	interval.End = common.NextMonday(interval.Start)

	return ws.storage.AvailableSlotsInRange(ctx, ws.businessID, interval)
}
