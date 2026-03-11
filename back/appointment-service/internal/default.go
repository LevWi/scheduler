package common

import "time"

const (
	DefaultBookingSlotChunk = 15 * time.Minute
	MaxBookingSlotChunk     = 1 * time.Hour
	MinBookingSlotChunk     = 5 * time.Minute
)
