package business

import (
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/holidays"
)

type BusinessMeta struct {
	Id          common.ID
	Login       string
	Description string
}

type SlotProducer interface {
	GetSlots() []common.Interval
}

type Business struct {
	Id           common.ID
	WorkingTime  common.IntervalsProducers
	BlockedTime  common.IntervalsProducer
	Appointments SlotProducer
	Interval     common.Interval
}

func PrepareBusiness(id common.ID, interval common.Interval) (Business, error) {
	out := Business{Id: id, Interval: interval}
	out.BlockedTime = holidays.KzHolidaysProducer()

	//TODO
	return out, nil
}
