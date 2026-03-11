package common

import "time"

const (
	DefaultBookingSlotChunk    = 15 * time.Minute
	DefaultMaxBookingSlotChunk = 1 * time.Hour
	MinBookingSlotChunk        = 5 * time.Minute
)
