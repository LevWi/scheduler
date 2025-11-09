package common

type ID = string

type BusySlot struct {
	Customer ID
	Interval
}

type Appointment struct {
	Business ID
	Slots    []BusySlot
}
