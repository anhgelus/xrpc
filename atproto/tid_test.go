package atproto

import (
	"testing"

	"pgregory.net/rapid"
)

func genTID(t *rapid.T, label string) TID {
	beg := rapid.StringOfN(rapid.RuneFrom([]rune("234567abcdefhij")), 1, -1, 1).Draw(t, label+" begin")
	end := rapid.StringOfN(rapid.RuneFrom([]rune("234567abcdefghijklmnopqrstuvxyz")), 12, -1, 12).Draw(t, label+" end")
	return TID(beg + end)
}

var (
	invalidTID = []string{
		"3jzfcijpj2z21",
		"0000000000000",
		"3JZFCIJPJ2Z2A",
		"3jzfcijpj2z2aa",
		"3jzfcijpj2z2",
		"222",
		"3jzf-cij-pj2z-2a",
		"zzzzzzzzzzzzz",
		"kjzfcijpj2z2a",
	}
)

func TestParseTID(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		tid := genTID(t, "tid")
		parsed, err := ParseTID(string(tid))
		if err != nil {
			t.Fatal(err)
		}
		if parsed != tid {
			t.Errorf("invalid TID: %s, wanted %s", parsed, tid)
		}
	})
	for _, tid := range invalidTID {
		_, err := ParseTID(tid)
		if err == nil {
			t.Errorf("expected error for %s", tid)
		}
	}
}

func TestNewTID(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		tm := genTime(t, "time")
		clock := rapid.Uint().Filter(func(u uint) bool { return u <= clockIdBits }).Draw(t, "clock")
		tid := NewTID(tm, clock)
		if tid.Time() != tm {
			t.Errorf("invalid time: %s, wanted %s", tid.Time(), tm)
		}
		if tid.ClockID() != clock {
			t.Errorf("invalid clockId: %d, wanted %d", tid.ClockID(), clock)
		}
	})
}
