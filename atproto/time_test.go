package atproto

import (
	"testing"
	"time"

	"pgregory.net/rapid"
)

var (
	validTimes = []string{
		"1985-04-12T23:20:50.123Z",
		"1985-04-12T23:20:50.123456Z",
		"1985-04-12T23:20:50.120Z",
		"1985-04-12T23:20:50.120000Z",
		"0001-01-01T00:00:00.000Z",
		"0000-01-01T00:00:00.000Z",
	}
	supportedTimes = []string{
		"1985-04-12T23:20:50.12345678912345Z",
		"1985-04-12T23:20:50Z",
		"1985-04-12T23:20:50.0Z",
		"1985-04-12T23:20:50.123+00:00",
		"1985-04-12T23:20:50.123-07:00",
	}
	invalidTimes = []string{
		"1985-04-12",
		"1985-04-12T23:20Z",
		"1985-04-12T23:20:5Z",
		"1985-04-12T23:20:50.123",
		"+001985-04-12T23:20:50.123Z",
		"23:20:50.123Z",
		"-1985-04-12T23:20:50.123Z",
		"1985-4-12T23:20:50.123Z",
		"01985-04-12T23:20:50.123Z",
		"1985-04-12T23:20:50.123+00",
		"1985-04-12T23:20:50.123+0000",
		"1985-04-12t23:20:50.123Z",
		"1985-04-12T23:20:50.123z",
		//"1985-04-12T23:20:50.123-00:00", // we are supporting this because it is easier
		"1985-04-12 23:20:50.123Z",
		"1985-04-12T23:20:50.123",
		"1985-04-12T23:99:50.123Z",
		"1985-00-12T23:20:50.123Z",
		"0000-01-01T00:00:00+01:00",
	}
)

func genTime(t *rapid.T, label string) time.Time {
	return time.UnixMicro(int64(rapid.Uint32().Draw(t, label))).UTC()
}

func TestParseTime(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ti := genTime(t, "time")
		tt, err := ParseTime(ti.Format(TimeFormat))
		if err != nil {
			t.Fatal(err)
		}
		if tt.Format(TimeFormat) != ti.Format(TimeFormat) {
			t.Errorf("invalid time: %s, wanted %s", tt, ti)
		}
	})
	for _, raw := range validTimes {
		_, err := ParseTime(raw)
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, raw := range supportedTimes {
		_, err := ParseTime(raw)
		if err != nil {
			t.Errorf("unsupported time: %s, %v", raw, err)
		}
	}
	for _, raw := range invalidTimes {
		_, err := ParseTime(raw)
		if err == nil {
			t.Errorf("expected error for %s", raw)
		}
	}
}
