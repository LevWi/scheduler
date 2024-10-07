package common

import (
	"time"
)

type ID string

type Slot struct {
	Client ID
	Start  time.Time
	Len    int
}

type Appointment struct {
	Business ID
	Slots    []Slot
}
