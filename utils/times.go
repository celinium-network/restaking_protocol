package utils

import "time"

func ConvertNanoSecondToTime(timestamp int64) time.Time {
	return time.Unix(timestamp/1e9, timestamp%1e9)
}
