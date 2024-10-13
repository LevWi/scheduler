package common

import (
	"time"
)

type ID = string

type Slot struct {
	Client ID
	Start  time.Time
	Len    int
}

type Interval struct {
	Start time.Time
	End   time.Time
}

type Appointment struct {
	Business ID
	Slots    []Slot
}
