package util

import "time"

func GetCurrentTimeUnixNano() int64 {
	return time.Now().UnixNano()
}
