package atproto

import (
	"errors"
	"fmt"
	"time"
)

const (
	// TimeFormat is the standard time format specified by the ATProto.
	//
	// See [ParseTime]
	TimeFormat = "2006-01-02T15:04:05.000Z07:00"
)

// Errors returned while parsing a [time.Time].
var (
	// ErrCannotParseTime is returned by [ParseTime] if an error occurs.
	ErrCannotParseTime = errors.New("cannot parse time")
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
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w: %w", ErrCannotParseTime, err)
		}
	}()
	if err != nil {
		return
	}
	t = t.UTC()
	if t.Year() < 0 {
		err = errors.New("invalid time (following spec)")
	}
	return
}
