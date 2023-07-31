package utils

import "time"

func ConvertNanoSecondToTime(timestamp int64) time.Time {
	return time.Unix(timestamp/1e9, timestamp%1e9)
}

func SliceHasRepeatedElement[T comparable](slice []T) bool {
	exist := make(map[T]struct{})

	for _, e := range slice {
		_, ok := exist[e]
		if ok {
			return true
		}
		exist[e] = struct{}{}
	}

	return false
}
