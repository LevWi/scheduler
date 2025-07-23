package common

import "time"

func TsSec(t time.Time) int64 {
	return t.UTC().Unix()
}
