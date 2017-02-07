package bc

import "time"

// Millis converts a time.Time to a number of milliseconds since 1970.
func Millis(t time.Time) uint64 {
	return uint64(t.UnixNano()) / uint64(time.Millisecond)
}

// DurationMillis converts a time.Duration to a number of milliseconds.
func DurationMillis(d time.Duration) uint64 {
	return uint64(d / time.Millisecond)
}

// Time converts a number of milliseconds since 1970 to a time.Time.
func Time(millis uint64) time.Time {
	nano := millis * uint64(time.Millisecond)
	return time.Unix(0, int64(nano)).UTC()
}
