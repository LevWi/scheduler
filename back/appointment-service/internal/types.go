package common

type ID = string

type Slot struct {
	Client ID
	Interval
}

type Appointment struct {
	Business ID
	Slots    []Slot
}
