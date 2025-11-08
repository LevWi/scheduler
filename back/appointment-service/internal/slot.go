package common

import "time"

type Slot struct {
	Start time.Time
	Dur   time.Duration
}
