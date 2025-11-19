package command

import (
	"context"
	common "scheduler/appointment-service/internal"
	"time"
)

type SlotsProvider interface {
	AvailableSlotsInRange(ctx context.Context, interval common.Interval) ([]common.Slot, error)
}

type WeekSlots struct {
	storage SlotsProvider
}

func NewWeekSlots(storage SlotsProvider) *WeekSlots {
	return &WeekSlots{
		storage: storage,
	}
}

func (ws *WeekSlots) ThisWeek(ctx context.Context, now time.Time) ([]common.Slot, error) {
	interval := common.Interval{
		Start: now,
		End:   common.NextMonday(now),
	}

	return ws.storage.AvailableSlotsInRange(ctx, interval)
}

func (ws *WeekSlots) NextWeek(ctx context.Context, now time.Time) ([]common.Slot, error) {
	interval := common.Interval{
		Start: common.NextMonday(now),
	}
	interval.End = common.NextMonday(interval.Start)

	return ws.storage.AvailableSlotsInRange(ctx, interval)
}
