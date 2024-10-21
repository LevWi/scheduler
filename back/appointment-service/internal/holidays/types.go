package holidays

import common "scheduler/appointment-service/internal"

// Define a struct to represent a holiday with a description and interval (start and end date)
type Holiday struct {
	Description string
	Interval    common.Interval
}
