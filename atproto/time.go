package atproto

import (
	"errors"
	"time"
)

const (
	// TimeFormat is the standard time format specified by the ATProto.
	//
	// See [ParseTime]
	TimeFormat = "2006-01-02T15:04:05.000Z07:00"
)

// ParseTime returns a [time.Time] if it follows the standard time format specified by the ATProto.
//
// See [TimeFormat].
// Fallback to [time.RFC3339] if it doesn't work.
func ParseTime(raw string) (t time.Time, err error) {
	t, err = time.Parse(TimeFormat, raw)
	if err != nil {
		t, err = time.Parse(time.RFC3339, raw)
	}
	if err != nil {
		return
	}
	t = t.UTC()
	if t.Year() < 0 {
		err = errors.New("invalid time (following spec)")
	}
	return
}
