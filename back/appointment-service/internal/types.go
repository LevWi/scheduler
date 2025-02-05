package common

type ID = string

type BusySlot struct {
	Client ID
	Interval
}

type Appointment struct {
	Business ID
	Slots    []BusySlot
}
